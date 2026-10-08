package attach_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/shaielc/code-winch/internal/runner"
)

func TestRawStreamRecordGolden(t *testing.T) {
	record := runner.Record{Ordinal: 1, Kind: "stream.raw", OccurredAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC), Sensitivity: "user-content", Payload: runner.StreamPayload{Stream: "stdout", Encoding: "utf-8", Data: "hello\n"}}
	got, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("testdata/stream-raw.json")
	if err != nil {
		t.Fatal(err)
	}
	if string(append(got, '\n')) != string(want) {
		t.Fatalf("got %s, want %s", got, want)
	}
}
