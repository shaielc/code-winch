// Package local runs trusted commands directly on the host under a PTY.
// It is a process-lifecycle boundary, not a security isolation boundary.
package local

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/creack/pty"
	"github.com/shaielc/code-winch/internal/application"
)

const defaultGracePeriod = 5 * time.Second

type execution struct {
	cmd      *exec.Cmd
	pty      *os.File
	pgid     int
	done     chan struct{}
	observed application.ObservedExecution
	attached bool
}

// Driver owns all processes it starts. A Driver must not be copied after use.
type Driver struct {
	mu         sync.Mutex
	prepared   map[string]bool
	executions map[string]*execution
	bySandbox  map[string]map[string]bool
	newID      func() (string, error)
}

// New constructs a local driver with cryptographically opaque execution IDs.
func New() *Driver {
	return &Driver{
		prepared:   make(map[string]bool),
		executions: make(map[string]*execution),
		bySandbox:  make(map[string]map[string]bool),
		newID:      randomID,
	}
}

func (*Driver) Capabilities(context.Context) application.SandboxCapabilities {
	return application.SandboxCapabilities{Isolation: "unisolated", Resize: true, Attach: true, AttachSingleUse: true}
}

func (d *Driver) Attach(_ context.Context, handle application.ExecutionHandle) (io.ReadWriteCloser, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	e, ok := d.executions[handle.ID]
	if !ok {
		return nil, fmt.Errorf("local sandbox: code=NOT_FOUND execution_id=%s", handle.ID)
	}
	if e.attached {
		return nil, fmt.Errorf("local sandbox: code=ALREADY_ATTACHED execution_id=%s", handle.ID)
	}
	e.attached = true
	return e.pty, nil
}

func (d *Driver) Prepare(_ context.Context, spec application.SandboxSpec) (application.PreparedSandbox, error) {
	if spec.ID == "" {
		return application.PreparedSandbox{}, errors.New("local sandbox: code=INVALID_SPEC field=sandbox_id")
	}
	if spec.DisableNetwork {
		return application.PreparedSandbox{}, fmt.Errorf("local sandbox: code=UNSUPPORTED_CAPABILITY sandbox_id=%s capability=network_policy", spec.ID)
	}
	if spec.MemoryBytes != 0 {
		return application.PreparedSandbox{}, fmt.Errorf("local sandbox: code=UNSUPPORTED_CAPABILITY sandbox_id=%s capability=resource_limits", spec.ID)
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.prepared[spec.ID] = true
	return application.PreparedSandbox{ID: spec.ID}, nil
}

func (d *Driver) Start(ctx context.Context, prepared application.PreparedSandbox, launch application.LaunchSpec) (application.ExecutionHandle, error) {
	if err := ctx.Err(); err != nil {
		return application.ExecutionHandle{}, fmt.Errorf("local sandbox: code=START_CANCELED sandbox_id=%s: %w", prepared.ID, err)
	}
	if launch.Command == "" {
		return application.ExecutionHandle{}, fmt.Errorf("local sandbox: code=INVALID_LAUNCH sandbox_id=%s field=command", prepared.ID)
	}
	d.mu.Lock()
	if !d.prepared[prepared.ID] {
		d.mu.Unlock()
		return application.ExecutionHandle{}, fmt.Errorf("local sandbox: code=NOT_PREPARED sandbox_id=%s", prepared.ID)
	}
	id, err := d.newID()
	if err != nil {
		d.mu.Unlock()
		return application.ExecutionHandle{}, fmt.Errorf("local sandbox: code=ID_GENERATION_FAILED sandbox_id=%s: %w", prepared.ID, err)
	}
	d.mu.Unlock()

	cmd := exec.Command(launch.Command, launch.Args...)
	cmd.Env = os.Environ()
	for key, value := range launch.Env {
		cmd.Env = append(cmd.Env, key+"="+value)
	}
	terminal, err := pty.StartWithSize(cmd, &pty.Winsize{Rows: 24, Cols: 80})
	if err != nil {
		return application.ExecutionHandle{}, fmt.Errorf("local sandbox: code=START_FAILED sandbox_id=%s: %w", prepared.ID, err)
	}
	disableEcho(terminal)
	e := &execution{cmd: cmd, pty: terminal, pgid: cmd.Process.Pid, done: make(chan struct{}), observed: application.ObservedExecution{Running: true}}
	d.mu.Lock()
	// Cleanup may have raced with process creation; never publish an unowned process.
	if !d.prepared[prepared.ID] {
		d.mu.Unlock()
		_ = signalGroup(e.pgid, syscall.SIGKILL)
		_ = cmd.Wait()
		_ = terminal.Close()
		return application.ExecutionHandle{}, fmt.Errorf("local sandbox: code=NOT_PREPARED sandbox_id=%s", prepared.ID)
	}
	d.executions[id] = e
	if d.bySandbox[prepared.ID] == nil {
		d.bySandbox[prepared.ID] = make(map[string]bool)
	}
	d.bySandbox[prepared.ID][id] = true
	d.mu.Unlock()
	go d.wait(id, e)
	return application.ExecutionHandle{ID: id}, nil
}

func disableEcho(terminal *os.File) {
	var state syscall.Termios
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, terminal.Fd(), syscall.TCGETS, uintptr(unsafe.Pointer(&state))); errno != 0 {
		return
	}
	state.Lflag &^= syscall.ECHO
	_, _, _ = syscall.Syscall(syscall.SYS_IOCTL, terminal.Fd(), syscall.TCSETS, uintptr(unsafe.Pointer(&state)))
}

