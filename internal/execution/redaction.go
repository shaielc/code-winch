package execution

import (
	"context"

	"github.com/shaielc/code-winch/internal/application"
)

// ClassificationRedactor is the phase-1 redactor. It reads an event's declared
// sensitivity and rewrites nothing.
//
// What it does not do: it does not scan payloads for credentials, tokens, or
// keys, so an event a harness misclassifies is persisted as it arrived. That is
// tolerable only while the sole harness is the fake one, whose output is the
// operator's own input echoed back on a local-trusted deployment. Scanning for
// known credential values belongs to P3-032, which layers it on this step.
//
// Secret-classified events never reach storage regardless: the supervisor
// refuses them after redaction and the store refuses them again on write.
type ClassificationRedactor struct{}

func (ClassificationRedactor) Redact(_ context.Context, event application.UnsequencedEvent) (application.UnsequencedEvent, error) {
	return event, nil
}

var _ application.EventRedactor = ClassificationRedactor{}
