// Package testkit provides deterministic application-port implementations.
package testkit

import (
	"sync"
	"time"

	"github.com/shaielc/code-winch/internal/domain"
)

type Clock struct {
	mu  sync.Mutex
	now domain.Timestamp
}

func NewClock(now domain.Timestamp) *Clock {
	if now.IsZero() {
		panic("test clock: zero initial timestamp")
	}
	return &Clock{now: now}
}
func (clock *Clock) Now() domain.Timestamp {
	clock.mu.Lock()
	defer clock.mu.Unlock()
	return clock.now
}
func (clock *Clock) Advance(elapsed time.Duration) {
	clock.mu.Lock()
	defer clock.mu.Unlock()
	next, err := domain.NewTimestamp(clock.now.Time().Add(elapsed))
	if err != nil {
		panic(err)
	}
	clock.now = next
}

// IDs returns the same configured IDs on every call, making retries and replay
// assertions deterministic. Tests that need a sequence can replace this small
// port directly.
type IDs struct {
	Workspace  domain.WorkspaceID
	Run        domain.RunID
	Attempt    domain.AttemptID
	Event      domain.EventID
	Command    domain.CommandID
	Artifact   domain.ArtifactID
	Credential domain.CredentialID
	Workflow   domain.WorkflowID
}

func (ids IDs) NewWorkspaceID() domain.WorkspaceID   { return ids.Workspace }
func (ids IDs) NewRunID() domain.RunID               { return ids.Run }
func (ids IDs) NewAttemptID() domain.AttemptID       { return ids.Attempt }
func (ids IDs) NewEventID() domain.EventID           { return ids.Event }
func (ids IDs) NewCommandID() domain.CommandID       { return ids.Command }
func (ids IDs) NewArtifactID() domain.ArtifactID     { return ids.Artifact }
func (ids IDs) NewCredentialID() domain.CredentialID { return ids.Credential }
func (ids IDs) NewWorkflowID() domain.WorkflowID     { return ids.Workflow }
