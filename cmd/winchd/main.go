// Command winchd is the Code Winch daemon composition root.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/shaielc/code-winch/internal/adapters/harness/fake"
	"github.com/shaielc/code-winch/internal/adapters/postgres"
	_ "github.com/shaielc/code-winch/internal/adapters/sandbox/fake"
	_ "github.com/shaielc/code-winch/internal/adapters/sandbox/local"
	"github.com/shaielc/code-winch/internal/adapters/transport/httpapi"
	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
	"github.com/shaielc/code-winch/internal/execution"
	"github.com/shaielc/code-winch/internal/platform/config"
	"github.com/shaielc/code-winch/internal/platform/telemetry"
	"github.com/shaielc/code-winch/internal/runner/local"
	"github.com/shaielc/code-winch/internal/supervisor"
	"github.com/shaielc/code-winch/pkg/protocol"
)

// Bounds how long a client may dribble request headers; unrelated to the
// shutdown drain budget.
const readHeaderTimeout = 10 * time.Second

const (
	// leaseDuration bounds how long this daemon's claim on a run survives
	// without renewal. A single local daemon never contends for one, but the
	// fencing that makes takeover safe is on from the start.
	leaseDuration = time.Minute
	// stopGrace is how long a harness has to exit after a stop before the
	// sandbox escalates to killing its process group.
	stopGrace = 5 * time.Second
)

// The metrics docs/architecture.md §7 requires of Phase 1. The registry is a
// guard rather than a client: it bounds names and label values, and an
// observation is emitted as one structured log record, the same way startup
// time has always been reported.
const (
	metricStartupTime = "winch_startup_time_seconds"
	metricQueueTime   = "winch_queue_time_seconds"
	metricActiveRuns  = "winch_active_runs"
)

// metricStatuses is every `status` value an emitter in this daemon passes.
// declareMetrics bounds them and startup validates the whole table, so an
// emitter reaching for a label nobody declared refuses to boot rather than
// widening telemetry cardinality unnoticed.
var metricStatuses = map[string][]string{
	metricStartupTime: {"ready"},
	// A queued run leaves for exactly one of these: the lease that begins
	// preparation, or the cancellation reconciliation records for a run whose
	// daemon died before it reached a harness.
	metricQueueTime: {string(domain.RunStatePreparing), string(domain.RunStateCancelled)},
	metricActiveRuns: {
		string(domain.RunStateRunning), string(domain.RunStateStopping),
	},
}

func declareMetrics(registry *telemetry.Registry) {
	registry.Declare(metricStartupTime, "status", "ready")
	registry.Declare(metricQueueTime, "status", string(domain.RunStatePreparing), string(domain.RunStateCancelled))
	registry.Declare(metricActiveRuns, "status", string(domain.RunStateRunning), string(domain.RunStateStopping))
}

// validateMetrics refuses a daemon whose declarations do not cover everything
// its emitters will emit. It runs before the listener accepts a request, so a
// drift between the two is a failed start rather than a silently dropped
// measurement in production.
func validateMetrics(registry *telemetry.Registry) error {
	for _, name := range []string{metricStartupTime, metricQueueTime, metricActiveRuns} {
		for _, status := range metricStatuses[name] {
			if err := registry.Validate(name, map[string]string{"status": status}); err != nil {
				return fmt.Errorf("metric %s status %s: %w", name, status, err)
			}
		}
	}
	return nil
}

