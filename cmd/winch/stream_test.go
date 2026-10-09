package main

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

	"github.com/shaielc/code-winch/internal/adapters/transport/attach"
	"github.com/shaielc/code-winch/internal/runner"
)

// terminalSink collects the CLI's output one record per line and cancels the
// stream once the terminal record has been printed, because the stream itself
// stays open after the harness ends.
type terminalSink struct {
	lines  []string
	cancel context.CancelFunc
}

func (s *terminalSink) Write(p []byte) (int, error) {
	line := strings.TrimSuffix(string(p), "\n")
	s.lines = append(s.lines, line)
	if strings.Contains(line, `"kind":"session.terminated"`) {
		s.cancel()
	}
	return len(p), nil
}

// A child writes a burst well over the 32 KiB default message limit of the
// websocket library and over the runner's read buffer. `winch stream` must
// print every record: the concatenated data and the terminal record.
func TestStreamPrintsBurstsLargerThanTheReadBuffer(t *testing.T) {
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
			handler, err := attach.New(directory, attach.DefaultPosture(), session)
			if err != nil {
				t.Fatal(err)
			}
			server := httptest.NewServer(handler)
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			command := fmt.Sprintf(`head -c %d /dev/zero | tr '\000' '%s'`, size, test.octal)
			done := runner.Start(ctx, session, runner.HarnessConfig{Executable: "sh", Args: []string{"-c", command}})

			sink := &terminalSink{cancel: cancel}
			if err := runStream(ctx, []string{"--url", server.URL}, sink); err != nil {
				t.Fatalf("after %d records: %v", len(sink.lines), err)
			}

			var got []byte
			var terminal *runner.TerminatedPayload
			for index, line := range sink.lines {
				var record struct {
					Ordinal uint64          `json:"ordinal"`
					Kind    string          `json:"kind"`
					Payload json.RawMessage `json:"payload"`
				}
				if err := json.Unmarshal([]byte(line), &record); err != nil {
					t.Fatal(err)
				}
				if record.Ordinal != uint64(index+1) {
					t.Fatalf("ordinal = %d, want %d", record.Ordinal, index+1)
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
				t.Fatalf("printed %d bytes, want %d identical bytes", len(got), size)
			}
			if terminal == nil || terminal.Outcome != "completed" || terminal.ExitCode == nil || *terminal.ExitCode != 0 {
				t.Fatalf("terminal = %#v", terminal)
			}
			if err := <-done; err != nil {
				t.Fatal(err)
			}
		})
	}
}
