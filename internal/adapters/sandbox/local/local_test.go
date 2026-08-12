package local

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/creack/pty"
	"github.com/shaielc/code-winch/internal/application"
	sandboxcontract "github.com/shaielc/code-winch/test/contract/sandbox"
)

func TestSandboxContract(t *testing.T) {
	dir := t.TempDir()
	command := filepath.Join(dir, "contract-command")
	if err := os.WriteFile(command, []byte("#!/bin/sh\ntrap 'exit 0' TERM\nwhile :; do sleep 1; done\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	sandboxcontract.Run(t, func(*testing.T) application.SandboxDriver { return New() })
}

func TestCapabilitiesAreTruthful(t *testing.T) {
	d := New()
	caps := d.Capabilities(context.Background())
	if caps.Isolation != "unisolated" || !caps.Resize || caps.NetworkPolicy || caps.ResourceLimits {
		t.Fatalf("unexpected capabilities: %#v", caps)
	}
	for _, spec := range []application.SandboxSpec{
		{ID: "network", DisableNetwork: true},
		{ID: "limits", MemoryBytes: 1},
	} {
		if _, err := d.Prepare(context.Background(), spec); err == nil || !strings.Contains(err.Error(), "UNSUPPORTED_CAPABILITY") {
			t.Fatalf("Prepare(%#v) error = %v", spec, err)
		}
	}
}

func TestResizeChangesPTYDimensions(t *testing.T) {
	d, prepared, handle := startShell(t, "while :; do sleep 1; done", nil)
	t.Cleanup(func() { _ = d.Cleanup(context.Background(), prepared) })
	if err := d.Resize(context.Background(), handle, application.TerminalSize{Rows: 43, Cols: 119}); err != nil {
		t.Fatal(err)
	}
	d.mu.Lock()
	rows, cols, err := pty.Getsize(d.executions[handle.ID].pty)
	d.mu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	if rows != 43 || cols != 119 {
		t.Fatalf("PTY size = %dx%d, want 43x119", rows, cols)
	}
}

func TestCleanupKillsChildAndToleratesRetries(t *testing.T) {
	pidFile := filepath.Join(t.TempDir(), "child.pid")
	d, prepared, handle := startShell(t, "sleep 30 & echo $! > \"$PID_FILE\"; wait", map[string]string{"PID_FILE": pidFile})
	childPID := waitForPID(t, pidFile)
	if err := d.Cleanup(context.Background(), prepared); err != nil {
		t.Fatal(err)
	}
	if err := d.Cleanup(context.Background(), prepared); err != nil {
		t.Fatalf("second cleanup: %v", err)
	}
	if err := d.Stop(context.Background(), handle, application.StopPolicy{}); err != nil {
		t.Fatalf("stop after cleanup: %v", err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for processExists(childPID) && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if processExists(childPID) {
		t.Fatalf("owned child process %d survived cleanup", childPID)
	}
}

func startShell(t *testing.T, script string, env map[string]string) (*Driver, application.PreparedSandbox, application.ExecutionHandle) {
	t.Helper()
	d := New()
	prepared, err := d.Prepare(context.Background(), application.SandboxSpec{ID: "test-sandbox"})
	if err != nil {
		t.Fatal(err)
	}
	handle, err := d.Start(context.Background(), prepared, application.LaunchSpec{Command: "/bin/sh", Args: []string{"-c", script}, Env: env})
	if err != nil {
		t.Fatal(err)
	}
	return d, prepared, handle
}

func waitForPID(t *testing.T, path string) int {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(path)
		if err == nil {
			pid, convertErr := strconv.Atoi(strings.TrimSpace(string(data)))
			if convertErr != nil {
				t.Fatal(convertErr)
			}
			return pid
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", path)
	return 0
}

func processExists(pid int) bool {
	err := syscall.Kill(pid, 0)
	return err == nil || !errors.Is(err, syscall.ESRCH)
}
