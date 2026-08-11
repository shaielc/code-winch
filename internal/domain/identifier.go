// Package domain contains dependency-free business value types.
package domain

import (
	"encoding/hex"
	"errors"
)

var errInvalidID = errors.New("invalid identifier: expected a canonical, non-zero UUID")

type identifier [16]byte

func parseIdentifier(value string) (identifier, error) {
	var id identifier
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' {
		return id, errInvalidID
	}

	encoded := value[0:8] + value[9:13] + value[14:18] + value[19:23] + value[24:36]
	if _, err := hex.Decode(id[:], []byte(encoded)); err != nil || id == (identifier{}) {
		return identifier{}, errInvalidID
	}
	// Requiring the formatter's output rejects uppercase and other non-canonical forms.
	if formatIdentifier(id) != value {
		return identifier{}, errInvalidID
	}
	return id, nil
}

func formatIdentifier(id identifier) string {
	var value [36]byte
	hex.Encode(value[0:8], id[0:4])
	value[8] = '-'
	hex.Encode(value[9:13], id[4:6])
	value[13] = '-'
	hex.Encode(value[14:18], id[6:8])
	value[18] = '-'
	hex.Encode(value[19:23], id[8:10])
	value[23] = '-'
	hex.Encode(value[24:36], id[10:16])
	return string(value[:])
}

// Each identifier is a distinct Go type so IDs for different resources cannot
// be interchanged accidentally. External values use canonical UUID text.
//
// The repeated declarations are intentional: aliases or a shared exported ID
// type would weaken the compile-time boundary.
//
//go:generate echo "identifier declarations are maintained in this file"
type WorkspaceID identifier
type RunID identifier
type AttemptID identifier
type EventID identifier
type CommandID identifier
type ArtifactID identifier
type CredentialID identifier
type WorkflowID identifier

func ParseWorkspaceID(value string) (WorkspaceID, error) {
	id, err := parseIdentifier(value)
	return WorkspaceID(id), err
}
func ParseRunID(value string) (RunID, error) {
	id, err := parseIdentifier(value)
	return RunID(id), err
}
func ParseAttemptID(value string) (AttemptID, error) {
	id, err := parseIdentifier(value)
	return AttemptID(id), err
}
func ParseEventID(value string) (EventID, error) {
	id, err := parseIdentifier(value)
	return EventID(id), err
}
func ParseCommandID(value string) (CommandID, error) {
	id, err := parseIdentifier(value)
	return CommandID(id), err
}
func ParseArtifactID(value string) (ArtifactID, error) {
	id, err := parseIdentifier(value)
	return ArtifactID(id), err
}
func ParseCredentialID(value string) (CredentialID, error) {
	id, err := parseIdentifier(value)
	return CredentialID(id), err
}
func ParseWorkflowID(value string) (WorkflowID, error) {
	id, err := parseIdentifier(value)
	return WorkflowID(id), err
}

func (id WorkspaceID) String() string  { return formatIdentifier(identifier(id)) }
func (id RunID) String() string        { return formatIdentifier(identifier(id)) }
func (id AttemptID) String() string    { return formatIdentifier(identifier(id)) }
func (id EventID) String() string      { return formatIdentifier(identifier(id)) }
func (id CommandID) String() string    { return formatIdentifier(identifier(id)) }
func (id ArtifactID) String() string   { return formatIdentifier(identifier(id)) }
func (id CredentialID) String() string { return formatIdentifier(identifier(id)) }
func (id WorkflowID) String() string   { return formatIdentifier(identifier(id)) }

func (id WorkspaceID) IsZero() bool  { return identifier(id) == (identifier{}) }
func (id RunID) IsZero() bool        { return identifier(id) == (identifier{}) }
func (id AttemptID) IsZero() bool    { return identifier(id) == (identifier{}) }
func (id EventID) IsZero() bool      { return identifier(id) == (identifier{}) }
func (id CommandID) IsZero() bool    { return identifier(id) == (identifier{}) }
func (id ArtifactID) IsZero() bool   { return identifier(id) == (identifier{}) }
func (id CredentialID) IsZero() bool { return identifier(id) == (identifier{}) }
func (id WorkflowID) IsZero() bool   { return identifier(id) == (identifier{}) }
