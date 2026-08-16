package local

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	harnessfake "github.com/shaielc/code-winch/internal/adapters/harness/fake"
	sandboxfake "github.com/shaielc/code-winch/internal/adapters/sandbox/fake"
	sandboxlocal "github.com/shaielc/code-winch/internal/adapters/sandbox/local"
	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/pkg/protocol"
)

const (
	testExecutionID = "test-execution"
	testLease       = "test-lease"
	testWorkspaceID = "00000000-0000-0000-0000-000000000001"
)

type sender func(kind string, payload any) error

// TestObservationsAreOrderedAndFlushPrecedesExit drives a real PTY so the
// ordering guarantees are asserted against the substrate the runner ships on.
// The script's last write has no trailing newline, so the codec can only report
// it from Flush — which must therefore run before the exit observation.
func TestObservationsAreOrderedAndFlushPrecedesExit(t *testing.T) {
	r, send := newPTYRunner(t, `#!/bin/sh
printf '{"kind":"harness.ready","payload":{"step":1}}\n'
IFS= read -r line
printf '{"kind":"harness.echo","payload":{"step":2}}\n'
printf 'trailing-partial-record'
exit 0
`, nil)
	prepareAndStart(t, send)
	if err := send("input", protocol.InputPayload{InputID: "i-1", Text: "go"}); err != nil {
		t.Fatalf("input: %v", err)
	}

	observations := collectUntilExit(t, r)
	if observations[0].Type != "start" {
		t.Fatalf("first observation = %q, want start", observations[0].Type)
	}
	for i, o := range observations {
		if o.Ordinal != uint64(i+1) {
			t.Fatalf("observation %d has ordinal %d, want %d", i, o.Ordinal, i+1)
		}
		if o.ExecutionID != testExecutionID {
			t.Fatalf("observation %d has execution %q", i, o.ExecutionID)
		}
	}

	var kinds []string
	for _, o := range observations {
		if o.Type == "output" {
			kinds = append(kinds, o.Event.Kind)
		}
	}
	// raw.output and diagnostic are the codec's report of the unterminated
	// record; both must appear before the exit observation that follows them.
	want := []string{"harness.ready", "harness.echo", "raw.output", "diagnostic"}
	if strings.Join(kinds, ",") != strings.Join(want, ",") {
		t.Fatalf("output kinds = %v, want %v", kinds, want)
	}

	exit := observations[len(observations)-1]
	if !exit.Exit.Successful || exit.Exit.Code != "OK" {
		t.Fatalf("exit = %+v, want successful OK", *exit.Exit)
	}

	// Close must end a range over the channel rather than leaving it open.
	r.Close()
	if _, open := <-r.Observations(); open {
		t.Fatal("Observations stayed open after Close")
	}
}

// TestExitCodeSurvivesSelfTermination pins the ordering bug where the stream
// reached EOF before the driver had reaped the process, so the exit code was
// still unknown and a clean exit was reported as a failure.
func TestExitCodeSurvivesSelfTermination(t *testing.T) {
	for _, tc := range []struct {
		name       string
		script     string
		successful bool
		code       string
	}{
		{"clean exit", "#!/bin/sh\nexit 0\n", true, "OK"},
		{"failed exit", "#!/bin/sh\nexit 3\n", false, "PROCESS_FAILED"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, send := newPTYRunner(t, tc.script, nil)
			prepareAndStart(t, send)
			exit := collectUntilExit(t, r)
			last := exit[len(exit)-1]
			if last.Exit.Successful != tc.successful || last.Exit.Code != tc.code {
				t.Fatalf("exit = %+v, want successful=%t code=%s", *last.Exit, tc.successful, tc.code)
			}
		})
	}
}

// TestInputDoesNotStarveThePump is the regression test for the lock inversion:
// input was written while holding the lock the pump needed to emit output, so
// a write that blocked until the pump drained could never be satisfied.
func TestInputDoesNotStarveThePump(t *testing.T) {
	r, send := newFakeRunner(t)
	prepareAndStart(t, send)
	drained := drainObservations(r)

	const inputs = 200
	done := make(chan error, 1)
	go func() {
		for i := 0; i < inputs; i++ {
			if err := send("input", protocol.InputPayload{InputID: fmt.Sprintf("i-%d", i), Text: "hello"}); err != nil {
				done <- err
				return
			}
		}
		done <- nil
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("input: %v", err)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("input blocked: a write is holding the lock the pump needs to drain the stream")
	}

	// Commands must still be served while input is in flight.
	if err := send("stop", protocol.StopPayload{GraceMilliseconds: 10}); err != nil {
		t.Fatalf("stop: %v", err)
	}
	if err := send("cleanup", nil); err != nil {
		t.Fatalf("cleanup: %v", err)
	}
	r.Close()
	<-drained
}

