package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

const secret = "e2e-development-secret-000000000000"

func TestCreateThenGet(t *testing.T) {
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
	command.Env = append(os.Environ(), "WINCH_ADDR="+address, "WINCH_DATABASE_URL="+database, "WINCH_ALLOWED_ORIGIN=http://"+address, "WINCH_TOKEN="+secret, "WINCH_CSRF_TOKEN="+secret, "WINCH_STATIC_DIR="+t.TempDir())
	var logs bytes.Buffer
	command.Stdout, command.Stderr = &logs, &logs
	if err = command.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { cancel(); _ = command.Wait() }()
	base := "http://" + address
	waitHealthy(t, base, &logs)
	request := map[string]string{"workspacePath": "/tmp/ws", "harnessProfile": "fake", "sandboxProfile": "local"}
	body, _ := json.Marshal(request)
	created := call(t, http.MethodPost, base+"/api/v1/runs", body)
	if created.StatusCode != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", created.StatusCode, created.Body)
	}
	var run map[string]any
	if err = json.Unmarshal(created.Body, &run); err != nil {
		t.Fatal(err)
	}
	read := call(t, http.MethodGet, base+"/api/v1/runs/"+run["id"].(string), nil)
	if read.StatusCode != http.StatusOK {
		t.Fatalf("get status=%d body=%s", read.StatusCode, read.Body)
	}
	var got map[string]any
	_ = json.Unmarshal(read.Body, &got)
	if got["state"] != "created" || got["lastSequence"] != float64(0) || got["workspacePath"] != "/tmp/ws" {
		t.Fatalf("unexpected run: %s", read.Body)
	}
}

type response struct {
	StatusCode int
	Body       []byte
}

func call(t *testing.T, method, url string, body []byte) response {
	t.Helper()
	req, _ := http.NewRequest(method, url, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+secret)
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF-Token", secret)
		req.Header.Set("Origin", strings.TrimSuffix(url, "/api/v1/runs"))
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
func waitHealthy(t *testing.T, base string, logs *bytes.Buffer) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		r, err := http.Get(base + "/api/v1/health")
		if err == nil {
			_ = r.Body.Close()
			if r.StatusCode == 200 {
				return
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("daemon did not become healthy: %s", logs.String())
}
