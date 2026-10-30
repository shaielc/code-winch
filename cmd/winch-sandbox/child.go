package main

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/shaielc/code-winch/internal/adapters/transport/attach"
)

// childOwner is the concurrency-safe observation of the exact process it starts.
type childOwner struct {
	mu     sync.RWMutex
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	status attach.ChildStatus
	done   chan struct{}
}

func shippedFakeHarness() string {
	executable, err := os.Executable()
	if err != nil {
		return "fake-harness"
	}
	return filepath.Join(filepath.Dir(executable), "fake-harness")
}

func newChildOwner(name string, args ...string) *childOwner {
	cmd := exec.Command(name, args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return &childOwner{cmd: cmd, status: attach.ChildDown, done: make(chan struct{})}
}

func (o *childOwner) Start() error {
	stdin, err := o.cmd.StdinPipe()
	if err != nil {
		return err
	}
	if err := o.cmd.Start(); err != nil {
		_ = stdin.Close()
		return err
	}
	o.mu.Lock()
	o.stdin = stdin
	o.status = attach.ChildRunning
	o.mu.Unlock()
	go func() {
		_ = o.cmd.Wait()
		o.mu.Lock()
		o.status = attach.ChildExited
		o.mu.Unlock()
		close(o.done)
	}()
	return nil
}

func (o *childOwner) Status() attach.ChildStatus {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.status
}

func (o *childOwner) Stop(ctx context.Context) {
	o.mu.RLock()
	status := o.status
	stdin := o.stdin
	o.mu.RUnlock()
	if status != attach.ChildRunning {
		return
	}
	if stdin != nil {
		_ = stdin.Close()
	}
	select {
	case <-o.done:
		return
	case <-ctx.Done():
		_ = o.cmd.Process.Kill()
	}
	select {
	case <-o.done:
	case <-ctx.Done():
	}
}
