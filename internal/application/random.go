package application

import (
	"crypto/rand"

	"github.com/shaielc/code-winch/internal/domain"
)

// RandomIDs generates RFC 4122 version-4 identifiers for the composition root.
type RandomIDs struct{}

func randomUUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("secure random identifier source unavailable")
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	const h = "0123456789abcdef"
	out := make([]byte, 36)
	j := 0
	for i, v := range b {
		if i == 4 || i == 6 || i == 8 || i == 10 {
			out[j] = '-'
			j++
		}
		out[j] = h[v>>4]
		out[j+1] = h[v&15]
		j += 2
	}
	return string(out)
}
func (RandomIDs) NewWorkspaceID() domain.WorkspaceID {
	v, _ := domain.ParseWorkspaceID(randomUUID())
	return v
}
func (RandomIDs) NewRunID() domain.RunID { v, _ := domain.ParseRunID(randomUUID()); return v }
func (RandomIDs) NewAttemptID() domain.AttemptID {
	v, _ := domain.ParseAttemptID(randomUUID())
	return v
}
func (RandomIDs) NewEventID() domain.EventID { v, _ := domain.ParseEventID(randomUUID()); return v }
func (RandomIDs) NewCommandID() domain.CommandID {
	v, _ := domain.ParseCommandID(randomUUID())
	return v
}
func (RandomIDs) NewArtifactID() domain.ArtifactID {
	v, _ := domain.ParseArtifactID(randomUUID())
	return v
}
func (RandomIDs) NewCredentialID() domain.CredentialID {
	v, _ := domain.ParseCredentialID(randomUUID())
	return v
}
func (RandomIDs) NewWorkflowID() domain.WorkflowID {
	v, _ := domain.ParseWorkflowID(randomUUID())
	return v
}
