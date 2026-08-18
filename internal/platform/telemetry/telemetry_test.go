package telemetry

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestHandlerDropsFreeFormAndSecret(t *testing.T) {
	var b bytes.Buffer
	l := slog.New(NewHandler(slog.NewTextHandler(&b, nil)))
	l.InfoContext(context.Background(), "start", "component", "daemon", "message_content", "canary-secret")
	s := b.String()
	if strings.Contains(s, "message_content") || strings.Contains(s, "canary-secret") {
		t.Fatalf("unsafe log: %s", s)
	}
	if !strings.Contains(s, "component=daemon") {
		t.Fatal(s)
	}
}
func TestMetricLabelsBounded(t *testing.T) {
	r := NewRegistry()
	r.Declare("winch_active_runs", "status", "running", "stopping")
	if e := r.Validate("winch_active_runs", map[string]string{"status": "running"}); e != nil {
		t.Fatal(e)
	}
	if e := r.Validate("winch_active_runs", map[string]string{"status": "arbitrary"}); e == nil {
		t.Fatal("accepted unbounded value")
	}
}

// Every metric name is pre-registered, so a name on its own proves nothing:
// the label is what Declare bounds, and an emitter that invents one has to be
// refused before it reaches the log.
func TestUndeclaredMetricsAndLabelsAreRefused(t *testing.T) {
	r := NewRegistry()
	if e := r.Validate("winch_queue_time_seconds", map[string]string{"status": "preparing"}); e == nil {
		t.Fatal("accepted a label nothing declared")
	}
	r.Declare("winch_queue_time_seconds", "status", "preparing", "cancelled")
	if e := r.Validate("winch_queue_time_seconds", map[string]string{"status": "preparing"}); e != nil {
		t.Fatal(e)
	}
	if e := r.Validate("winch_queue_time_seconds", map[string]string{"reason": "preparing"}); e == nil {
		t.Fatal("accepted an undeclared label name")
	}
	if e := r.Validate("winch_invented_total", nil); e == nil {
		t.Fatal("accepted an unregistered metric")
	}
}
