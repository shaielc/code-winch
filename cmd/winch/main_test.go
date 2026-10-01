package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestRunStopReadsCurrentVersionAndSendsConditionalCommand(t *testing.T) {
	var sawGet, sawStop bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/runs/RUN":
			sawGet = true
			_ = json.NewEncoder(w).Encode(apiRun{ID: "RUN", State: "running", Version: 17})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/runs/RUN/stop":
			sawStop = true
			if got := r.Header.Get("If-Match"); got != `"17"` {
				t.Errorf("If-Match = %q", got)
			}
			if got := r.Header.Get("Idempotency-Key"); got != "stop-key" {
				t.Errorf("Idempotency-Key = %q", got)
			}
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if body["reason"] != "operator request" {
				t.Errorf("reason = %q", body["reason"])
			}
			_ = json.NewEncoder(w).Encode(apiRun{ID: "RUN", State: "stopping", Version: 18})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	t.Setenv("WINCH_API_URL", server.URL)
	t.Setenv("WINCH_TOKEN", "token")
	t.Setenv("WINCH_CSRF_TOKEN", "csrf")
	oldArgs := os.Args
	os.Args = []string{"winch", "run", "stop", "--idempotency-key", "stop-key", "--reason", "operator request", "RUN"}
	t.Cleanup(func() { os.Args = oldArgs })

	runStop()
	if !sawGet || !sawStop {
		t.Fatalf("requests: get=%t stop=%t", sawGet, sawStop)
	}
}
