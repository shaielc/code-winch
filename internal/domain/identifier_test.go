package domain

import "testing"

const validID = "123e4567-e89b-12d3-a456-426614174000"

func TestIdentifierTypesParseCanonicalUUID(t *testing.T) {
	tests := map[string]struct {
		parse func(string) (interface{ String() string }, error)
	}{
		"workspace":  func(value string) (interface{ String() string }, error) { return ParseWorkspaceID(value) },
		"run":        func(value string) (interface{ String() string }, error) { return ParseRunID(value) },
		"attempt":    func(value string) (interface{ String() string }, error) { return ParseAttemptID(value) },
		"event":      func(value string) (interface{ String() string }, error) { return ParseEventID(value) },
		"command":    func(value string) (interface{ String() string }, error) { return ParseCommandID(value) },
		"artifact":   func(value string) (interface{ String() string }, error) { return ParseArtifactID(value) },
		"credential": func(value string) (interface{ String() string }, error) { return ParseCredentialID(value) },
		"workflow":   func(value string) (interface{ String() string }, error) { return ParseWorkflowID(value) },
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			id, err := test.parse(validID)
			if err != nil {
				t.Fatalf("parse valid ID: %v", err)
			}
			if got := id.String(); got != validID {
				t.Fatalf("String() = %q, want %q", got, validID)
			}
		})
	}
}

func TestIdentifierParsingRejectsInvalidExternalValues(t *testing.T) {
	for _, value := range []string{
		"", "not-an-id", "00000000-0000-0000-0000-000000000000",
		"123E4567-E89B-12D3-A456-426614174000", "123e4567e89b12d3a456426614174000",
	} {
		t.Run(value, func(t *testing.T) {
			if _, err := ParseRunID(value); err == nil || err.Error() != errInvalidID.Error() {
				t.Fatalf("ParseRunID(%q) error = %v, want %q", value, err, errInvalidID)
			}
		})
	}
}

func TestZeroIdentifierIsDetectable(t *testing.T) {
	if !(RunID{}).IsZero() {
		t.Fatal("zero RunID was accepted")
	}
}
