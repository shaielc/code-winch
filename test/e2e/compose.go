package e2e

import (
	"bytes"
	"context"
	"os/exec"
	"testing"
	"time"
)

func compose(t *testing.T, args ...string) string {
	t.Helper()
	commandArgs := append([]string{"compose", "-f", "deployments/compose.yml"}, args...)
	command := exec.Command("docker", commandArgs...)
	command.Dir = "../.."
	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = &output
	if err := command.Run(); err != nil {
		t.Fatalf("docker %v: %v\n%s", commandArgs, err, output.String())
	}
	return output.String()
}

func waitForHealth(t *testing.T) {
	t.Helper()
	deadline := time.Now().Add(45 * time.Second)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		command := exec.CommandContext(ctx, "curl", "--noproxy", "*", "-fsS", "http://127.0.0.1:8080/healthz")
		if output, err := command.Output(); err == nil && string(output) == "{\"service\":\"winch-sandbox\",\"status\":\"ok\"}\n" {
			cancel()
			return
		}
		cancel()
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("sandbox did not become healthy:\n%s", compose(t, "logs", "sandbox"))
}

func commandOutput(t *testing.T, name string, args ...string) string {
	t.Helper()
	output, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, output)
	}
	return string(output)
}

func requireFailure(t *testing.T, name string, args ...string) {
	t.Helper()
	if output, err := exec.Command(name, args...).CombinedOutput(); err == nil {
		t.Fatalf("expected %s %v to fail, got %s", name, args, output)
	}
}

func assertContains(t *testing.T, value, wanted string) {
	t.Helper()
	if !bytes.Contains([]byte(value), []byte(wanted)) {
		t.Fatalf("%q does not contain %q", value, wanted)
	}
}
