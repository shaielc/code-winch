// Command winchd is the Code Winch daemon composition root.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	harnessfake "github.com/shaielc/code-winch/internal/adapters/harness/fake"
	"github.com/shaielc/code-winch/internal/adapters/postgres"
	sandboxlocal "github.com/shaielc/code-winch/internal/adapters/sandbox/local"
	"github.com/shaielc/code-winch/internal/adapters/transport/httpapi"
	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
	"github.com/shaielc/code-winch/internal/platform/config"
	"github.com/shaielc/code-winch/internal/platform/telemetry"
	runnerlocal "github.com/shaielc/code-winch/internal/runner/local"
	"github.com/shaielc/code-winch/internal/supervisor"
	"github.com/shaielc/code-winch/pkg/protocol"
)

// Bounds how long a client may dribble request headers; unrelated to the
// shutdown drain budget.
const readHeaderTimeout = 10 * time.Second

// The one harness and sandbox pair this build constructs. Phase 0 ships exactly
// this combination, so registration stays explicit in the composition root
// (docs/code-structure.md §4) and no name-to-driver registry exists to resolve.
const (
	harnessProfile = harnessfake.AdapterID
	sandboxProfile = "local"
)

// runLeaseDuration bounds how long a crashed daemon's ownership blocks a fresh
// lease. It does not bound a run: every durable write is fenced by the lease
// owner, token, and epoch rather than by its expiry, so a run outliving its
// lease keeps writing until some other owner actually takes over.
const runLeaseDuration = time.Minute

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil {
		var validation *config.ValidationError
		if errors.As(err, &validation) {
			// Validation errors contain field names only and are safe before the logger exists.
			fmt.Fprintln(os.Stderr, validation.Error())
		}
		slog.Error("daemon stopped", "component", "daemon", "error_code", "startup_failed")
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	started := time.Now()
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	logger := slog.New(telemetry.NewHandler(slog.NewJSONHandler(os.Stdout, nil)))
	// Anything reaching for the package-level logger must redact too.
	slog.SetDefault(logger)
	metrics := telemetry.NewRegistry()
	metrics.Declare("winch_startup_time_seconds", "status", "ready")
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("database pool: %w", err)
	}
	defer pool.Close()
	if err = pool.Ping(ctx); err != nil {
		return fmt.Errorf("database connection: %w", err)
	}
	migration, err := postgres.MigrateUp(ctx, pool)
	if err != nil {
		return fmt.Errorf("database migration: %w", err)
	}
	// A start that applied nothing must not claim it migrated the database.
	migrationStatus := "current"
	if migration.Applied > 0 {
		migrationStatus = "applied"
	}
	logger.Info("schema checked", "component", "database", "operation", "migrate", "sequence", migration.Version, "status", migrationStatus)
	store := postgres.New(pool)
	ids := randomIDs{}
	clock := application.SystemClock{}
	runs, err := application.NewRunService(store, clock, ids)
	if err != nil {
		return err
	}
	// The one supported pair, constructed directly: the fake harness under the
	// local sandbox, pumped by the in-process runner and serialized per run by
	// the supervisor, which fences every durable write behind a run lease.
	runner := runnerlocal.New(sandboxlocal.New(), harnessfake.Driver{Config: fakeHarnessConfig(cfg.FakeHarness)})
	control := supervisorControl{supervisor.New(store, runner, application.SensitivityRedactor{}, clock, "winchd-"+uuid.NewString(), runLeaseDuration)}
	starts, err := application.NewStartService(store, control, clock, ids, application.SupportedProfiles{Harness: harnessProfile, Sandbox: sandboxProfile})
	if err != nil {
		return err
	}
	events, err := application.NewEventService(store, store)
	if err != nil {
		return err
	}
	// One consumer drains the runner, so observations reach durable storage in
	// the order the runner produced them.
	observations := make(chan struct{})
	go func() {
		defer close(observations)
		for observation := range runner.Observations() {
			if obsErr := starts.Observe(context.WithoutCancel(ctx), application.RunnerObservation{ExecutionID: observation.ExecutionID, Ordinal: observation.Ordinal, Type: observation.Type, Event: observation.Event, Exit: observation.Exit}); obsErr != nil {
				logger.Error("run observation not recorded", "component", "supervisor", "operation", "observe", "error_code", "observation_failed")
			}
		}
	}()
	// Executions are released before the runner is, so no harness process
	// outlives the daemon that owns it. Registered here rather than after the
	// listener, so a start that fails later still takes its executions with it.
	defer func() {
		starts.Shutdown(context.WithoutCancel(ctx))
		runner.Close()
		<-observations
	}()
	stream := httpapi.NewEventStream(64)
	defer stream.Close()
	api, err := httpapi.NewHandler(httpapi.Config{Token: cfg.Token, CSRFToken: cfg.CSRFToken, AllowedOrigin: cfg.AllowedOrigin, Actor: cfg.Actor, Logger: logger, RequestID: requestID, EventStream: stream}, runBackend{runs: runs, starts: starts, events: events})
	if err != nil {
		return err
	}
	assets, served := staticHandler(cfg.StaticDir, cfg.CSRFToken)
	if !served {
		logger.Warn("web assets unavailable", "component", "http", "operation", "serve", "status", "degraded")
	}
	root := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/") {
			api.ServeHTTP(w, r)
			return
		}
		assets.ServeHTTP(w, r)
	})
	server := &http.Server{Addr: cfg.Addr, Handler: root, ReadHeaderTimeout: readHeaderTimeout}
	// Refuses to start on an undeclared metric or label, so emitters cannot
	// widen telemetry cardinality unnoticed.
	if err = metrics.Validate("winch_startup_time_seconds", map[string]string{"status": "ready"}); err != nil {
		return fmt.Errorf("startup metric: %w", err)
	}
	logger.Info("startup complete", "component", "daemon", "operation", "start", "status", "ready", "duration_ms", time.Since(started).Milliseconds())
	logger.Info("listener started", "component", "http", "operation", "listen", "status", "ready")
	return serve(ctx, server, stream, cfg.ShutdownTimeout)
}

