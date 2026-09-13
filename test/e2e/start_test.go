package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCreateStartPollEventsThenGet(t *testing.T) {
	database := os.Getenv("PG_TEST_DATABASE_URL")
	if database == "" {
		t.Skip("PG_TEST_DATABASE_URL is required")
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	_ = listener.Close()
	binary := os.Getenv("WINCHD_BIN")
	if binary == "" {
		binary = "../../bin/winchd"
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	command := exec.CommandContext(ctx, binary)
	command.Env = append(os.Environ(), "PATH="+filepath.Dir(binary)+string(os.PathListSeparator)+os.Getenv("PATH"), "WINCH_ADDR="+address, "WINCH_DATABASE_URL="+database, "WINCH_ALLOWED_ORIGIN=http://"+address, "WINCH_TOKEN="+secret, "WINCH_CSRF_TOKEN="+secret, "WINCH_STATIC_DIR="+t.TempDir())
	var logs bytes.Buffer
	command.Stdout, command.Stderr = &logs, &logs
	if err = command.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { cancel(); _ = command.Wait() }()
	base := "http://" + address
	waitHealthy(t, base, &logs)

	body, _ := json.Marshal(map[string]string{"workspacePath": "/tmp/ws", "harnessProfile": "fake", "sandboxProfile": "local"})
	created := scenarioCall(t, http.MethodPost, base+"/api/v1/runs", body, map[string]string{"Idempotency-Key": uuid.NewString()})
	if created.StatusCode != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", created.StatusCode, created.Body)
	}
	var run map[string]any
	_ = json.Unmarshal(created.Body, &run)
	id, version := run["id"].(string), int64(run["version"].(float64))
	started := scenarioCall(t, http.MethodPost, base+"/api/v1/runs/"+id+"/start", nil, map[string]string{"Idempotency-Key": uuid.NewString(), "If-Match": `"` + strconv.FormatInt(version, 10) + `"`})
	if started.StatusCode != http.StatusAccepted {
		t.Fatalf("start status=%d body=%s logs=%s", started.StatusCode, started.Body, logs.String())
	}

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		read := scenarioCall(t, http.MethodGet, base+"/api/v1/runs/"+id, nil, nil)
		_ = json.Unmarshal(read.Body, &run)
		if run["state"] == "completed" {
			break
		}
		time.Sleep(25 * time.Millisecond)
	}
	if run["state"] != "completed" {
		t.Fatalf("run did not complete: %v logs=%s", run, logs.String())
	}
	events := scenarioCall(t, http.MethodGet, base+"/api/v1/runs/"+id+"/events?limit=200", nil, nil)
	var page struct {
		Events []struct {
			Sequence   int64     `json:"sequence"`
			Kind       string    `json:"kind"`
			OccurredAt time.Time `json:"occurredAt"`
		} `json:"events"`
	}
	_ = json.Unmarshal(events.Body, &page)
	if len(page.Events) == 0 {
		t.Fatalf("no events: %s", events.Body)
	}
	for i, event := range page.Events {
		if event.Sequence != int64(i+1) {
			t.Fatalf("event gap at %d: %s", i, events.Body)
		}
		if event.OccurredAt.IsZero() {
			t.Fatalf("event has zero occurredAt: %s", events.Body)
		}
	}
}

func scenarioCall(t *testing.T, method, url string, body []byte, headers map[string]string) response {
	t.Helper()
	req, _ := http.NewRequest(method, url, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+secret)
	if method != http.MethodGet {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF-Token", secret)
		req.Header.Set("Origin", "http://"+req.URL.Host)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = r.Body.Close() }()
	var out bytes.Buffer
	_, _ = out.ReadFrom(r.Body)
	return response{r.StatusCode, out.Bytes()}
}
