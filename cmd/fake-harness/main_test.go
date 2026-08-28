package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"
)

func TestTranscriptPlaybackSelectedAtRuntime(t *testing.T) {
	path := writeTranscript(t, "echo hello from transcript\nfail\n")
	var output bytes.Buffer
	code := run(config{runID: "run-1", transcript: path}, strings.NewReader(""), &output)
	if code != 1 || !strings.Contains(output.String(), "hello from transcript") || !strings.Contains(output.String(), "exiting with a failure") {
		t.Fatalf("code=%d output=%q", code, output.String())
	}
}

func TestInjectedControls(t *testing.T) {
	tests := []struct {
		name string
		cfg  config
		code int
		text string
	}{
		{name: "forced failure", cfg: config{forceFailure: true}, code: 1, text: "forced failure"},
		{name: "malformed output", cfg: config{malformedLine: true}, code: 1, text: "fake-harness-malformed-record"},
		{name: "early exit", cfg: config{earlyExit: true}, text: "exiting early"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.cfg.runID = "run-1"
			var output bytes.Buffer
			if code := run(test.cfg, strings.NewReader("ignored\n"), &output); code != test.code || !strings.Contains(output.String(), test.text) {
				t.Fatalf("code=%d output=%q", code, output.String())
			}
		})
	}
}

func TestTranscriptDelay(t *testing.T) {
	path := writeTranscript(t, "echo delayed\nexit\n")
	started := time.Now()
	if code := run(config{runID: "run-1", transcript: path, delay: 20 * time.Millisecond}, strings.NewReader(""), &bytes.Buffer{}); code != 0 {
		t.Fatalf("code=%d", code)
	}
	if elapsed := time.Since(started); elapsed < 40*time.Millisecond {
		t.Fatalf("elapsed=%s; want at least 40ms", elapsed)
	}
}

func writeTranscript(t *testing.T, content string) string {
	t.Helper()
	path := t.TempDir() + "/transcript.txt"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
