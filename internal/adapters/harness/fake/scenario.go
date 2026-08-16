package fake

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"time"
)

// Scenario is the data-driven behavior understood by the fake-harness binary.
type Scenario struct {
	Version         int            `json:"version"`
	IgnoreSIGTERMMS int            `json:"ignore_sigterm_ms,omitempty"`
	Steps           []ScenarioStep `json:"steps"`
}

// ScenarioStep is one action in a scenario. Exactly the fields belonging to
// Kind are accepted by Validate, keeping scenario mistakes from failing later
// and less visibly in a run.
type ScenarioStep struct {
	Kind       string          `json:"kind"`
	EventKind  string          `json:"event_kind,omitempty"`
	Payload    json.RawMessage `json:"payload,omitempty"`
	Pattern    string          `json:"pattern,omitempty"`
	DurationMS int             `json:"duration_ms,omitempty"`
	Code       int             `json:"code,omitempty"`
	Data       string          `json:"data,omitempty"`
}

func LoadScenario(path string) (Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Scenario{}, fmt.Errorf("read scenario: %w", err)
	}
	var scenario Scenario
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&scenario); err != nil {
		return Scenario{}, fmt.Errorf("decode scenario: %w", err)
	}
	if err := scenario.Validate(); err != nil {
		return Scenario{}, err
	}
	return scenario, nil
}

func (s Scenario) Validate() error {
	if s.Version != 1 || len(s.Steps) == 0 || s.IgnoreSIGTERMMS < 0 {
		return fmt.Errorf("invalid scenario: version must be 1 and steps must not be empty")
	}
	for i, step := range s.Steps {
		bad := func(message string) error { return fmt.Errorf("invalid scenario step %d: %s", i+1, message) }
		switch step.Kind {
		case "emit":
			if step.EventKind == "" || len(step.Payload) == 0 || !json.Valid(step.Payload) {
				return bad("emit requires event_kind and JSON payload")
			}
		case "wait_input":
			if step.Pattern == "" {
				return bad("wait_input requires pattern")
			}
			if _, err := regexp.Compile(step.Pattern); err != nil {
				return bad("wait_input pattern is not a valid regular expression")
			}
		case "sleep":
			if step.DurationMS < 0 {
				return bad("sleep duration_ms must be non-negative")
			}
		case "exit":
			if step.Code < 0 || step.Code > 255 {
				return bad("exit code must be between 0 and 255")
			}
		case "malformed":
			if step.Data == "" || bytes.ContainsRune([]byte(step.Data), '\n') {
				return bad("malformed requires single-line data")
			}
		default:
			return bad("unknown kind " + step.Kind)
		}
	}
	return nil
}

func (s ScenarioStep) Duration() time.Duration { return time.Duration(s.DurationMS) * time.Millisecond }
