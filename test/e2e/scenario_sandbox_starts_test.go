package e2e

import (
	"net"
	"net/http"
	"strings"
	"testing"
)

func TestSandboxStarts(t *testing.T) {
	compose(t, "down", "--remove-orphans")
	t.Cleanup(func() { compose(t, "down", "--remove-orphans") })
	compose(t, "up", "--build", "-d", "sandbox")
	waitForHealth(t)

	response, err := (&http.Client{Transport: &http.Transport{Proxy: nil}}).Get("http://127.0.0.1:8080/api/session")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK || response.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("session response: %s %q", response.Status, response.Header.Get("Content-Type"))
	}

	status := compose(t, "exec", "-T", "sandbox", "winch", "status")
	want := "service: winch-sandbox (ok)\nprofile: container-standard\nunenforced control: network-egress\n"
	if status != want {
		t.Fatalf("status = %q, want %q", status, want)
	}
	id := compose(t, "exec", "-T", "sandbox", "id", "-u")
	if strings.TrimSpace(id) == "0" {
		t.Fatal("sandbox runs as root")
	}

	config := compose(t, "config")
	assertContains(t, config, "host_ip: 127.0.0.1")
	connection, err := net.Dial("udp", "1.1.1.1:80")
	if err != nil {
		t.Fatalf("determine routable host address: %v", err)
	}
	host := strings.Split(connection.LocalAddr().String(), ":")[0]
	if err := connection.Close(); err != nil {
		t.Fatalf("close address probe: %v", err)
	}
	requireFailure(t, "curl", "--noproxy", "*", "-fsS", "--max-time", "2", "http://"+host+":8080/healthz")
	if commandOutput(t, "curl", "--noproxy", "*", "-fsS", "http://127.0.0.1:8080/healthz") == "" {
		t.Fatal("loopback became unhealthy")
	}
}
