package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/shaielc/code-winch/internal/adapters/transport/attach"
)

type config struct {
	addr                 string
	staticDir            string
	posture              attach.Posture
	harnessProfile       string
	harnessExecutable    string
	harnessTranscript    string
	harnessDelay         time.Duration
	harnessForceFailure  bool
	harnessEarlyExit     bool
	harnessMalformedLine bool
}

func loadConfig() (config, error) {
	c := config{addr: valueOrDefault("WINCH_SANDBOX_ADDR", "127.0.0.1:8080"), staticDir: valueOrDefault("WINCH_SANDBOX_STATIC_DIR", "/opt/winch/web"), posture: attach.DefaultPosture(), harnessProfile: valueOrDefault("WINCH_HARNESS_PROFILE", "fake"), harnessExecutable: "fake-harness"}
	if profile := valueOrDefault("WINCH_SANDBOX_PROFILE", "container-standard"); profile != "container-standard" {
		return config{}, fmt.Errorf("unknown sandbox profile")
	}
	if _, err := net.ResolveTCPAddr("tcp", c.addr); err != nil {
		return config{}, fmt.Errorf("invalid sandbox address")
	}
	if info, err := os.Stat(filepath.Join(c.staticDir, "index.html")); err != nil || info.IsDir() {
		return config{}, fmt.Errorf("static page unavailable")
	}
	if c.harnessProfile != "fake" {
		return config{}, fmt.Errorf("unknown harness profile")
	}
	var err error
	c.harnessTranscript = os.Getenv("WINCH_HARNESS_TRANSCRIPT")
	if value := os.Getenv("WINCH_HARNESS_DELAY"); value != "" {
		if c.harnessDelay, err = time.ParseDuration(value); err != nil || c.harnessDelay < 0 {
			return config{}, fmt.Errorf("invalid harness delay")
		}
	}
	if c.harnessForceFailure, err = boolEnv("WINCH_HARNESS_FORCE_FAILURE"); err != nil {
		return config{}, err
	}
	if c.harnessEarlyExit, err = boolEnv("WINCH_HARNESS_EARLY_EXIT"); err != nil {
		return config{}, err
	}
	if c.harnessMalformedLine, err = boolEnv("WINCH_HARNESS_MALFORMED_LINE"); err != nil {
		return config{}, err
	}
	return c, nil
}

func (c config) harnessArgs() []string {
	args := []string{"-run-id", "sandbox-session"}
	if c.harnessTranscript != "" {
		args = append(args, "-transcript", c.harnessTranscript)
	}
	if c.harnessDelay != 0 {
		args = append(args, "-delay", c.harnessDelay.String())
	}
	if c.harnessForceFailure {
		args = append(args, "-force-failure")
	}
	if c.harnessEarlyExit {
		args = append(args, "-early-exit")
	}
	if c.harnessMalformedLine {
		args = append(args, "-malformed-line")
	}
	return args
}

func boolEnv(key string) (bool, error) {
	value := os.Getenv(key)
	if value == "" {
		return false, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("invalid %s", key)
	}
	return parsed, nil
}

func valueOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
