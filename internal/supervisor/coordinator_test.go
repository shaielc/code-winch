package supervisor

import (
	"context"
	"testing"

	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/pkg/protocol"
)

func TestFakeProfileRedactorRejectsSecretsAndDropsExtensions(t *testing.T) {
	redactor := fakeProfileRedactor{}
	_, err := redactor.Redact(context.Background(), application.UnsequencedEvent{Sensitivity: protocol.SensitivitySecret, Payload: []byte(`{}`)})
	if err == nil {
		t.Fatal("secret event was accepted")
	}
	event, err := redactor.Redact(context.Background(), application.UnsequencedEvent{Sensitivity: protocol.SensitivityPublic, Payload: []byte(`{"safe":true}`), Extensions: map[string][]byte{"unreviewed": []byte(`true`)}})
	if err != nil {
		t.Fatal(err)
	}
	if event.Extensions != nil {
		t.Fatalf("extensions were persisted: %#v", event.Extensions)
	}
}
