package workflow

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func fixture(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "schemas", "workflow", "v1", "fixtures", "compat-v1-minimal.json"))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestSchemaCompatibility(t *testing.T) {
	b := fixture(t)
	var doc any
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	schemaBytes, err := os.ReadFile(filepath.Join("..", "..", "schemas", "workflow", "v1", "workflow-definition.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	var schemaDoc any
	if err = json.Unmarshal(schemaBytes, &schemaDoc); err != nil {
		t.Fatal(err)
	}
	c := jsonschema.NewCompiler()
	c.AssertFormat()
	const loc = "https://code-winch.dev/schemas/workflow/v1/workflow-definition.schema.json"
	if err = c.AddResource(loc, schemaDoc); err != nil {
		t.Fatal(err)
	}
	s, err := c.Compile(loc)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Validate(doc); err != nil {
		t.Fatal(err)
	}
	d, err := ParseDefinition(b)
	if err != nil {
		t.Fatal(err)
	}
	round, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ParseDefinition(round); err != nil {
		t.Fatal(err)
	}
}

func TestGraphAndTypeValidation(t *testing.T) {
	lit := func(typ ValueType, value string) Binding { return Binding{Type: typ, Literal: json.RawMessage(value)} }
	ref := func(typ ValueType, id, out string) Binding {
		return Binding{Type: typ, Ref: &Reference{StepID: id, Output: out}}
	}
	base := func() Definition {
		return Definition{SchemaVersion: 1, DefinitionID: "flow", Version: "1", Steps: []Step{{ID: "start", Type: "run.start", Inputs: map[string]Binding{"prompt": lit(TypeString, `"hi"`)}}}}
	}
	tests := []struct {
		name, code string
		change     func(*Definition)
	}{
		{"valid", "", func(d *Definition) {}},
		{"cycle", "WF_CYCLE", func(d *Definition) {
			d.Steps = append(d.Steps, Step{ID: "send", Type: "run.send", DependsOn: []string{"start"}, Inputs: map[string]Binding{"run": ref(TypeRun, "start", "run"), "message": lit(TypeString, `"x"`)}})
			d.Steps[0].DependsOn = []string{"send"}
		}},
		{"missing reference", "WF_UNKNOWN_STEP", func(d *Definition) { d.Steps[0].DependsOn = []string{"missing"} }},
		{"wrong output type", "WF_TYPE_MISMATCH", func(d *Definition) {
			d.Steps = append(d.Steps, Step{ID: "send", Type: "run.send", Inputs: map[string]Binding{"run": ref(TypeString, "start", "run"), "message": lit(TypeString, `"x"`)}})
		}},
		{"parallel bound", "WF_CONCURRENCY_BOUND_REQUIRED", func(d *Definition) { d.Steps = append(d.Steps, Step{ID: "fanout", Type: "parallel"}) }},
		{"foreach bound", "WF_CONCURRENCY_BOUND_REQUIRED", func(d *Definition) {
			d.Steps = append(d.Steps, Step{ID: "each", Type: "foreach", Inputs: map[string]Binding{"items": lit(TypeArray, `[]`)}})
		}},
		{"timeout", "WF_INVALID_TIMEOUT", func(d *Definition) { d.Steps[0].Timeout = "forever" }},
		{"compensation", "WF_UNKNOWN_STEP", func(d *Definition) { d.Steps[0].CompensateWith = "missing" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := base()
			tt.change(&d)
			err := d.Validate()
			if tt.code == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), "code="+tt.code) || !strings.Contains(err.Error(), "path=$") {
				t.Fatalf("got %v, want safe path diagnostic %s", err, tt.code)
			}
		})
	}
}

func TestParseRejectsCodeAndLimits(t *testing.T) {
	if _, err := ParseDefinition([]byte(`{"schemaVersion":1,"definitionId":"x","version":"1","steps":[],"code":"os.Exit(1)"}`)); err == nil {
		t.Fatal("accepted executable/unknown field")
	}
	if _, err := ParseDefinition(make([]byte, MaxDefinitionBytes+1)); err == nil || !strings.Contains(err.Error(), "WF_LIMIT_BYTES") {
		t.Fatalf("got %v", err)
	}
}

func TestDiagnosticExcludesContent(t *testing.T) {
	secret := "never-log-me"
	raw := `{"schemaVersion":1,"definitionId":"flow","version":"1","steps":[{"id":"start","type":"run.start","inputs":{"prompt":{"type":"boolean","literal":"` + secret + `"}}}]}`
	_, err := ParseDefinition([]byte(raw))
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatalf("leaked content: %v", err)
	}
}

func TestInstancePins(t *testing.T) {
	if err := (InstancePins{DefinitionID: "flow", DefinitionVersion: "1", HarnessProfile: "codex", SandboxProfile: "locked"}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (InstancePins{DefinitionID: "flow", DefinitionVersion: "1"}).Validate(); err == nil {
		t.Fatal("accepted unpinned profiles")
	}
}

func FuzzParseDefinitionLimits(f *testing.F) {
	f.Add([]byte(`{"schemaVersion":1,"definitionId":"flow","version":"1","steps":[{"id":"s","type":"parallel","maxConcurrency":1}]}`))
	f.Add([]byte(`{}`))
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > MaxDefinitionBytes+64 {
			b = b[:MaxDefinitionBytes+64]
		}
		_, _ = ParseDefinition(b)
	})
}
