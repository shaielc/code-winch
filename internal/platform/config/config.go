// Package config resolves and validates daemon startup configuration.
package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Addr                string        `yaml:"addr"`
	DatabaseURL         string        `yaml:"database_url"`
	AllowedOrigin       string        `yaml:"allowed_origin"`
	Token               string        `yaml:"token"`
	CSRFToken           string        `yaml:"csrf_token"`
	Actor               string        `yaml:"actor"`
	StaticDir           string        `yaml:"static_dir"`
	ShutdownTimeout     time.Duration `yaml:"-"`
	ShutdownTimeoutText string        `yaml:"shutdown_timeout"`
	// FakeHarness holds the named-profile controls for `harnessProfile=fake`.
	// They are the daemon-side half of the controllable fake profile: the same
	// transcript, latency, and failure injections `winch dev run` exposes as
	// flags, resolved through the configuration layering rather than compiled in.
	FakeHarness FakeHarnessConfig `yaml:"fake_harness"`
}

// FakeHarnessConfig configures the shipped fake harness profile. With no
// transcript the daemon runs the harness in early-exit mode, because a run
// started through the API has no way to answer an interactive prompt yet and a
// harness blocked on its terminal would never reach a terminal state.
type FakeHarnessConfig struct {
	Binary        string        `yaml:"binary"`
	Transcript    string        `yaml:"transcript"`
	Delay         time.Duration `yaml:"-"`
	DelayText     string        `yaml:"delay"`
	ForceFailure  bool          `yaml:"force_failure"`
	MalformedLine bool          `yaml:"malformed_line"`
}

// ValidationError contains field names only, never rejected values.
type ValidationError struct{ Fields []string }

func (e *ValidationError) Error() string {
	return "invalid configuration fields: " + strings.Join(e.Fields, ", ")
}

func Defaults() Config {
	return Config{Addr: ":8080", DatabaseURL: "postgres://winch:winch-local-development@localhost:5432/winch?sslmode=disable", AllowedOrigin: "http://localhost:8080", Actor: "local-user", StaticDir: "web/dist", ShutdownTimeout: 10 * time.Second}
}

// Load applies safe compiled defaults, an optional YAML file, and environment variables.
func Load() (Config, error) {
	c := Defaults()
	if path := os.Getenv("WINCH_CONFIG_FILE"); path != "" {
		b, err := os.ReadFile(path)
		if err != nil {
			return c, fmt.Errorf("config file: %w", err)
		}
		if err = yaml.Unmarshal(b, &c); err != nil {
			return c, fmt.Errorf("config file: %w", err)
		}
	}
	env := map[string]*string{"WINCH_ADDR": &c.Addr, "WINCH_DATABASE_URL": &c.DatabaseURL, "WINCH_ALLOWED_ORIGIN": &c.AllowedOrigin, "WINCH_TOKEN": &c.Token, "WINCH_CSRF_TOKEN": &c.CSRFToken, "WINCH_ACTOR": &c.Actor, "WINCH_STATIC_DIR": &c.StaticDir, "WINCH_SHUTDOWN_TIMEOUT": &c.ShutdownTimeoutText, "WINCH_FAKE_HARNESS_BINARY": &c.FakeHarness.Binary, "WINCH_FAKE_HARNESS_TRANSCRIPT": &c.FakeHarness.Transcript, "WINCH_FAKE_HARNESS_DELAY": &c.FakeHarness.DelayText}
	for key, dst := range env {
		if value, ok := os.LookupEnv(key); ok {
			*dst = value
		}
	}
	var invalid []string
	flags := map[string]*bool{"fake_harness.force_failure": &c.FakeHarness.ForceFailure, "fake_harness.malformed_line": &c.FakeHarness.MalformedLine}
	for field, dst := range flags {
		// An unparseable flag is reported rather than silently resolving to false:
		// starting a profile the operator did not ask for is the worse answer.
		value, ok := os.LookupEnv("WINCH_FAKE_HARNESS_" + strings.ToUpper(strings.TrimPrefix(field, "fake_harness.")))
		if !ok {
			continue
		}
		if parsed, err := strconv.ParseBool(value); err == nil {
			*dst = parsed
		} else {
			invalid = append(invalid, field)
		}
	}
	if c.ShutdownTimeoutText != "" {
		d, err := time.ParseDuration(c.ShutdownTimeoutText)
		if err == nil {
			c.ShutdownTimeout = d
		}
	}
	if c.FakeHarness.DelayText != "" {
		d, err := time.ParseDuration(c.FakeHarness.DelayText)
		if err == nil {
			c.FakeHarness.Delay = d
		}
	}
	return c, c.validate(invalid)
}

func (c Config) Validate() error { return c.validate(nil) }

// validate reports every invalid field at once. Callers pass fields Load could
// not resolve at all, so a rejected boolean is named alongside the rest.
func (c Config) validate(fields []string) error {
	if strings.TrimSpace(c.Addr) == "" {
		fields = append(fields, "addr")
	}
	if u, e := url.Parse(c.DatabaseURL); e != nil || u.Scheme == "" || u.Host == "" {
		fields = append(fields, "database_url")
	}
	if u, e := url.Parse(c.AllowedOrigin); e != nil || u.Scheme == "" || u.Host == "" || u.Path != "" {
		fields = append(fields, "allowed_origin")
	}
	if len(c.Token) < 32 {
		fields = append(fields, "token")
	}
	if len(c.CSRFToken) < 32 {
		fields = append(fields, "csrf_token")
	}
	if strings.TrimSpace(c.Actor) == "" {
		fields = append(fields, "actor")
	}
	if strings.TrimSpace(c.StaticDir) == "" {
		fields = append(fields, "static_dir")
	}
	if c.ShutdownTimeoutText != "" {
		if _, e := time.ParseDuration(c.ShutdownTimeoutText); e != nil {
			fields = append(fields, "shutdown_timeout")
		}
	}
	if c.ShutdownTimeout <= 0 {
		fields = append(fields, "shutdown_timeout")
	}
	if c.FakeHarness.DelayText != "" {
		if d, e := time.ParseDuration(c.FakeHarness.DelayText); e != nil || d < 0 {
			fields = append(fields, "fake_harness.delay")
		}
	}
	if c.FakeHarness.Delay < 0 {
		fields = append(fields, "fake_harness.delay")
	}
	if len(fields) > 0 {
		return &ValidationError{Fields: fields}
	}
	return nil
}

func ParseInt(value string, fallback int) int {
	n, e := strconv.Atoi(value)
	if e != nil {
		return fallback
	}
	return n
}
