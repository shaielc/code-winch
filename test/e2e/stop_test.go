package e2e

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// TestCreateStartStopGet exercises stop through the deployed HTTP surface and
// inspects only the process identified by this test's unique transcript path.
func TestCreateStartStopGet(t *testing.T) {
	database := requireDatabase(t)
	transcript := writeTranscript(t, "echo stop scenario is alive", "exit")
	base := startDaemon(t, database,
		"WINCH_FAKE_HARNESS_TRANSCRIPT="+transcript,
		"WINCH_FAKE_HARNESS_DELAY=10s",
	)

	run := createRun(t, base, "stop-1", t.TempDir(), "fake", "local")
	runID := run["id"].(string)
	started := request(t, http.MethodPost, base+"/api/v1/runs/"+runID+"/start", nil, map[string]string{
		"Idempotency-Key": "start-stop-scenario",
		"If-Match":        etag(run),
	})
	if started.StatusCode != http.StatusAccepted {
		t.Fatalf("start status=%d body=%s", started.StatusCode, started.Body)
	}
	var running map[string]any
	if err := json.Unmarshal(started.Body, &running); err != nil {
		t.Fatal(err)
	}

	pid, pgid := findOwnedProcess(t, transcript)
	if err := syscall.Kill(pid, 0); err != nil {
		t.Fatalf("started harness pid %d is not alive: %v", pid, err)
	}
	stopped := request(t, http.MethodPost, base+"/api/v1/runs/"+runID+"/stop", []byte(`{"reason":"e2e stop"}`), map[string]string{
		"Idempotency-Key": "stop-scenario",
		"If-Match":        etag(running),
	})
	if stopped.StatusCode != http.StatusAccepted {
		t.Fatalf("stop status=%d body=%s", stopped.StatusCode, stopped.Body)
	}

	final := awaitTerminal(t, base, runID)
	if final["state"] != "completed" && final["state"] != "failed" {
		t.Fatalf("stopped run is not truthfully terminal: %v", final)
	}
	if final["id"] != runID || final["workspacePath"] != run["workspacePath"] {
		t.Fatalf("stop changed run identity: created=%v final=%v", run, final)
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) && (processExists(pid) || processGroupExists(pgid)) {
		time.Sleep(20 * time.Millisecond)
	}
	if processExists(pid) {
		t.Fatalf("test-owned harness leader %d was not reaped", pid)
	}
	if processGroupExists(pgid) {
		t.Fatalf("test-owned process group %d still exists", pgid)
	}
}

func findOwnedProcess(t *testing.T, marker string) (int, int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		entries, err := filepath.Glob("/proc/[0-9]*/cmdline")
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			cmdline, err := os.ReadFile(entry)
			if err != nil || !strings.Contains(string(cmdline), marker) {
				continue
			}
			pid, _ := strconv.Atoi(filepath.Base(filepath.Dir(entry)))
			pgid, err := syscall.Getpgid(pid)
			if err == nil {
				return pid, pgid
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("test-owned harness for %q did not start", marker)
	return 0, 0
}

func processExists(pid int) bool { return syscall.Kill(pid, 0) == nil }

func processGroupExists(pgid int) bool {
	err := syscall.Kill(-pgid, 0)
	return err == nil || !errors.Is(err, syscall.ESRCH)
}
