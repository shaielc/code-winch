package attach_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shaielc/code-winch/internal/adapters/transport/attach"
)

func TestStaticAssetsAndSPAFallback(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "index.html"), "attach shell")
	writeFile(t, filepath.Join(dir, "asset.js"), "asset")
	handler, err := attach.New(dir, attach.DefaultPosture())
	if err != nil {
		t.Fatal(err)
	}
	for path, wanted := range map[string]string{"/asset.js": "asset", "/future/route": "attach shell"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), wanted) {
			t.Errorf("%s: %d %q", path, response.Code, response.Body.String())
		}
	}
}

func TestTraversalDoesNotEscapeStaticRoot(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "web")
	if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "index.html"), "shell")
	writeFile(t, filepath.Join(parent, "secret"), "secret-value")
	if err := os.Symlink(filepath.Join(parent, "secret"), filepath.Join(root, "linked-secret")); err != nil {
		t.Fatal(err)
	}
	handler, _ := attach.New(root, attach.DefaultPosture())
	for _, path := range []string{"http://example/%2e%2e/secret", "/linked-secret"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if strings.Contains(response.Body.String(), "secret-value") {
			t.Fatalf("served a file outside static root through %s", path)
		}
	}
}

func TestMissingIndexIsRejected(t *testing.T) {
	if _, err := attach.New(t.TempDir(), attach.DefaultPosture()); err == nil {
		t.Fatal("expected missing index to fail")
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}
