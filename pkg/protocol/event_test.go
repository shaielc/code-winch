package protocol

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const schemaLocation = "https://code-winch.dev/schemas/events/v1/canonical-event.schema.json"

func eventSchema(t *testing.T) *jsonschema.Schema {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "schemas", "events", "v1", "canonical-event.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	var document any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatalf("schema is not JSON: %v", err)
	}
	compiler := jsonschema.NewCompiler()
	compiler.AssertFormat()
	if err := compiler.AddResource(schemaLocation, document); err != nil {
		t.Fatalf("add schema: %v", err)
	}
	schema, err := compiler.Compile(schemaLocation)
	if err != nil {
		t.Fatalf("compile schema: %v", err)
	}
	return schema
}

func fixtureFiles(t *testing.T) []string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join("..", "..", "schemas", "events", "v1", "fixtures", "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no event fixtures found")
	}
	return files
}

func TestSchemaAndFixtures(t *testing.T) {
	schema := eventSchema(t)
	for _, file := range fixtureFiles(t) {
		t.Run(filepath.Base(file), func(t *testing.T) {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			var document any
			if err := json.Unmarshal(data, &document); err != nil {
				t.Fatalf("fixture is not JSON: %v", err)
			}
			if err := schema.Validate(document); err != nil {
				t.Fatalf("fixture does not validate: %v", err)
			}
			var event Event
			if err := json.Unmarshal(data, &event); err != nil {
				t.Fatalf("decode fixture: %v", err)
			}
			if err := event.ValidateEnvelope(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCompatibilityRoundTrip(t *testing.T) {
	for _, name := range []string{"compat-v1-minimal.json", "compat-unknown.json"} {
		t.Run(name, func(t *testing.T) {
			original, err := os.ReadFile(filepath.Join("..", "..", "schemas", "events", "v1", "fixtures", name))
			if err != nil {
				t.Fatal(err)
			}
			var event Event
			if err := json.Unmarshal(original, &event); err != nil {
				t.Fatal(err)
			}
			roundTripped, err := json.Marshal(event)
			if err != nil {
				t.Fatal(err)
			}
			var before, after any
			if err := json.Unmarshal(original, &before); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(roundTripped, &after); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(before, after) {
				t.Fatalf("round trip changed event\nbefore: %s\nafter: %s", original, roundTripped)
			}
		})
	}
}

func TestConsumersOrderOnlyBySequence(t *testing.T) {
	events := []Event{{Sequence: 3, EventID: "a"}, {Sequence: 1, EventID: "z"}, {Sequence: 2, EventID: "m"}}
	SortBySequence(events)
	for index, want := range []uint64{1, 2, 3} {
		if events[index].Sequence != want {
			t.Fatalf("event %d has sequence %d, want %d", index, events[index].Sequence, want)
		}
	}
}

func TestValidationErrorExcludesContent(t *testing.T) {
	secret := "do-not-log-this"
	event := Event{EventID: "evt-safe", RunID: "run-safe", Payload: json.RawMessage(`{"value":"` + secret + `"}`)}
	err := event.ValidateEnvelope()
	if err == nil {
		t.Fatal("expected validation failure")
	}
	message := err.Error()
	if strings.Contains(message, secret) {
		t.Fatalf("validation error leaked payload: %s", message)
	}
	for _, safe := range []string{"EVENT_INVALID_SEQUENCE", "evt-safe", "run-safe", "sequence"} {
		if !strings.Contains(message, safe) {
			t.Fatalf("validation error %q does not contain %q", message, safe)
		}
	}
}
