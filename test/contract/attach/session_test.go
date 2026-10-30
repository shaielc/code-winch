package attach_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shaielc/code-winch/internal/adapters/transport/attach"
)

func fixture(t *testing.T, posture attach.Posture, childStatus attach.ChildStatus) http.Handler {
	t.Helper()
	return fixtureWithSession(t, attach.Session{Posture: posture, ChildStatus: childStatus})
}

func fixtureWithSession(t *testing.T, session attach.Session) http.Handler {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("Attach page"), 0o600); err != nil {
		t.Fatal(err)
	}
	handler, err := attach.New(dir, session)
	if err != nil {
		t.Fatal(err)
	}
	return handler
}

func TestEndpointContracts(t *testing.T) {
	handler := fixture(t, attach.DefaultPosture(), attach.ChildDown)
	cases := []struct{ path, body, contentType string }{
		{"/healthz", "{\"service\":\"winch-sandbox\",\"status\":\"ok\"}\n", "application/json"},
		{"/api/session", "{\"profile\":\"container-standard\",\"unenforcedControls\":[\"network-egress\"],\"childStatus\":\"Down\"}\n", "application/json"},
	}
	for _, tc := range cases {
		request := httptest.NewRequest(http.MethodGet, tc.path, nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusOK || response.Body.String() != tc.body || response.Header().Get("Content-Type") != tc.contentType {
			t.Errorf("%s: code=%d type=%q body=%q", tc.path, response.Code, response.Header().Get("Content-Type"), response.Body.String())
		}
	}
}

func TestSessionSerializesSuppliedChildStatus(t *testing.T) {
	for _, status := range []attach.ChildStatus{attach.ChildDown, attach.ChildRunning, attach.ChildExited} {
		handler := fixture(t, attach.DefaultPosture(), status)
		request := httptest.NewRequest(http.MethodGet, "/api/session", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if !strings.Contains(response.Body.String(), `"childStatus":"`+string(status)+`"`) {
			t.Fatalf("session did not retain supplied child status: %q", response.Body.String())
		}
	}
}

func TestSessionReadsCurrentChildStatus(t *testing.T) {
	status := attach.ChildRunning
	handler := fixtureWithSession(t, attach.Session{Posture: attach.DefaultPosture(), ChildStatusSource: func() attach.ChildStatus { return status }})
	status = attach.ChildExited
	request := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if !strings.Contains(response.Body.String(), `"childStatus":"Exited"`) {
		t.Fatalf("session copied startup status: %q", response.Body.String())
	}
}

func TestRejectsMethodsAndUnknownAPI(t *testing.T) {
	handler := fixture(t, attach.DefaultPosture(), attach.ChildDown)
	request := httptest.NewRequest(http.MethodPost, "/api/session", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusMethodNotAllowed || response.Header().Get("Allow") != "GET, HEAD" {
		t.Fatalf("unexpected method response: %d %#v", response.Code, response.Header())
	}
	for _, path := range []string{"/api", "/api/missing"} {
		request = httptest.NewRequest(http.MethodGet, path, nil)
		response = httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusNotFound || response.Header().Get("Content-Type") != "application/json" || strings.Contains(response.Body.String(), "Attach page") {
			t.Fatalf("%s returned SPA or non-JSON: %d %q", path, response.Code, response.Body.String())
		}
	}
}

func TestDefaultPostureNamesNetworkEgress(t *testing.T) {
	posture := attach.DefaultPosture()
	if len(posture.UnenforcedControls) == 0 || posture.UnenforcedControls[0] != "network-egress" {
		t.Fatalf("unsafe default posture: %#v", posture)
	}
}

func TestHeadHasNoBody(t *testing.T) {
	server := httptest.NewServer(fixture(t, attach.DefaultPosture(), attach.ChildDown))
	defer server.Close()
	response, err := http.Head(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if len(body) != 0 {
		t.Fatalf("HEAD body = %q", body)
	}
}
