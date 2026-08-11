package testkit

import (
	"testing"
	"time"

	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
)

var _ application.Clock = (*Clock)(nil)
var _ application.IDSource = IDs{}

func TestClockAdvancesWithoutSleeping(t *testing.T) {
	initial, err := domain.ParseTimestamp("2026-08-11T10:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	clock := NewClock(initial)
	clock.Advance(90 * time.Second)
	if got, want := clock.Now().String(), "2026-08-11T10:01:30Z"; got != want {
		t.Fatalf("Now() = %q, want %q", got, want)
	}
}

func TestIDsAreDeterministic(t *testing.T) {
	run, err := domain.ParseRunID("123e4567-e89b-12d3-a456-426614174000")
	if err != nil {
		t.Fatal(err)
	}
	source := IDs{Run: run}
	if first, second := source.NewRunID(), source.NewRunID(); first != second || first != run {
		t.Fatalf("IDs were not deterministic: %v, %v", first, second)
	}
}

func TestClockRejectsZeroInitialTime(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("NewClock accepted zero time")
		}
	}()
	NewClock(domain.Timestamp{})
}
