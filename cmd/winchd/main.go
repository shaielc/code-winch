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
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shaielc/code-winch/internal/adapters/postgres"
	"github.com/shaielc/code-winch/internal/adapters/transport/httpapi"
	"github.com/shaielc/code-winch/internal/platform/config"
	"github.com/shaielc/code-winch/internal/platform/telemetry"
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
	stream := httpapi.NewEventStream(64)
	defer stream.Close()
	api, err := httpapi.NewHandler(httpapi.Config{Token: cfg.Token, CSRFToken: cfg.CSRFToken, AllowedOrigin: cfg.AllowedOrigin, Actor: cfg.Actor, Logger: logger, RequestID: requestID, EventStream: stream}, unavailableBackend{})
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

// unavailableBackend keeps the run routes mounted but inert until P1-050 binds
// the real use cases. Reads answer not-found truthfully — no run can exist yet;
// creation reports an internal error rather than blaming the caller's body.
type unavailableBackend struct{}

var errRunBackendUnbound = errors.New("run use cases are not bound")

func (unavailableBackend) CreateRun(context.Context, string, string, httpapi.CreateRunRequest) (httpapi.Run, error) {
	return httpapi.Run{}, errRunBackendUnbound
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
