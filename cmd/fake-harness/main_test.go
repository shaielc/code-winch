package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/shaielc/code-winch/internal/adapters/harness/fake"
)

func randomScenario() fake.Scenario {
	payload, _ := json.Marshal(map[string]string{"value": "{{random}}"})
	return fake.Scenario{Version: 1, Steps: []fake.ScenarioStep{{Kind: "emit", EventKind: "diagnostic", Payload: payload}}}
}

func TestSeedReproducibilityAndOffVariation(t *testing.T) {
	run := func(seed string) string {
		rng, err := seeded(seed)
		if err != nil {
			t.Fatal(err)
		}
		var out bytes.Buffer
		runScenario(randomScenario(), bytes.NewBuffer(nil), &out, options{rng: rng})
		return out.String()
	}
	if a, b := run("7"), run("7"); a != b {
		t.Fatalf("same seed differed: %q != %q", a, b)
	}
	if a, b := run("off"), run("off"); a == b {
		t.Fatalf("seeding off unexpectedly matched: %q", a)
	}
}

func TestScenarioFaultsAreFramed(t *testing.T) {
	rng, _ := seeded("1")
	var out bytes.Buffer
	s := fake.Scenario{Version: 1, Steps: []fake.ScenarioStep{{Kind: "malformed", Data: "{broken"}}}
	runScenario(s, bytes.NewBuffer(nil), &out, options{rng: rng, truncateRecord: 1, oversizedBytes: 32})
	if got := out.Bytes(); len(got) == 0 || got[len(got)-1] != '\n' {
		t.Fatalf("fault was not newline framed: %q", got)
	}
}
