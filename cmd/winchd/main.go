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

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shaielc/code-winch/internal/adapters/postgres"
	"github.com/shaielc/code-winch/internal/adapters/transport/httpapi"
	"github.com/shaielc/code-winch/internal/platform/config"
	"github.com/shaielc/code-winch/internal/platform/telemetry"
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
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	logger := slog.New(telemetry.NewHandler(slog.NewJSONHandler(os.Stdout, nil)))
	_ = telemetry.NewRegistry()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("database pool: %w", err)
	}
	defer pool.Close()
	if err = pool.Ping(ctx); err != nil {
		return fmt.Errorf("database connection: %w", err)
	}
	if err = postgres.MigrateUp(ctx, pool); err != nil {
		return fmt.Errorf("database migration: %w", err)
	}
	logger.Info("migration applied", "component", "database", "operation", "migrate", "sequence", 5, "status", "ok")
	stream := httpapi.NewEventStream(64)
	defer stream.Close()
	api, err := httpapi.NewHandler(httpapi.Config{Token: cfg.Token, CSRFToken: cfg.CSRFToken, AllowedOrigin: cfg.AllowedOrigin, Actor: cfg.Actor, Logger: logger, RequestID: requestID, EventStream: stream}, unavailableBackend{})
	if err != nil {
		return err
	}
	assets, err := staticHandler(cfg.StaticDir, cfg.CSRFToken)
	if err != nil {
		return err
	}
	root := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/") {
			api.ServeHTTP(w, r)
			return
		}
		assets.ServeHTTP(w, r)
	})
	server := &http.Server{Addr: cfg.Addr, Handler: root, ReadHeaderTimeout: cfg.ShutdownTimeout}
	failure := make(chan error, 1)
	go func() {
		logger.Info("listener started", "component", "http", "operation", "listen", "status", "ready")
		failure <- server.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
		stream.Close()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err = server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()
			return fmt.Errorf("http shutdown: %w", err)
		}
		return nil
	case err = <-failure:
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

func staticHandler(dir, csrf string) (http.Handler, error) {
	index, err := os.ReadFile(filepath.Join(dir, "index.html"))
	if err != nil {
		return nil, fmt.Errorf("static assets: %w", err)
	}
	injected := []byte(strings.ReplaceAll(string(index), "__WINCH_CSRF_TOKEN__", csrf))
	files := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	}), nil
}

type unavailableBackend struct{}

func (unavailableBackend) CreateRun(context.Context, string, string, httpapi.CreateRunRequest) (httpapi.Run, error) {
	return httpapi.Run{}, httpapi.ErrValidation
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
