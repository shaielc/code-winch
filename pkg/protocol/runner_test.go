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

func runnerSchema(t *testing.T) *jsonschema.Schema {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "schemas", "runner", "v1", "runner-message.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	var document any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("https://code-winch.dev/schemas/runner/v1/runner-message.schema.json", document); err != nil {
		t.Fatal(err)
	}
	schema, err := compiler.Compile("https://code-winch.dev/schemas/runner/v1/runner-message.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	return schema
}

func fixtureMessages(t *testing.T) [][]byte {
	t.Helper()
	commands, err := os.ReadFile(filepath.Join("..", "..", "schemas", "runner", "v1", "fixtures", "commands.json"))
	if err != nil {
		t.Fatal(err)
	}
	var raw []json.RawMessage
	if err := json.Unmarshal(commands, &raw); err != nil {
		t.Fatal(err)
	}
	observation, err := os.ReadFile(filepath.Join("..", "..", "schemas", "runner", "v1", "fixtures", "observation.json"))
	if err != nil {
		t.Fatal(err)
	}
	result := make([][]byte, 0, len(raw)+1)
	for _, message := range raw {
		result = append(result, message)
	}
	return append(result, observation)
}

func TestRunnerProtocolFixturesRoundTrip(t *testing.T) {
	schema := runnerSchema(t)
	for index, original := range fixtureMessages(t) {
		var document any
		if err := json.Unmarshal(original, &document); err != nil {
			t.Fatal(err)
		}
		if err := schema.Validate(document); err != nil {
			t.Fatalf("fixture %d: %v", index, err)
		}
		message, err := DecodeRunnerMessage(original, DefaultMaxRunnerMessageBytes)
		if err != nil {
			t.Fatalf("fixture %d: %v", index, err)
		}
		encoded, err := json.Marshal(message)
		if err != nil {
			t.Fatal(err)
		}
		var after any
		if err := json.Unmarshal(encoded, &after); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(document, after) {
			t.Fatalf("fixture %d changed during round trip\nbefore: %s\nafter: %s", index, original, encoded)
		}
	}
}

func TestRunnerProtocolNegotiation(t *testing.T) {
	local := RunnerProtocolOffer{Major: 1, MinMinor: 0, MaxMinor: 3, Limits: RunnerLimits{2_000, 10_000, 20}}
	remote := RunnerProtocolOffer{Major: 1, MinMinor: 1, MaxMinor: 2, Limits: RunnerLimits{1_000, 8_000, 10}}
	version, limits, err := NegotiateRunnerProtocol(local, remote)
	if err != nil {
		t.Fatal(err)
	}
	if version.Major != 1 || version.Minor != 2 {
		t.Fatalf("version = %+v", version)
	}
	if limits != (RunnerLimits{1_000, 8_000, 10}) {
		t.Fatalf("limits = %+v", limits)
	}

	remote.Major = 2
	_, _, err = NegotiateRunnerProtocol(local, remote)
	if err == nil || !strings.Contains(err.Error(), "RUNNER_MAJOR_VERSION_MISMATCH") {
		t.Fatalf("error = %v", err)
	}
}

func TestRunnerMessageBoundaryAndSafeError(t *testing.T) {
	content := `{"version":{"major":1,"minor":0},"kind":"input","commandId":"cmd-safe","executionId":"exec-safe","leaseToken":"do-not-log","payload":{"text":"secret-content"}}`
	_, err := DecodeRunnerMessage([]byte(content), uint64(len(content)-1))
	if err == nil || !strings.Contains(err.Error(), "RUNNER_MESSAGE_TOO_LARGE") {
		t.Fatalf("error = %v", err)
	}
	if strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), "do-not-log") {
		t.Fatalf("error leaked content: %v", err)
	}
	message, err := DecodeRunnerMessage([]byte(content), uint64(len(content)))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := EncodeRunnerMessage(message, uint64(len(content)-1)); err == nil || !strings.Contains(err.Error(), "RUNNER_MESSAGE_TOO_LARGE") {
		t.Fatalf("encode error = %v", err)
	}
}

func TestRunnerProtocolRejectsInvalidLimitsAndMinorRange(t *testing.T) {
	valid := RunnerLimits{100, 200, 1}
	for _, remote := range []RunnerProtocolOffer{
		{Major: 1, MinMinor: 2, MaxMinor: 3, Limits: valid},
		{Major: 1, MinMinor: 0, MaxMinor: 0, Limits: RunnerLimits{100, 50, 1}},
	} {
		_, _, err := NegotiateRunnerProtocol(RunnerProtocolOffer{Major: 1, MinMinor: 0, MaxMinor: 1, Limits: valid}, remote)
		if err == nil {
			t.Fatal("expected negotiation failure")
		}
	}
}