// TestConcurrentInputAndOutputShareTheCodec drives encoding and decoding at the
// same time on the one codec an execution owns. A PTY buffers writes, so the
// caller does not wait for the pump and the two genuinely overlap; the harness
// echoes every line to keep output flowing for as long as input does.
//
// Neither codec in the tree keeps state in Encode today, so this does not fail
// without the runner's codecMu. It guards the requirement rather than a known
// bug: HarnessCodec promises no concurrency safety, and under -race this is
// where a codec that starts keeping state would be caught.
func TestConcurrentInputAndOutputShareTheCodec(t *testing.T) {
	r, send := newPTYRunner(t, `#!/bin/sh
while IFS= read -r line ; do
printf '{"kind":"harness.echo","payload":{"step":1}}\n'
done
`, nil)
	drained := drainObservations(r)
	prepareAndStart(t, send)

	var wg sync.WaitGroup
	for writer := 0; writer < 16; writer++ {
		wg.Add(1)
		go func(writer int) {
			defer wg.Done()
			for i := 0; i < 10; i++ {
				if err := send("input", protocol.InputPayload{
					InputID: fmt.Sprintf("i-%d-%d", writer, i), Text: "go",
				}); err != nil {
					return // The stream is gone; later writers stop with it.
				}
			}
		}(writer)
	}
	wg.Wait()

	if err := send("cleanup", nil); err != nil {
		t.Fatalf("cleanup: %v", err)
	}
	r.Close()
	<-drained
}

// TestForcedStopLeavesNoDescendant covers the acceptance criterion at the
// runner seam: the sandbox owns the process group, and cleanup through the
// runner must take the whole group with it.
func TestForcedStopLeavesNoDescendant(t *testing.T) {
	pidFile := filepath.Join(t.TempDir(), "child.pid")
	r, send := newPTYRunner(t, `#!/bin/sh
sleep 30 &
echo $! > "$PID_FILE"
printf '{"kind":"harness.ready","payload":{"step":1}}\n'
wait
`, map[string]string{"PID_FILE": pidFile})
	drained := drainObservations(r)
	prepareAndStart(t, send)

	childPID := waitForPID(t, pidFile)
	if err := send("stop", protocol.StopPayload{GraceMilliseconds: 50}); err != nil {
		t.Fatalf("stop: %v", err)
	}
	if err := send("cleanup", nil); err != nil {
		t.Fatalf("cleanup: %v", err)
	}
	// Cleanup is idempotent; the second call must not report a failure.
	if err := r.Cleanup(context.Background(), testExecutionID); err != nil {
		t.Fatalf("repeated cleanup: %v", err)
	}

	deadline := time.Now().Add(5 * time.Second)
	for processExists(childPID) && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if processExists(childPID) {
		t.Fatalf("descendant %d survived a forced stop", childPID)
	}
	r.Close()
	<-drained
}

// TestRejectionsAreStableAndContentFree checks every refusal the brief names,
// and that none of them echo the payload back: these errors reach logs, which
// docs/security.md forbids from carrying message content.
func TestRejectionsAreStableAndContentFree(t *testing.T) {
	const secret = "do-not-leak-this-content"
	r, send := newFakeRunner(t)

	if err := requireCode(t, send("teleport", nil), "RUNNER_UNKNOWN_MESSAGE_KIND"); err != nil {
		t.Fatal(err)
	}
	if err := requireCode(t, r.Send(context.Background(), protocol.RunnerMessage{
		Kind: "inspect", CommandID: "c", ExecutionID: "unknown", LeaseToken: testLease,
	}), "RUNNER_EXECUTION_NOT_FOUND"); err != nil {
		t.Fatal(err)
	}
	if err := requireCode(t, r.Send(context.Background(), protocol.RunnerMessage{
		Kind: "inspect", CommandID: "c", ExecutionID: testExecutionID,
	}), "RUNNER_REQUIRED_FIELD"); err != nil {
		t.Fatal(err)
	}

	if err := send("prepare", protocol.PreparePayload{WorkspaceID: testWorkspaceID}); err != nil {
		t.Fatalf("prepare: %v", err)
	}

	// Input before the process is ready fails rather than being dropped.
	notReady := send("input", protocol.InputPayload{InputID: "i-0", Text: secret})
	if err := requireCode(t, notReady, "RUNNER_NOT_READY"); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(notReady.Error(), secret) {
		t.Fatalf("rejection leaked input content: %v", notReady)
	}

	stale := r.Send(context.Background(), protocol.RunnerMessage{
		Kind: "input", CommandID: "c", ExecutionID: testExecutionID, LeaseToken: "wrong-lease",
		Payload: mustMarshal(t, protocol.InputPayload{InputID: "i-0", Text: secret}),
	})
	if err := requireCode(t, stale, "RUNNER_STALE_LEASE"); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(stale.Error(), secret) || strings.Contains(stale.Error(), "wrong-lease") {
		t.Fatalf("rejection leaked credentials or content: %v", stale)
	}

	if err := send("prepare", protocol.PreparePayload{WorkspaceID: testWorkspaceID}); err == nil {
		t.Fatal("second prepare for the same execution was accepted")
	}
	r.Close()
}

