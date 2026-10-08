package attach

import (
	"context"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/shaielc/code-winch/internal/runner"
)

func TestStreamFansOutToConcurrentReaders(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "index.html"), []byte("ok"), 0o600); err != nil {
		t.Fatal(err)
	}
	session := runner.NewSession()
	handler, err := New(directory, DefaultPosture(), session)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()
	endpoint := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/session/stream"
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	one, _, err := websocket.Dial(ctx, endpoint, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = one.CloseNow() }()
	two, _, err := websocket.Dial(ctx, endpoint, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = two.CloseNow() }()
	done := runner.Start(ctx, session, runner.HarnessConfig{Executable: "sh", Args: []string{"-c", "printf arbitrary-output"}})
	for index, connection := range []*websocket.Conn{one, two} {
		for ordinal := 1; ordinal <= 2; ordinal++ {
			_, message, err := connection.Read(ctx)
			if err != nil {
				t.Fatalf("reader %d: %v", index, err)
			}
			if !strings.Contains(string(message), `"ordinal":`+string(rune('0'+ordinal))) {
				t.Fatalf("reader %d record %d: %s", index, ordinal, message)
			}
		}
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}
