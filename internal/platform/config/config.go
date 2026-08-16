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
	env := map[string]*string{"WINCH_ADDR": &c.Addr, "WINCH_DATABASE_URL": &c.DatabaseURL, "WINCH_ALLOWED_ORIGIN": &c.AllowedOrigin, "WINCH_TOKEN": &c.Token, "WINCH_CSRF_TOKEN": &c.CSRFToken, "WINCH_ACTOR": &c.Actor, "WINCH_STATIC_DIR": &c.StaticDir, "WINCH_SHUTDOWN_TIMEOUT": &c.ShutdownTimeoutText}
	for key, dst := range env {
		if value, ok := os.LookupEnv(key); ok {
			*dst = value
		}
	}
	if c.ShutdownTimeoutText != "" {
		d, err := time.ParseDuration(c.ShutdownTimeoutText)
		if err == nil {
			c.ShutdownTimeout = d
		}
	}
	return c, c.Validate()
}

func (c Config) Validate() error {
	var fields []string
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
