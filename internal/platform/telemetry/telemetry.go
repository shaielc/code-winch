// Package telemetry provides content-safe structured logging and bounded metrics.
package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

var AllowedAttributes = map[string]struct{}{"resource_id": {}, "run_id": {}, "workflow_id": {}, "error_code": {}, "component": {}, "operation": {}, "status": {}, "duration_ms": {}, "retry_count": {}, "size": {}, "sequence": {}}

type Handler struct{ next slog.Handler }

func NewHandler(next slog.Handler) *Handler                     { return &Handler{next: next} }
func (h *Handler) Enabled(c context.Context, l slog.Level) bool { return h.next.Enabled(c, l) }
func (h *Handler) Handle(c context.Context, r slog.Record) error {
	nr := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
	r.Attrs(func(a slog.Attr) bool {
		if _, ok := AllowedAttributes[a.Key]; ok {
			nr.AddAttrs(a)
		}
		return true
	})
	return h.next.Handle(c, nr)
}
func (h *Handler) WithAttrs(as []slog.Attr) slog.Handler {
	safe := as[:0]
	for _, a := range as {
		if _, ok := AllowedAttributes[a.Key]; ok {
			safe = append(safe, a)
		}
	}
	return &Handler{next: h.next.WithAttrs(safe)}
}
func (h *Handler) WithGroup(n string) slog.Handler { return &Handler{next: h.next.WithGroup(n)} }

// Redact returns a safe attribute only for explicitly bounded fields; unsafe fields are dropped.
func Redact(key string, value any) (slog.Attr, bool) {
	if _, ok := AllowedAttributes[key]; !ok {
		return slog.Attr{}, false
	}
	return slog.Any(key, value), true
}

type Metric struct {
	Name   string
	Labels map[string]map[string]struct{}
}
type Registry struct {
	mu      sync.RWMutex
	metrics map[string]Metric
}

func NewRegistry() *Registry {
	r := &Registry{metrics: map[string]Metric{}}
	for _, n := range []string{"winch_queue_time_seconds", "winch_startup_time_seconds", "winch_active_runs", "winch_dropped_live_subscribers_total", "winch_parser_failures_total", "winch_workflow_retries_total", "winch_cleanup_failures_total"} {
		r.metrics[n] = Metric{Name: n, Labels: map[string]map[string]struct{}{}}
	}
	return r
}
func (r *Registry) Declare(name, label string, values ...string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	m := r.metrics[name]
	if m.Labels == nil {
		m.Labels = map[string]map[string]struct{}{}
	}
	m.Labels[label] = map[string]struct{}{}
	for _, v := range values {
		m.Labels[label][v] = struct{}{}
	}
	r.metrics[name] = m
}
func (r *Registry) Validate(name string, labels map[string]string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.metrics[name]
	if !ok {
		return fmt.Errorf("undeclared metric")
	}
	for k, v := range labels {
		set, ok := m.Labels[k]
		if !ok {
			return fmt.Errorf("undeclared label")
		}
		if _, ok = set[v]; !ok {
			return fmt.Errorf("undeclared label value")
		}
	}
	return nil
}
