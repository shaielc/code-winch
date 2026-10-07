package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/shaielc/code-winch/internal/adapters/transport/attach"
)

type config struct {
	addr      string
	staticDir string
	posture   attach.Posture
}

func loadConfig() (config, error) {
	c := config{addr: valueOrDefault("WINCH_SANDBOX_ADDR", "127.0.0.1:8080"), staticDir: valueOrDefault("WINCH_SANDBOX_STATIC_DIR", "/opt/winch/web"), posture: attach.DefaultPosture()}
	if profile := valueOrDefault("WINCH_SANDBOX_PROFILE", "container-standard"); profile != "container-standard" {
		return config{}, fmt.Errorf("unknown sandbox profile")
	}
	if _, err := net.ResolveTCPAddr("tcp", c.addr); err != nil {
		return config{}, fmt.Errorf("invalid sandbox address")
	}
	if info, err := os.Stat(filepath.Join(c.staticDir, "index.html")); err != nil || info.IsDir() {
		return config{}, fmt.Errorf("static page unavailable")
	}
	return c, nil
}

func valueOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
