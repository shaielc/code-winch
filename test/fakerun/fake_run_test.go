//go:build linux

package fakerun_test

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// TestShippedFakeRun exercises the same process boundary as an operator: both
// executables are built from their shipping entrypoints and winch launches the
// fake harness through the local PTY sandbox. Keeping this outside cmd and the
// runner package prevents the test from substituting an in-process adapter.
func TestShippedFakeRun(t *testing.T) {
	repository := repositoryRoot(t)
	binDir := t.TempDir()
	winch := build(t, repository, binDir, "winch", "./cmd/winch")
	fakeHarness := build(t, repository, binDir, "fake-harness", "./cmd/fake-harness")
	transcript := filepath.Join(t.TempDir(), "successful.transcript")
	if err := os.WriteFile(transcript, []byte("echo hello from CI\nexit\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	command := exec.Command(winch,
		"dev", "run", "--harness", "fake", "--sandbox", "local",
		"--fake-binary", fakeHarness, "--fake-transcript", transcript,
		"--fake-delay", "250ms",
	)
	output, harnessPID := runAndObserveHarness(t, command, fakeHarness)
	for _, expected := range []string{
		"[started]",
		"fake harness ready: run_id=00000000-0000-0000-0000-000000000001",
		"hello from CI",
		"fake harness exiting",
		"[exit] successful=true code=OK",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("output did not contain %q:\n%s", expected, output)
		}
	}
	assertProcessGone(t, harnessPID)
	t.Logf("complete fake run output:\n%s", output)

	stopTranscript := filepath.Join(t.TempDir(), "stopped.transcript")
	if err := os.WriteFile(stopTranscript, []byte("echo this action is deliberately delayed\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	stopCommand := exec.Command(winch,
		"dev", "run", "--harness", "fake", "--sandbox", "local",
		"--fake-binary", fakeHarness, "--fake-transcript", stopTranscript,
		"--fake-delay", "5s", "--stop-after", "250ms",
	)
	stopOutput, stoppedHarnessPID := runAndObserveHarness(t, stopCommand, fakeHarness)
	for _, expected := range []string{
		"fake harness received a termination signal",
		"[exit] successful=true code=OK",
	} {
		if !strings.Contains(stopOutput, expected) {
			t.Fatalf("stopped-run output did not contain %q:\n%s", expected, stopOutput)
		}
	}
	assertProcessGone(t, stoppedHarnessPID)
	t.Logf("stop escalation output:\n%s", stopOutput)
}

func build(t *testing.T, repository, binDir, name, pkg string) string {
	t.Helper()
	binary := filepath.Join(binDir, name)
	command := exec.Command("go", "build", "-o", binary, pkg)
	command.Dir = repository
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build %s: %v\n%s", name, err, output)
	}
	return binary
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	directory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(directory, "go.mod")); err == nil {
			return directory
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			t.Fatal("could not find repository root")
		}
		directory = parent
	}
}

func runAndObserveHarness(t *testing.T, command *exec.Cmd, fakeHarness string) (string, int) {
	t.Helper()
	var output strings.Builder
	command.Stdout = &output
	command.Stderr = &output
	if err := command.Start(); err != nil {
		t.Fatalf("start winch: %v", err)
	}

	harnessPID := waitForExecutable(t, fakeHarness, 5*time.Second)
	if err := command.Wait(); err != nil {
		t.Fatalf("winch dev run: %v\n%s", err, output.String())
	}
	return output.String(), harnessPID
}

// waitForExecutable observes the actual child rather than merely inferring
// that it ran from output. /proc is available on the Linux hosts used by CI.
func waitForExecutable(t *testing.T, executable string, timeout time.Duration) int {
	t.Helper()
	want, err := filepath.EvalSymlinks(executable)
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		entries, err := os.ReadDir("/proc")
		if err != nil {
			t.Fatalf("read /proc: %v", err)
		}
		for _, entry := range entries {
			pid, err := strconv.Atoi(entry.Name())
			if err != nil {
				continue
			}
			actual, err := os.Readlink(filepath.Join("/proc", entry.Name(), "exe"))
			if err == nil && strings.TrimSuffix(actual, " (deleted)") == want {
				return pid
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("did not observe %s running", executable)
	return 0
}

func assertProcessGone(t *testing.T, pid int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		err := syscall.Kill(pid, 0)
		if errors.Is(err, syscall.ESRCH) {
			return
		}
		if err != nil && !errors.Is(err, syscall.EPERM) {
			t.Fatalf("check fake-harness pid %d: %v", pid, err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("fake-harness descendant pid %d remained after winch exited: %s", pid, fmt.Sprint(commandLine(pid)))
}

func commandLine(pid int) string {
	data, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "cmdline"))
	if err != nil {
		return err.Error()
	}
	return strings.ReplaceAll(string(data), "\x00", " ")
}