// serve runs until ctx is cancelled or the listener fails, then disconnects
// live subscribers and drains in-flight requests within timeout.
func serve(ctx context.Context, server *http.Server, stream *httpapi.EventStream, timeout time.Duration) error {
	failure := make(chan error, 1)
	go func() { failure <- server.ListenAndServe() }()
	select {
	case <-ctx.Done():
		stream.Close()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()
			return fmt.Errorf("http shutdown: %w", err)
		}
		return nil
	case err := <-failure:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("http serve: %w", err)
	}
}

func requestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "request-unavailable"
	}
	return hex.EncodeToString(b[:])
}

// The API boots without a web build: a fresh clone has no web/dist, and only
// the container image runs `npm run build`. Missing assets serve 404 rather
// than holding back the listener.
func staticHandler(dir, csrf string) (http.Handler, bool) {
	index, err := os.ReadFile(filepath.Join(dir, "index.html"))
	if err != nil {
		return http.NotFoundHandler(), false
	}
	injected := []byte(strings.ReplaceAll(string(index), "__WINCH_CSRF_TOKEN__", csrf))
	files := http.FileServer(http.Dir(dir))
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(dir, filepath.Clean(r.URL.Path))
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(injected)
			return
		}
		if info, e := os.Stat(path); e == nil && !info.IsDir() {
			files.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(injected)
	})
	return handler, true
}

type randomIDs struct{}

func (randomIDs) NewWorkspaceID() domain.WorkspaceID {
	id, _ := domain.ParseWorkspaceID(uuid.NewString())
	return id
}
func (randomIDs) NewRunID() domain.RunID { id, _ := domain.ParseRunID(uuid.NewString()); return id }
func (randomIDs) NewAttemptID() domain.AttemptID {
	id, _ := domain.ParseAttemptID(uuid.NewString())
	return id
}
func (randomIDs) NewEventID() domain.EventID {
	id, _ := domain.ParseEventID(uuid.NewString())
	return id
}
func (randomIDs) NewCommandID() domain.CommandID {
	id, _ := domain.ParseCommandID(uuid.NewString())
	return id
}
func (randomIDs) NewArtifactID() domain.ArtifactID {
	id, _ := domain.ParseArtifactID(uuid.NewString())
	return id
}
func (randomIDs) NewCredentialID() domain.CredentialID {
	id, _ := domain.ParseCredentialID(uuid.NewString())
	return id
}
func (randomIDs) NewWorkflowID() domain.WorkflowID {
	id, _ := domain.ParseWorkflowID(uuid.NewString())
	return id
}

