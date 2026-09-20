package application_test

import (
	"context"
	"testing"

	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/pkg/protocol"
)

func TestRedactorTreatsMissingAndUnknownSensitivityAsConfidential(t *testing.T) {
	for name, declared := range map[string]protocol.Sensitivity{"missing": "", "unknown": "extremely-fine"} {
		t.Run(name, func(t *testing.T) {
			got, err := application.SensitivityRedactor{}.Redact(context.Background(), application.UnsequencedEvent{Sensitivity: declared, Payload: []byte(`{}`)})
			if err != nil {
				t.Fatal(err)
			}
			if got.Sensitivity != protocol.SensitivityConfidential {
				t.Fatalf("%q was classified %q", declared, got.Sensitivity)
			}
		})
	}
}

func TestRedactorKeepsEveryDeclaredClass(t *testing.T) {
	for _, declared := range []protocol.Sensitivity{protocol.SensitivityPublic, protocol.SensitivityOperational, protocol.SensitivityUserContent, protocol.SensitivityConfidential, protocol.SensitivitySecret} {
		got, err := application.SensitivityRedactor{}.Redact(context.Background(), application.UnsequencedEvent{Sensitivity: declared, Payload: []byte(`{}`)})
		if err != nil {
			t.Fatal(err)
		}
		// A declared secret is passed through unchanged rather than downgraded:
		// the supervisor refuses it, which surfaces the adapter's mistake instead
		// of quietly persisting the event under a weaker class.
		if got.Sensitivity != declared {
			t.Fatalf("%q was rewritten to %q", declared, got.Sensitivity)
		}
	}
}
