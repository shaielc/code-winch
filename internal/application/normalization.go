package application

import (
	"encoding/json"

	"github.com/shaielc/code-winch/internal/domain"
	"github.com/shaielc/code-winch/pkg/protocol"
)

// NormalizeEvent maps adapter output into the canonical envelope and applies
// provider-neutral validation before an EventStore persists it.
func NormalizeEvent(runID domain.RunID, sequence uint64, value UnsequencedEvent, policy protocol.NormalizationPolicy) protocol.NormalizationResult {
	extensions := make(map[string]json.RawMessage, len(value.Extensions))
	for name, extension := range value.Extensions {
		extensions[name] = json.RawMessage(extension)
	}
	return protocol.NormalizeEvent(protocol.Event{
		EventID: value.EventID.String(), RunID: runID.String(), Sequence: sequence,
		OccurredAt: value.OccurredAt.Time(), Kind: value.Kind, SchemaVersion: value.SchemaVersion,
		Source: value.Source, Sensitivity: value.Sensitivity, Payload: json.RawMessage(value.Payload), Extensions: extensions,
	}, policy)
}
