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
	"sync"
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
	runs, err := application.NewRunService(store, application.SystemClock{}, randomIDs{})
	if err != nil {
		return err
	}
	// The daemon's shipped fake profile is a finite transcript: it emits its
	// deterministic startup records and exits without waiting for input.
	runner := runnerlocal.New(sandboxlocal.New(), harnessfake.Driver{Config: harnessfake.Config{EarlyExit: true}})
	defer runner.Close()
	coordinator := newRunCoordinator(store, runner, randomIDs{}, logger)
	stream := httpapi.NewEventStream(64)
	defer stream.Close()
	api, err := httpapi.NewHandler(httpapi.Config{Token: cfg.Token, CSRFToken: cfg.CSRFToken, AllowedOrigin: cfg.AllowedOrigin, Actor: cfg.Actor, Logger: logger, RequestID: requestID, EventStream: stream}, runBackend{runs: runs, runtime: coordinator, events: store})
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

type runBackend struct {
	runs    *application.RunService
	runtime application.RunRuntime
	events  application.EventStore
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
func (b runBackend) StartRun(ctx context.Context, _ string, id httpapi.RunId, _ string, version int64) (httpapi.Run, error) {
	runID, err := apiRunID(id)
	if err != nil {
		return httpapi.Run{}, httpapi.ErrRunNotFound
	}
	view, err := b.runs.Start(ctx, runID, uint64(version), b.runtime)
	switch {
	case errors.Is(err, application.ErrNotFound):
		return httpapi.Run{}, httpapi.ErrRunNotFound
	case errors.Is(err, application.ErrUnsupportedRunProfiles):
		return httpapi.Run{}, httpapi.ErrValidation
	case errors.Is(err, application.ErrConflict):
		return httpapi.Run{}, httpapi.ErrPreconditionFailed
	case err != nil:
		return httpapi.Run{}, err
	}
	return apiRun(view), nil
}
func (b runBackend) ListRunEvents(ctx context.Context, _ string, id httpapi.RunId, after int64, limit int) (httpapi.EventPage, error) {
	runID, err := apiRunID(id)
	if err != nil {
		return httpapi.EventPage{}, httpapi.ErrRunNotFound
	}
	values, err := b.runs.Events(ctx, runID, uint64(after), limit+1, b.events)
	if errors.Is(err, application.ErrNotFound) {
		return httpapi.EventPage{}, httpapi.ErrRunNotFound
	}
	if err != nil {
		return httpapi.EventPage{}, err
	}
	hasMore := len(values) > limit
	if hasMore {
		values = values[:limit]
	}
	out := httpapi.EventPage{Events: make([]httpapi.Event, 0, len(values)), HasMore: hasMore, NextAfterSequence: after}
	for _, value := range values {
		var payload map[string]interface{}
		_ = json.Unmarshal(value.Payload, &payload)
		sourceBytes, _ := json.Marshal(value.Source)
		var source map[string]interface{}
		_ = json.Unmarshal(sourceBytes, &source)
		extensions := map[string]interface{}{}
		for key, raw := range value.Extensions {
			var extension interface{}
			_ = json.Unmarshal(raw, &extension)
			extensions[key] = extension
		}
		out.Events = append(out.Events, httpapi.Event{EventId: value.EventID, RunId: httpapi.RunId(formatAPIRunID(runID)), Sequence: int64(value.Sequence), OccurredAt: value.OccurredAt, Kind: value.Kind, SchemaVersion: int(value.SchemaVersion), Source: source, Sensitivity: httpapi.EventSensitivity(value.Sensitivity), Payload: payload, Extensions: &extensions})
		out.NextAfterSequence = int64(value.Sequence)
	}
	return out, nil
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

func (runBackend) StopRun(context.Context, string, httpapi.RunId, string, int64, httpapi.StopRunRequest) (httpapi.Run, error) {
	return httpapi.Run{}, errStopDeferred
}
func (runBackend) SendRunInput(context.Context, string, httpapi.RunId, string, int64, httpapi.RunInputRequest) (httpapi.InputAccepted, error) {
	return httpapi.InputAccepted{}, errInputDeferred
}

type activeExecution struct {
	runID domain.RunID
	lease application.RunLease
}

type runCoordinator struct {
	store  *postgres.Store
	runner *runnerlocal.Runner
	super  *supervisor.Supervisor
	ids    randomIDs
	logger *slog.Logger
	mu     sync.Mutex
	active map[string]activeExecution
}

type persistRedactor struct{}

func (persistRedactor) Redact(_ context.Context, event application.UnsequencedEvent) (application.UnsequencedEvent, error) {
	return event, nil
}

func newRunCoordinator(store *postgres.Store, runner *runnerlocal.Runner, ids randomIDs, logger *slog.Logger) *runCoordinator {
	c := &runCoordinator{store: store, runner: runner, ids: ids, logger: logger, active: map[string]activeExecution{}}
	c.super = supervisor.New(store, runner, persistRedactor{}, application.SystemClock{}, "winchd", 30*time.Second)
	go c.consume()
	return c
}

func (c *runCoordinator) Start(ctx context.Context, record application.RunRecord) error {
	executionID, leaseToken := uuid.NewString(), uuid.NewString()
	lease, err := c.super.Acquire(ctx, record.ID, leaseToken)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.active[executionID] = activeExecution{runID: record.ID, lease: lease}
	c.mu.Unlock()
	prepare, _ := json.Marshal(protocol.PreparePayload{WorkspaceID: record.ID.String()})
	if err = c.super.Execute(ctx, lease, supervisor.Command{DesiredState: domain.RunStatePreparing, HarnessDriver: "fake", SandboxDriver: "local", ExecutionID: executionID, Message: protocol.RunnerMessage{Version: protocol.RunnerVersion{Major: protocol.RunnerProtocolMajor}, Kind: "prepare", CommandID: uuid.NewString(), Payload: prepare}}); err != nil {
		return err
	}
	if err = c.setState(ctx, record.ID, domain.RunStatePreparing); err != nil {
		return err
	}
	start, _ := json.Marshal(protocol.StartPayload{LaunchProfile: "fake"})
	if err = c.super.Execute(ctx, lease, supervisor.Command{DesiredState: domain.RunStateRunning, HarnessDriver: "fake", SandboxDriver: "local", ExecutionID: executionID, Message: protocol.RunnerMessage{Version: protocol.RunnerVersion{Major: protocol.RunnerProtocolMajor}, Kind: "start", CommandID: uuid.NewString(), Payload: start}}); err != nil {
		return err
	}
	return c.setState(ctx, record.ID, domain.RunStateRunning)
}

func (c *runCoordinator) consume() {
	for observation := range c.runner.Observations() {
		c.mu.Lock()
		active, ok := c.active[observation.ExecutionID]
		c.mu.Unlock()
		if !ok {
			continue
		}
		ctx := context.Background()
		if observation.Event != nil {
			event := *observation.Event
			if event.EventID.IsZero() {
				event.EventID = c.ids.NewEventID()
			}
			if _, err := c.super.Observe(ctx, active.lease, observation.Ordinal, []application.UnsequencedEvent{event}); err != nil {
				c.logger.Error("run observation failed", "component", "supervisor", "operation", "observe", "run_id", active.runID.String(), "error_code", "observation_failed")
			}
		}
		if observation.Exit != nil {
			state := domain.RunStateFailed
			if observation.Exit.Successful {
				state = domain.RunStateCompleted
			}
			_ = c.setState(ctx, active.runID, state)
			_ = c.runner.Cleanup(ctx, observation.ExecutionID)
			_ = c.super.Release(ctx, active.lease)
			c.mu.Lock()
			delete(c.active, observation.ExecutionID)
			c.mu.Unlock()
		}
	}
}

func (c *runCoordinator) setState(ctx context.Context, id domain.RunID, state domain.RunState) error {
	record, version, err := c.store.Get(ctx, id)
	if err != nil {
		return err
	}
	if record.Attempts[len(record.Attempts)-1].State.IsTerminal() {
		return nil
	}
	record.Attempts[len(record.Attempts)-1].State = state
	record.UpdatedAt = time.Now().UTC()
	_, err = c.store.Save(ctx, record, version)
	return err
}
