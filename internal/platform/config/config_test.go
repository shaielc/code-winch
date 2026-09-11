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

func TestMemoryProfileDoesNotRequireDatabaseURL(t *testing.T) {
	c := Defaults()
	c.StoreProfile = "memory"
	c.DatabaseURL = ""
	c.Token, c.CSRFToken = strings.Repeat("a", 32), strings.Repeat("b", 32)
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestStoreProfileAndMemoryFailureAreValidated(t *testing.T) {
	c := Defaults()
	c.Token, c.CSRFToken = strings.Repeat("a", 32), strings.Repeat("b", 32)
	c.StoreProfile = "disk"
	c.MemoryInjectRunSave = "secret-canary"
	err := c.Validate()
	if err == nil || !strings.Contains(err.Error(), "store_profile") || !strings.Contains(err.Error(), "memory_inject_run_save") {
		t.Fatalf("validation error=%v", err)
	}
	if strings.Contains(err.Error(), "secret-canary") {
		t.Fatal("rejected configuration value leaked")
	}
}
