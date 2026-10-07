package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "index.html"), []byte("ok"))
	t.Setenv("WINCH_SANDBOX_STATIC_DIR", dir)
	t.Setenv("WINCH_SANDBOX_ADDR", "127.0.0.1:0")
	t.Setenv("WINCH_SANDBOX_PROFILE", "container-standard")
	config, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if config.staticDir != dir || config.addr != "127.0.0.1:0" {
		t.Fatalf("unexpected config: %#v", config)
	}
}
func TestLoadConfigEmptyOverridesUseDefaults(t *testing.T) {
	t.Setenv("WINCH_SANDBOX_STATIC_DIR", "")
	t.Setenv("WINCH_SANDBOX_ADDR", "")
	t.Setenv("WINCH_SANDBOX_PROFILE", "")
	config, err := loadConfig()
	if err != nil && err.Error() != "static page unavailable" {
		t.Fatal(err)
	}
	if err == nil && config.staticDir != "/opt/winch/web" {
		t.Fatalf("static dir = %q", config.staticDir)
	}
	if got := valueOrDefault("WINCH_SANDBOX_STATIC_DIR", "/opt/winch/web"); got != "/opt/winch/web" {
		t.Fatalf("empty override resolved to %q", got)
	}
	if got := valueOrDefault("WINCH_SANDBOX_ADDR", "127.0.0.1:8080"); got != "127.0.0.1:8080" {
		t.Fatalf("empty override resolved to %q", got)
	}
}
func TestLoadConfigRejectsInvalidValues(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "index.html"), []byte("ok"))
	t.Setenv("WINCH_SANDBOX_STATIC_DIR", dir)
	t.Setenv("WINCH_SANDBOX_PROFILE", "unknown")
	if _, err := loadConfig(); err == nil {
		t.Fatal("accepted unknown profile")
	}
	t.Setenv("WINCH_SANDBOX_PROFILE", "container-standard")
	t.Setenv("WINCH_SANDBOX_ADDR", "bad address")
	if _, err := loadConfig(); err == nil {
		t.Fatal("accepted invalid address")
	}
	t.Setenv("WINCH_SANDBOX_ADDR", "127.0.0.1:0")
	t.Setenv("WINCH_SANDBOX_STATIC_DIR", t.TempDir())
	if _, err := loadConfig(); err == nil {
		t.Fatal("accepted missing index")
	}
}

func writeTestFile(t *testing.T, path string, contents []byte) {
	t.Helper()
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatal(err)
	}
}