// fakeHarnessConfig resolves the shipped fake profile. EarlyExit is always set:
// it only takes effect once a transcript has played out without ending the
// harness, and a run started through the API has no way to answer an
// interactive prompt, so a harness left reading its terminal would never reach
// a terminal state.
func fakeHarnessConfig(cfg config.FakeHarnessConfig) harnessfake.Config {
	return harnessfake.Config{Executable: cfg.Binary, Transcript: cfg.Transcript, Delay: cfg.Delay, ForceFailure: cfg.ForceFailure, MalformedLine: cfg.MalformedLine, EarlyExit: true}
}

// supervisorControl adapts the supervisor to the application's executor port.
// The application layer cannot name supervisor.Command: the supervisor package
// depends on the application's ports, so the translation belongs out here.
type supervisorControl struct{ *supervisor.Supervisor }

func (c supervisorControl) Execute(ctx context.Context, lease application.RunLease, command application.ExecutionCommand) error {
	return c.Supervisor.Execute(ctx, lease, supervisor.Command{DesiredState: command.DesiredState, HarnessDriver: command.HarnessDriver, SandboxDriver: command.SandboxDriver, ExecutionID: command.ExecutionID, Message: command.Message})
}

type runBackend struct {
	runs   *application.RunService
	starts *application.StartService
	events *application.EventService
}

func (b runBackend) CreateRun(ctx context.Context, actor string, key string, request httpapi.CreateRunRequest) (httpapi.Run, error) {
	view, err := b.runs.Create(ctx, application.CreateRunCommand{WorkspacePath: request.WorkspacePath, HarnessProfile: request.HarnessProfile, SandboxProfile: request.SandboxProfile, Actor: actor, IdempotencyKey: key})
	if err != nil {
		if errors.Is(err, application.ErrIdempotencyConflict) {
			return httpapi.Run{}, httpapi.ErrIdempotencyConflict
		}
		return httpapi.Run{}, err
	}
	return apiRun(view), nil
}
func (b runBackend) GetRun(ctx context.Context, _ string, id httpapi.RunId) (httpapi.Run, error) {
	runID, err := apiRunID(id)
	if err != nil {
		return httpapi.Run{}, httpapi.ErrRunNotFound
	}
	view, err := b.runs.Get(ctx, runID)
	if errors.Is(err, application.ErrNotFound) {
		return httpapi.Run{}, httpapi.ErrRunNotFound
	}
	if err != nil {
		return httpapi.Run{}, err
	}
	return apiRun(view), nil
}
func apiRun(view application.RunView) httpapi.Run {
	r := view.Record
	sequence := int64(r.LastSequence)
	return httpapi.Run{Id: formatAPIRunID(r.ID), State: httpapi.RunState(r.Attempts[len(r.Attempts)-1].State), Version: int64(view.Version), CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt, WorkspacePath: r.WorkspacePath, HarnessProfile: r.HarnessProfile, SandboxProfile: r.SandboxProfile, LastSequence: &sequence}
}

const crockford = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

func formatAPIRunID(id domain.RunID) string {
	value := new(big.Int)
	value.SetString(strings.ReplaceAll(id.String(), "-", ""), 16)
	out := make([]byte, 26)
	base, remainder := big.NewInt(32), new(big.Int)
	for i := len(out) - 1; i >= 0; i-- {
		value.QuoRem(value, base, remainder)
		out[i] = crockford[remainder.Int64()]
	}
	return string(out)
}
func apiRunID(value string) (domain.RunID, error) {
	if len(value) != 26 {
		return domain.RunID{}, errors.New("invalid API run ID")
	}
	n := new(big.Int)
	base := big.NewInt(32)
	for _, char := range value {
		digit := strings.IndexRune(crockford, char)
		if digit < 0 {
			return domain.RunID{}, errors.New("invalid API run ID")
		}
		n.Mul(n, base)
		n.Add(n, big.NewInt(int64(digit)))
	}
	if n.BitLen() > 128 {
		return domain.RunID{}, errors.New("invalid API run ID")
	}
	hexValue := fmt.Sprintf("%032x", n)
	canonical := hexValue[:8] + "-" + hexValue[8:12] + "-" + hexValue[12:16] + "-" + hexValue[16:20] + "-" + hexValue[20:]
	return domain.ParseRunID(canonical)
}