// Outbox delivery. The interval bounds how long a committed event waits before
// live subscribers are told about it; durable history is already readable.
const (
	outboxInterval    = 50 * time.Millisecond
	outboxBatch       = 64
	outboxLease       = 30 * time.Second
	outboxBaseBackoff = 100 * time.Millisecond
	outboxMaxBackoff  = 5 * time.Second
	outboxMaxAttempts = 8
)

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
	declareMetrics(metrics)
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
	stream := httpapi.NewEventStream(64)
	defer stream.Close()
	store := postgres.New(pool)
	ids := application.RandomIDs{}
	clock := application.SystemClock{}
	instance := ids.NewCommandID().String()

	// The runtime, from the runner outwards. Drivers arrive through the
	// registry the blank imports above populate, so this root names no adapter.
	runners := local.NewPool(application.DefaultDrivers)
	runSupervisor := supervisor.New(store, runners, execution.ClassificationRedactor{}, clock, instance, leaseDuration).WithReconciliationIDs(ids)
	runs := application.NewRunService(store, store, ids, clock, application.DefaultDrivers, nil).
		WithMetrics(runMetrics{registry: metrics, logger: logger})
	engine, err := execution.New(execution.Config{
		Runs: store, States: runs, Control: store, Supervisor: runSupervisor,
		Runner: runners, IDs: ids, Clock: clock, Logger: logger, StopGrace: stopGrace,
	})
	if err != nil {
		return err
	}
	runs.WithExecutor(engine)
	inputs := application.NewInputService(store, execution.Capabilities{Runs: store}, nil)
	outbox, err := application.NewOutboxWorker(store, execution.Publisher{
		Events: liveEvents{stream}, Commands: store, Input: engine,
	}, wallClock{}, application.OutboxWorkerConfig{
		WorkerID: instance, LeaseToken: ids.NewCommandID().String(), BatchSize: outboxBatch,
		LeaseDuration: outboxLease, BaseBackoff: outboxBaseBackoff, MaxBackoff: outboxMaxBackoff, MaxAttempts: outboxMaxAttempts,
	})
	if err != nil {
		return err
	}

	// Reconcile before the listener accepts a request. A run this daemon did
	// not launch cannot be running locally, so answering "running" for it —
	// and accepting input that no execution will ever receive — is a lie the
	// first client would otherwise read.
	if _, err = execution.NewReconciler(store, runs, runSupervisor, runners, ids, logger).Sweep(ctx); err != nil {
		return fmt.Errorf("startup reconciliation: %w", err)
	}

	backend := httpapi.NewBackend(runs, inputs, ids)
	api, err := httpapi.NewHandler(httpapi.Config{Token: cfg.Token, CSRFToken: cfg.CSRFToken, AllowedOrigin: cfg.AllowedOrigin, Actor: cfg.Actor, Logger: logger, RequestID: requestID, EventStream: stream}, backend)
	if err != nil {
		return err
	}

	// Durable work outlives the request context: an execution that ends while
	// the daemon is draining must still record how it ended.
	runtimeCtx := context.WithoutCancel(ctx)
	var runtime sync.WaitGroup
	runtime.Add(2)
	go func() { defer runtime.Done(); engine.Consume(runtimeCtx, runners.Observations()) }()
	go func() { defer runtime.Done(); deliver(ctx, outbox, logger) }()
	defer func() {
		// Order matters: executions are ended before the runner closes, because
		// its pumps finish only once their streams are released.
		engine.Shutdown(runtimeCtx)
		runners.Close()
		runtime.Wait()
	}()
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
	if err = validateMetrics(metrics); err != nil {
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

// runMetrics turns the run service's observations into the bounded records the
// registry allows. It validates every label before emitting rather than
// trusting startup alone, so a value that slipped past the declaration is
// dropped instead of published.
type runMetrics struct {
	registry *telemetry.Registry
	logger   *slog.Logger
}

func (m runMetrics) QueueTime(_ context.Context, id domain.RunID, state domain.RunState, waited time.Duration) {
	if !m.allowed(metricQueueTime, string(state)) {
		return
	}
	m.logger.Info("run left the queue", "component", "runs", "operation", "queue",
		"run_id", id.String(), "status", string(state), "duration_ms", waited.Milliseconds())
}

// ActiveRuns emits one record per declared status rather than ranging over the
// map, so a gauge reads the same on every emission and an undeclared state in
// the map cannot reach the log.
func (m runMetrics) ActiveRuns(_ context.Context, counts map[domain.RunState]int) {
	for _, status := range metricStatuses[metricActiveRuns] {
		if !m.allowed(metricActiveRuns, status) {
			continue
		}
		m.logger.Info("active runs", "component", "runs", "operation", "observe",
			"status", status, "size", counts[domain.RunState(status)])
	}
}

func (m runMetrics) allowed(name, status string) bool {
	if err := m.registry.Validate(name, map[string]string{"status": status}); err != nil {
		// The registry's error names the metric it refused, so only the class
		// is logged; the measurement itself is dropped.
		m.logger.Warn("metric not emitted", "component", "daemon", "operation", "observe", "error_code", "undeclared_metric")
		return false
	}
	return true
}

var _ application.RunMetrics = runMetrics{}

// wallClock is the outbox worker's clock, which measures lease and backoff
// deadlines in wall time rather than in domain timestamps.
type wallClock struct{}

func (wallClock) Now() time.Time { return time.Now().UTC() }

// liveEvents adapts the durable event stream to the WebSocket broadcaster. It
// is best effort: a subscriber that misses a notification repairs from storage.
type liveEvents struct{ stream *httpapi.EventStream }

func (l liveEvents) Notify(event protocol.Event) {
	l.stream.Publish(event.RunID, httpapi.APIEvent(event))
}

// deliver drains the outbox until the daemon stops. A full batch is followed
// immediately by another claim so a burst of output is not paced by the tick.
func deliver(ctx context.Context, worker *application.OutboxWorker, logger *slog.Logger) {
	ticker := time.NewTicker(outboxInterval)
	defer ticker.Stop()
	for {
		for {
			delivered, err := worker.RunOnce(ctx)
			if err != nil {
				if ctx.Err() == nil {
					// The cause can name a publisher's payload, so only the
					// stable class is logged.
					logger.Warn("outbox delivery failed", "component", "outbox", "operation", "deliver", "error_code", "delivery_failed")
				}
				break
			}
			if delivered < outboxBatch {
				break
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
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

// unavailableBackend remains a small test fixture for exercising middleware
// without a database; production wiring above always uses ApplicationBackend.
type unavailableBackend struct{}

func (unavailableBackend) CreateRun(context.Context, string, string, httpapi.CreateRunRequest) (httpapi.Run, error) {
	return httpapi.Run{}, errors.New("unavailable")
}
func (unavailableBackend) GetRun(context.Context, string, httpapi.RunId) (httpapi.Run, error) {
	return httpapi.Run{}, httpapi.ErrRunNotFound
}
func (unavailableBackend) StartRun(context.Context, string, httpapi.RunId, string, int64) (httpapi.Run, error) {
	return httpapi.Run{}, httpapi.ErrRunNotFound
}
func (unavailableBackend) StopRun(context.Context, string, httpapi.RunId, string, int64, httpapi.StopRunRequest) (httpapi.Run, error) {
	return httpapi.Run{}, httpapi.ErrRunNotFound
}
func (unavailableBackend) ListRunEvents(context.Context, string, httpapi.RunId, int64, int) (httpapi.EventPage, error) {
	return httpapi.EventPage{}, httpapi.ErrRunNotFound
}
func (unavailableBackend) SendRunInput(context.Context, string, httpapi.RunId, string, int64, httpapi.RunInputRequest) (httpapi.InputAccepted, error) {
	return httpapi.InputAccepted{}, httpapi.ErrRunNotFound
}