func (d *Driver) wait(id string, e *execution) {
	err := e.cmd.Wait()
	code := exitCode(err)
	d.mu.Lock()
	if d.executions[id] == e {
		e.observed = application.ObservedExecution{ExitCode: &code}
		close(e.done)
	}
	d.mu.Unlock()
}

func (d *Driver) Inspect(_ context.Context, handle application.ExecutionHandle) (application.ObservedExecution, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	e, ok := d.executions[handle.ID]
	if !ok {
		return application.ObservedExecution{}, fmt.Errorf("local sandbox: code=NOT_FOUND execution_id=%s", handle.ID)
	}
	return e.observed, nil
}

func (d *Driver) Resize(_ context.Context, handle application.ExecutionHandle, size application.TerminalSize) error {
	if size.Rows == 0 || size.Cols == 0 {
		return fmt.Errorf("local sandbox: code=INVALID_SIZE execution_id=%s", handle.ID)
	}
	d.mu.Lock()
	e, ok := d.executions[handle.ID]
	d.mu.Unlock()
	if !ok {
		return fmt.Errorf("local sandbox: code=NOT_FOUND execution_id=%s", handle.ID)
	}
	if err := pty.Setsize(e.pty, &pty.Winsize{Rows: size.Rows, Cols: size.Cols}); err != nil {
		return fmt.Errorf("local sandbox: code=RESIZE_FAILED execution_id=%s: %w", handle.ID, err)
	}
	return nil
}

func (d *Driver) Stop(ctx context.Context, handle application.ExecutionHandle, policy application.StopPolicy) error {
	d.mu.Lock()
	e, ok := d.executions[handle.ID]
	d.mu.Unlock()
	if !ok { // Stop is deliberately idempotent, including after cleanup.
		return nil
	}
	return stopExecution(ctx, e, policy.GracePeriod)
}

func stopExecution(ctx context.Context, e *execution, grace time.Duration) error {
	if grace <= 0 {
		grace = defaultGracePeriod
	}
	_ = signalGroup(e.pgid, syscall.SIGTERM)
	timer := time.NewTimer(grace)
	defer timer.Stop()
	select {
	case <-e.done:
	case <-timer.C:
	case <-ctx.Done():
		_ = signalGroup(e.pgid, syscall.SIGKILL)
		return fmt.Errorf("local sandbox: code=STOP_CANCELED: %w", ctx.Err())
	}
	// The leader may exit before its children. Always kill the owned group so
	// cleanup cannot leave descendants behind.
	_ = signalGroup(e.pgid, syscall.SIGKILL)
	return nil
}

func (d *Driver) Cleanup(ctx context.Context, prepared application.PreparedSandbox) error {
	d.mu.Lock()
	ids := d.bySandbox[prepared.ID]
	items := make([]*execution, 0, len(ids))
	for id := range ids {
		items = append(items, d.executions[id])
	}
	delete(d.prepared, prepared.ID)
	d.mu.Unlock()
	for _, e := range items {
		if err := stopExecution(ctx, e, 100*time.Millisecond); err != nil && ctx.Err() != nil {
			return fmt.Errorf("local sandbox: code=CLEANUP_CANCELED sandbox_id=%s: %w", prepared.ID, err)
		}
		_ = e.pty.Close()
	}
	return nil
}

func signalGroup(pgid int, signal syscall.Signal) error {
	err := syscall.Kill(-pgid, signal)
	if errors.Is(err, syscall.ESRCH) {
		return nil
	}
	return err
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}

func randomID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(value[:]), nil
}

var _ application.SandboxDriver = (*Driver)(nil)
