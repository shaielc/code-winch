package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/shaielc/code-winch/internal/adapters/transport/attach"
)

func TestChildOwnerReportsRunningAndStops(t *testing.T) {
	owner := helperOwner(t, "wait")
	if err := owner.Start(); err != nil {
		t.Fatal(err)
	}
	if got := owner.Status(); got != attach.ChildRunning {
		t.Fatalf("status after start = %q", got)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	owner.Stop(ctx)
	waitForStatus(t, owner, attach.ChildExited)
}

func TestChildOwnerFailedStartNeverReportsRunning(t *testing.T) {
	owner := newChildOwner(t.TempDir() + "/missing")
	if err := owner.Start(); err == nil {
		t.Fatal("missing executable started")
	}
	if got := owner.Status(); got != attach.ChildDown {
		t.Fatalf("status after failed start = %q", got)
	}
}

func TestChildOwnerReportsObservedExit(t *testing.T) {
	owner := helperOwner(t, "exit")
	if err := owner.Start(); err != nil {
		t.Fatal(err)
	}
	waitForStatus(t, owner, attach.ChildExited)
}

func helperOwner(t *testing.T, mode string) *childOwner {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return newChildOwner(executable, "-test.run=TestChildHelper", "--", mode)
}

func waitForStatus(t *testing.T, owner *childOwner, want attach.ChildStatus) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if owner.Status() == want {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("status = %q, want %q", owner.Status(), want)
}

func TestChildHelper(t *testing.T) {
	if len(os.Args) < 2 || os.Args[len(os.Args)-2] != "--" {
		return
	}
	if os.Args[len(os.Args)-1] == "wait" {
		_, _ = os.Stdin.Read(make([]byte, 1))
	}
	os.Exit(0)
}
