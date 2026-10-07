package e2e

import (
	"net/http"
	"os/exec"
	"strings"
	"testing"
)

func TestSandboxStarts(t *testing.T) {
	base := sandboxURL(t)

	assertJSON(t, get(t, base+"/healthz"), "{\"service\":\"winch-sandbox\",\"status\":\"ok\"}\n")
	assertJSON(t, get(t, base+"/api/session"), "{\"profile\":\"container-standard\",\"unenforcedControls\":[\"network-egress\"]}\n")

	page := get(t, base+"/")
	if page.status != http.StatusOK || !strings.HasPrefix(page.contentType, "text/html") || !strings.Contains(page.body, "Sandbox attach") {
		t.Fatalf("page: %d %q %q", page.status, page.contentType, page.body)
	}

	command := exec.Command("go", "run", "./cmd/winch", "status", "--url", base)
	command.Dir = "../.."
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("winch status: %v\n%s", err, output)
	}
	want := "service: winch-sandbox (ok)\nprofile: container-standard\nunenforced control: network-egress\n"
	if string(output) != want {
		t.Fatalf("status = %q, want %q", output, want)
	}
}

func assertJSON(t *testing.T, got response, wantBody string) {
	t.Helper()
	if got.status != http.StatusOK || got.contentType != "application/json" || got.body != wantBody {
		t.Fatalf("response: %d %q %q, want 200 \"application/json\" %q", got.status, got.contentType, got.body, wantBody)
	}
}
