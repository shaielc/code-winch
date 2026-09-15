package config

import (
	"os"
	"strings"
	"testing"
)

func TestLoadPrecedence(t *testing.T) {
	p := t.TempDir() + "/c.yaml"
	data := []byte("addr: ':7000'\ntoken: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\ncsrf_token: bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb\n")
	if err := os.WriteFile(p, data, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WINCH_CONFIG_FILE", p)
	t.Setenv("WINCH_ADDR", ":9000")
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.Addr != ":9000" {
		t.Fatalf("addr=%s", c.Addr)
	}
}

func TestValidationReportsAllFieldsWithoutValues(t *testing.T) {
	c := Defaults()
	c.Token = "canary"
	c.CSRFToken = "x"
	c.DatabaseURL = "bad"
	err := c.Validate()
	if err == nil {
		t.Fatal("expected error")
	}
	message := err.Error()
	for _, field := range []string{"token", "csrf_token", "database_url"} {
		if !strings.Contains(message, field) {
			t.Errorf("missing %s", field)
		}
	}
	if strings.Contains(message, "canary") {
		t.Fatal("secret leaked")
	}
}

func TestFakeHarnessProfileResolvesFromTheEnvironment(t *testing.T) {
	t.Setenv("WINCH_TOKEN", strings.Repeat("a", 32))
	t.Setenv("WINCH_CSRF_TOKEN", strings.Repeat("b", 32))
	t.Setenv("WINCH_FAKE_HARNESS_BINARY", "/opt/winch/fake-harness")
	t.Setenv("WINCH_FAKE_HARNESS_TRANSCRIPT", "/etc/winch/transcript.txt")
	t.Setenv("WINCH_FAKE_HARNESS_DELAY", "250ms")
	t.Setenv("WINCH_FAKE_HARNESS_FORCE_FAILURE", "true")
	t.Setenv("WINCH_FAKE_HARNESS_MALFORMED_LINE", "1")
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.FakeHarness.Binary != "/opt/winch/fake-harness" || c.FakeHarness.Transcript != "/etc/winch/transcript.txt" {
		t.Fatalf("paths: %#v", c.FakeHarness)
	}
	if c.FakeHarness.Delay.String() != "250ms" || !c.FakeHarness.ForceFailure || !c.FakeHarness.MalformedLine {
		t.Fatalf("injections: %#v", c.FakeHarness)
	}
}

func TestFakeHarnessProfileDefaultsToNoInjections(t *testing.T) {
	t.Setenv("WINCH_TOKEN", strings.Repeat("a", 32))
	t.Setenv("WINCH_CSRF_TOKEN", strings.Repeat("b", 32))
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.FakeHarness != (FakeHarnessConfig{}) {
		t.Fatalf("unconfigured fake profile: %#v", c.FakeHarness)
	}
}

// An unparseable control is named rather than silently resolving to a profile
// the operator did not ask for.
func TestFakeHarnessProfileRejectsUnparseableControls(t *testing.T) {
	t.Setenv("WINCH_TOKEN", strings.Repeat("a", 32))
	t.Setenv("WINCH_CSRF_TOKEN", strings.Repeat("b", 32))
	t.Setenv("WINCH_FAKE_HARNESS_FORCE_FAILURE", "sometimes")
	t.Setenv("WINCH_FAKE_HARNESS_DELAY", "a while")
	_, err := Load()
	if err == nil {
		t.Fatal("expected a validation error")
	}
	for _, field := range []string{"fake_harness.force_failure", "fake_harness.delay"} {
		if !strings.Contains(err.Error(), field) {
			t.Errorf("missing %s in %q", field, err.Error())
		}
	}
	if strings.Contains(err.Error(), "sometimes") {
		t.Fatalf("rejected value leaked: %s", err.Error())
	}
}
