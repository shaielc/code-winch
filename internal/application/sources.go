// Package application defines use-case ports that outer layers implement.
package application

import "github.com/shaielc/code-winch/internal/domain"

type Clock interface{ Now() domain.Timestamp }

// IDSource groups creation behind one injectable port while retaining a
// distinct return type for every domain resource.
type IDSource interface {
	NewWorkspaceID() domain.WorkspaceID
	NewRunID() domain.RunID
	NewAttemptID() domain.AttemptID
	NewEventID() domain.EventID
	NewCommandID() domain.CommandID
	NewArtifactID() domain.ArtifactID
	NewCredentialID() domain.CredentialID
	NewWorkflowID() domain.WorkflowID
}
