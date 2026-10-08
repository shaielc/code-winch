package e2e

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/shaielc/code-winch/internal/runner"
)

func TestHarnessOutputArrivesInOrder(t *testing.T) {
	endpoint := strings.Replace(sandboxURL(t), "http://", "ws://", 1) + "/api/session/stream"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	connection, _, err := websocket.Dial(ctx, endpoint, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = connection.CloseNow() }()
	var data strings.Builder
	for ordinal := uint64(1); ; ordinal++ {
		_, message, err := connection.Read(ctx)
		if err != nil {
			t.Fatal(err)
		}
		var record runner.Record
		if err := json.Unmarshal(message, &record); err != nil {
			t.Fatal(err)
		}
		if record.Ordinal != ordinal {
			t.Fatalf("ordinal = %d, want %d", record.Ordinal, ordinal)
		}
		if record.Kind == "stream.raw" {
			var payload runner.StreamPayload
			encoded, _ := json.Marshal(record.Payload)
			if err := json.Unmarshal(encoded, &payload); err != nil {
				t.Fatal(err)
			}
			data.WriteString(payload.Data)
		}
		if strings.Contains(data.String(), "fake harness ready") && strings.Contains(data.String(), "fake-harness accepts") {
			break
		}
	}
}
