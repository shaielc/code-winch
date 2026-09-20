package application

import (
	"context"

	"github.com/shaielc/code-winch/pkg/protocol"
)

// SensitivityRedactor classifies an observation before it is persisted. It
// applies the envelope rule from docs/contracts.md §2: a producer that declares
// no sensitivity, or declares one this build does not know, is treated as
// `confidential` rather than as the safest-to-share class.
//
// What it does not do: it inspects no payload and matches no secret patterns.
// It is sound for the profiles this daemon ships because nothing injects a
// credential into a harness yet — `ResolvedCredentials` is always empty on the
// launch path — so the only secret material a run could carry is material an
// adapter declared. An adapter that declares `secret` is refused outright by
// the supervisor rather than rewritten here, because silently dropping a field
// an adapter marked secret would hide the adapter's own mistake.
type SensitivityRedactor struct{}

func (SensitivityRedactor) Redact(_ context.Context, event UnsequencedEvent) (UnsequencedEvent, error) {
	switch event.Sensitivity {
	case protocol.SensitivityPublic, protocol.SensitivityOperational,
		protocol.SensitivityUserContent, protocol.SensitivityConfidential,
		protocol.SensitivitySecret:
	default:
		event.Sensitivity = protocol.SensitivityConfidential
	}
	return event, nil
}

var _ EventRedactor = SensitivityRedactor{}
