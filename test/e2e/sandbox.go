package e2e

import (
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

type response struct {
	status      int
	contentType string
	body        string
}

func sandboxURL(t *testing.T) string {
	t.Helper()
	url := os.Getenv("WINCH_E2E_URL")
	if url == "" {
		t.Fatal("WINCH_E2E_URL is not set: run `make docker-e2e` or `make test-cycle`, or point WINCH_E2E_URL at a running sandbox (`make test-env` prints its URL)")
	}
	return url
}

func get(t *testing.T, url string) response {
	t.Helper()
	client := &http.Client{Timeout: 5 * time.Second, Transport: &http.Transport{Proxy: nil}}
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v\nis the test environment up? start it with `make test-env`", url, err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read %s: %v", url, err)
	}
	return response{status: resp.StatusCode, contentType: resp.Header.Get("Content-Type"), body: string(body)}
}
