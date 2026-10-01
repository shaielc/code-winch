package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestCreateStartInputPollEventsAndReplay(t *testing.T) {
	database := requireDatabase(t)
	base := startDaemon(t, database, "WINCH_FAKE_HARNESS_TRANSCRIPT="+writeTranscript(t, "echo ready"))
	run := createRun(t, base, "input-1", "/tmp/ws", "fake", "local")
	runID := run["id"].(string)
	started := request(t, http.MethodPost, base+"/api/v1/runs/"+runID+"/start", nil, map[string]string{"Idempotency-Key": "start-input-1", "If-Match": etag(run)})
	if started.StatusCode != http.StatusAccepted {
		t.Fatalf("start status=%d body=%s", started.StatusCode, started.Body)
	}
	current := readRun(t, base, runID)
	body, _ := json.Marshal(map[string]string{"kind": "text", "text": "echo input-marker-73"})
	accepted := request(t, http.MethodPost, base+"/api/v1/runs/"+runID+"/input", body, map[string]string{"Idempotency-Key": "input-73", "If-Match": etag(current)})
	if accepted.StatusCode != http.StatusAccepted {
		t.Fatalf("input status=%d body=%s", accepted.StatusCode, accepted.Body)
	}
	var first map[string]any
	if err := json.Unmarshal(accepted.Body, &first); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if containsOutput(pollEvents(t, base, runID), "input-marker-73") {
			break
		}
		time.Sleep(25 * time.Millisecond)
	}
	if !containsOutput(pollEvents(t, base, runID), "input-marker-73") {
		t.Fatal("input response missing from durable events")
	}
	replay := request(t, http.MethodPost, base+"/api/v1/runs/"+runID+"/input", body, map[string]string{"Idempotency-Key": "input-73", "If-Match": `"0"`})
	if replay.StatusCode != http.StatusAccepted || !bytes.Contains(replay.Body, []byte(first["commandId"].(string))) {
		t.Fatalf("replay status=%d body=%s", replay.StatusCode, replay.Body)
	}
	awaitOutboxDrain(t, database, runID)
}