var (
	errInputDeferred = errors.New("run input is not implemented; owner=P0-009")
	errStopDeferred  = errors.New("run stop is not implemented; owner=P0-011")
)

// StartRun launches the run and answers as soon as the harness is running. The
// idempotency key is echoed, not stored: a replayed start against a run that
// has already left `created` is refused as a state conflict rather than
// launching a second execution for the same run.
func (b runBackend) StartRun(ctx context.Context, _ string, id httpapi.RunId, _ string, version int64) (httpapi.Run, error) {
	runID, err := apiRunID(id)
	if err != nil {
		return httpapi.Run{}, httpapi.ErrRunNotFound
	}
	view, err := b.starts.Start(ctx, runID, uint64(version))
	if err != nil {
		return httpapi.Run{}, startProblem(err)
	}
	return apiRun(view), nil
}

func startProblem(err error) error {
	switch {
	case errors.Is(err, application.ErrNotFound):
		return httpapi.ErrRunNotFound
	case errors.Is(err, application.ErrUnsupportedProfile):
		return httpapi.ErrUnsupportedProfile
	case errors.Is(err, application.ErrPreconditionFailed):
		return httpapi.ErrPreconditionFailed
	// A lost lease means another writer owns the run, which is the same answer
	// to the caller as a command its state does not allow: read it and retry.
	case errors.Is(err, application.ErrStateConflict), errors.Is(err, supervisor.ErrStaleLease):
		return httpapi.ErrStateConflict
	}
	return err
}

func (b runBackend) ListRunEvents(ctx context.Context, _ string, id httpapi.RunId, after int64, limit int) (httpapi.EventPage, error) {
	runID, err := apiRunID(id)
	if err != nil {
		return httpapi.EventPage{}, httpapi.ErrRunNotFound
	}
	page, err := b.events.List(ctx, runID, uint64(after), limit)
	if errors.Is(err, application.ErrNotFound) {
		return httpapi.EventPage{}, httpapi.ErrRunNotFound
	}
	if err != nil {
		return httpapi.EventPage{}, err
	}
	out := httpapi.EventPage{Events: make([]httpapi.Event, 0, len(page.Events)), NextAfterSequence: int64(page.NextAfterSequence), HasMore: page.HasMore}
	for _, event := range page.Events {
		converted, convertErr := apiEvent(id, event)
		if convertErr != nil {
			return httpapi.EventPage{}, convertErr
		}
		out.Events = append(out.Events, converted)
	}
	return out, nil
}

// apiEvent re-encodes a stored event for the API. The stored run ID is a UUID
// and the API's is its Crockford form, so the requested ID is carried through
// rather than reformatted from the payload.
func apiEvent(id httpapi.RunId, event protocol.Event) (httpapi.Event, error) {
	payload, err := jsonObject(event.Payload)
	if err != nil {
		return httpapi.Event{}, fmt.Errorf("event payload: event=%s: %w", event.EventID, err)
	}
	source, err := marshalObject(event.Source)
	if err != nil {
		return httpapi.Event{}, fmt.Errorf("event source: event=%s: %w", event.EventID, err)
	}
	converted := httpapi.Event{EventId: event.EventID, RunId: id, Sequence: int64(event.Sequence), OccurredAt: event.OccurredAt, Kind: event.Kind, SchemaVersion: int(event.SchemaVersion), Source: source, Sensitivity: httpapi.EventSensitivity(event.Sensitivity), Payload: payload}
	if len(event.Extensions) > 0 {
		extensions, extErr := marshalObject(event.Extensions)
		if extErr != nil {
			return httpapi.Event{}, fmt.Errorf("event extensions: event=%s: %w", event.EventID, extErr)
		}
		converted.Extensions = &extensions
	}
	return converted, nil
}

func jsonObject(raw []byte) (map[string]any, error) {
	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func marshalObject(value any) (map[string]any, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return jsonObject(encoded)
}

func (runBackend) StopRun(context.Context, string, httpapi.RunId, string, int64, httpapi.StopRunRequest) (httpapi.Run, error) {
	return httpapi.Run{}, errStopDeferred
}
func (runBackend) SendRunInput(context.Context, string, httpapi.RunId, string, int64, httpapi.RunInputRequest) (httpapi.InputAccepted, error) {
	return httpapi.InputAccepted{}, errInputDeferred
}