func TestSecondStartIsRejected(t *testing.T) {
	r, send := newPTYRunner(t, "#!/bin/sh\ncat\n", nil)
	drained := drainObservations(r)
	prepareAndStart(t, send)
	if err := requireCode(t, send("start", protocol.StartPayload{LaunchProfile: "p"}), "RUNNER_ALREADY_STARTED"); err != nil {
		t.Fatal(err)
	}
	if err := send("cleanup", nil); err != nil {
		t.Fatalf("cleanup: %v", err)
	}
	r.Close()
	<-drained
}

// ---------------------------------------------------------------- test setup

func newFakeRunner(t *testing.T) (*Runner, sender) {
	t.Helper()
	r := New(sandboxfake.New(application.SandboxCapabilities{Attach: true, AttachSingleUse: true}), harnessfake.Driver{})
	t.Cleanup(func() {
		_ = r.Cleanup(context.Background(), testExecutionID)
		r.Close()
	})
	return r, newSender(t, r)
}

func newPTYRunner(t *testing.T, script string, env map[string]string) (*Runner, sender) {
	t.Helper()
	dir := t.TempDir()
	const command = "runner-contract-script"
	if err := os.WriteFile(filepath.Join(dir, command), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	r := New(sandboxlocal.New(), scriptHarness{command: command, env: env})
	t.Cleanup(func() {
		_ = r.Cleanup(context.Background(), testExecutionID)
		r.Close()
	})
	return r, newSender(t, r)
}

func newSender(t *testing.T, r *Runner) sender {
	t.Helper()
	var n atomic.Uint64
	return func(kind string, payload any) error {
		return r.Send(context.Background(), protocol.RunnerMessage{
			Version: protocol.RunnerVersion{Major: 1}, Kind: kind,
			CommandID: fmt.Sprintf("c-%d", n.Add(1)), ExecutionID: testExecutionID,
			LeaseToken: testLease, Payload: mustMarshal(t, payload),
		})
	}
}

func prepareAndStart(t *testing.T, send sender) {
	t.Helper()
	if err := send("prepare", protocol.PreparePayload{WorkspaceID: testWorkspaceID}); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	if err := send("start", protocol.StartPayload{LaunchProfile: "fake"}); err != nil {
		t.Fatalf("start: %v", err)
	}
}

func mustMarshal(t *testing.T, payload any) []byte {
	t.Helper()
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func collectUntilExit(t *testing.T, r *Runner) []Observation {
	t.Helper()
	var got []Observation
	deadline := time.After(30 * time.Second)
	for {
		select {
		case o, open := <-r.Observations():
			if !open {
				t.Fatalf("observations closed after %d, before an exit", len(got))
			}
			got = append(got, o)
			if o.Type == "exit" {
				return got
			}
		case <-deadline:
			t.Fatalf("timed out after %d observations without an exit", len(got))
		}
	}
}

func drainObservations(r *Runner) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		for range r.Observations() {
		}
	}()
	return done
}

func requireCode(t *testing.T, err error, code string) error {
	t.Helper()
	if err == nil {
		return fmt.Errorf("expected %s, got success", code)
	}
	var protocolErr *protocol.RunnerProtocolError
	if !errors.As(err, &protocolErr) {
		return fmt.Errorf("expected %s, got %T: %v", code, err, err)
	}
	if protocolErr.Code != code {
		return fmt.Errorf("code = %s, want %s", protocolErr.Code, code)
	}
	return nil
}

// scriptHarness launches a shell script from PATH and decodes it with the fake
// harness codec, so runner tests can script exact output over a real PTY.
type scriptHarness struct {
	command string
	env     map[string]string
}

func (scriptHarness) Describe(context.Context) (application.HarnessDescriptor, error) {
	return application.HarnessDescriptor{ID: "script", Version: "1.0.0"}, nil
}
func (h scriptHarness) BuildLaunch(context.Context, application.RunSpec, application.ResolvedCredentials) (application.LaunchSpec, error) {
	return application.LaunchSpec{Command: h.command, Env: h.env}, nil
}
func (scriptHarness) NewCodec(context.Context, application.RunSpec) (application.HarnessCodec, error) {
	return &harnessfake.Codec{}, nil
}
func (scriptHarness) MapExit(code int) application.HarnessExit {
	return harnessfake.Driver{}.MapExit(code)
}

func waitForPID(t *testing.T, path string) int {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(path)
		if err == nil && len(strings.TrimSpace(string(data))) > 0 {
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

var _ application.HarnessDriver = scriptHarness{}
