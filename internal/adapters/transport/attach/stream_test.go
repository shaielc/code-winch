package attach

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/shaielc/code-winch/internal/runner"
)

// A child writes a burst well over the 32 KiB default message limit of the
// websocket library and over the runner's read buffer. Each case must arrive
// whole through a real socket: the concatenated data and the terminal record.
func TestStreamDeliversBurstsLargerThanTheReadBuffer(t *testing.T) {
	const size = 200_000
	tests := []struct {
		name  string
		octal string
		unit  byte
	}{
		{name: "plain ASCII", octal: `a`, unit: 'a'},
		{name: "control bytes", octal: `\001`, unit: 0x01},
		{name: "invalid UTF-8", octal: `\377`, unit: 0xff},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
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
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			connection, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/api/session/stream", nil)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = connection.CloseNow() }()
			connection.SetReadLimit(runner.MaxRecordBytes)
			command := fmt.Sprintf(`head -c %d /dev/zero | tr '\000' '%s'`, size, test.octal)
			done := runner.Start(ctx, session, runner.HarnessConfig{Executable: "sh", Args: []string{"-c", command}})

			var got []byte
			var terminal *runner.TerminatedPayload
			for ordinal := uint64(1); terminal == nil; ordinal++ {
				_, message, err := connection.Read(ctx)
				if err != nil {
					t.Fatalf("after %d bytes: %v", len(got), err)
				}
				var record struct {
					Ordinal uint64          `json:"ordinal"`
					Kind    string          `json:"kind"`
					Payload json.RawMessage `json:"payload"`
				}
				if err := json.Unmarshal(message, &record); err != nil {
					t.Fatal(err)
				}
				if record.Ordinal != ordinal {
					t.Fatalf("ordinal = %d, want %d", record.Ordinal, ordinal)
				}
				switch record.Kind {
				case "stream.raw":
					var payload runner.StreamPayload
					if err := json.Unmarshal(record.Payload, &payload); err != nil {
						t.Fatal(err)
					}
					data := []byte(payload.Data)
					if payload.Encoding == "base64" {
						if data, err = base64.StdEncoding.DecodeString(payload.Data); err != nil {
							t.Fatal(err)
						}
					}
					got = append(got, data...)
				case "session.terminated":
					terminal = new(runner.TerminatedPayload)
					if err := json.Unmarshal(record.Payload, terminal); err != nil {
						t.Fatal(err)
					}
				default:
					t.Fatalf("unexpected kind %q", record.Kind)
				}
			}
			if want := bytes.Repeat([]byte{test.unit}, size); !bytes.Equal(got, want) {
				t.Fatalf("received %d bytes, want %d identical bytes", len(got), size)
			}
			if terminal.Outcome != "completed" || terminal.ExitCode == nil || *terminal.ExitCode != 0 {
				t.Fatalf("terminal = %#v", terminal)
			}
			if err := <-done; err != nil {
				t.Fatal(err)
			}
		})
	}
}

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
