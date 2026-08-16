package fake_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/shaielc/code-winch/internal/adapters/harness/fake"
)

func TestScenarioFixturesValidateAgainstSchema(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile(filepath.Join(root, "schemas/scenarios/v1/fake-harness-scenario.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	fixtures, err := filepath.Glob(filepath.Join(root, "test/scenarios/*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(fixtures) < 5 {
		t.Fatalf("expected shipped scenarios, got %d", len(fixtures))
	}
	for _, fixture := range fixtures {
		t.Run(filepath.Base(fixture), func(t *testing.T) {
			data, err := os.ReadFile(fixture)
			if err != nil {
				t.Fatal(err)
			}
			var value any
			if err := json.Unmarshal(data, &value); err != nil {
				t.Fatal(err)
			}
			if err := schema.Validate(value); err != nil {
				t.Fatal(err)
			}
			if _, err := fake.LoadScenario(fixture); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestScenarioRejectsUnknownStepKind(t *testing.T) {
	s := fake.Scenario{Version: 1, Steps: []fake.ScenarioStep{{Kind: "teleport"}}}
	if err := s.Validate(); err == nil {
		t.Fatal("expected unknown step kind to be rejected")
	}
}
