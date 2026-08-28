// Package fake provides a deterministic harness for contract and end-to-end
// tests. It is deliberately provider-neutral and performs no network access.
package fake

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/pkg/protocol"
)

const AdapterID = "fake"

type Config struct {
	Executable    string
	Transcript    string
	Delay         time.Duration
	ForceFailure  bool
	MalformedLine bool
	EarlyExit     bool
}

type Driver struct{ Config Config }

func (Driver) Describe(context.Context) (application.HarnessDescriptor, error) {
	return application.HarnessDescriptor{ID: AdapterID, Version: "1.0.0", InputModes: []string{"json-lines"}, OutputModes: []string{"json-lines"}, Capabilities: map[string]bool{"structured-events": true}}, nil
}
func (d Driver) BuildLaunch(_ context.Context, spec application.RunSpec, _ application.ResolvedCredentials) (application.LaunchSpec, error) {
	if spec.RunID.IsZero() {
		return application.LaunchSpec{}, errors.New("fake harness: code=INVALID_RUN field=run_id")
	}
	command, err := resolveExecutable(d.Config.Executable)
	if err != nil {
		return application.LaunchSpec{}, err
	}
	if d.Config.Delay < 0 {
		return application.LaunchSpec{}, errors.New("fake harness: code=INVALID_CONFIG field=delay")
	}
	args := []string{"--run-id", spec.RunID.String()}
	if d.Config.Transcript != "" {
		args = append(args, "--transcript", d.Config.Transcript)
	}
	if d.Config.Delay != 0 {
		args = append(args, "--delay", d.Config.Delay.String())
	}
	if d.Config.ForceFailure {
		args = append(args, "--force-failure")
	}
	if d.Config.MalformedLine {
		args = append(args, "--malformed-line")
	}
	if d.Config.EarlyExit {
		args = append(args, "--early-exit")
	}
	return application.LaunchSpec{Command: command, Args: args}, nil
}

func resolveExecutable(configured string) (string, error) {
	if configured != "" {
		absolute, err := filepath.Abs(configured)
		if err != nil {
			return "", fmt.Errorf("fake harness: code=INVALID_CONFIG field=executable: %w", err)
		}
		return absolute, nil
	}
	resolved, err := exec.LookPath("fake-harness")
	if err != nil {
		return "", errors.New("fake harness: code=EXECUTABLE_NOT_FOUND field=executable")
	}
	absolute, err := filepath.Abs(resolved)
	if err != nil {
		return "", fmt.Errorf("fake harness: code=EXECUTABLE_NOT_FOUND field=executable: %w", err)
	}
	return absolute, nil
}
func (Driver) NewCodec(context.Context, application.RunSpec) (application.HarnessCodec, error) {
	return &Codec{}, nil
}
func (Driver) MapExit(code int) application.HarnessExit {
	if code == 0 {
		return application.HarnessExit{Successful: true, Code: "OK", Message: "fake harness exited successfully"}
	}
	return application.HarnessExit{Code: "PROCESS_FAILED", Message: "fake harness exited unsuccessfully"}
}

type record struct {
	Kind        string               `json:"kind"`
	Payload     json.RawMessage      `json:"payload"`
	Sensitivity protocol.Sensitivity `json:"sensitivity,omitempty"`
}

// Codec accepts newline-delimited JSON and retains partial records between
// calls, making every byte boundary equivalent to a single complete chunk.
type Codec struct{ pending []byte }

func (c *Codec) Consume(chunk application.OutputChunk) ([]application.UnsequencedEvent, error) {
	c.pending = append(c.pending, chunk.Data...)
	var events []application.UnsequencedEvent
	for {
		i := bytes.IndexByte(c.pending, '\n')
		if i < 0 {
			return events, nil
		}
		line := c.pending[:i]
		c.pending = c.pending[i+1:]
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		event, err := decode(line)
		if err != nil {
			events = append(events, fallback(line)...)
			continue
		}
		events = append(events, event)
	}
}

func decode(line []byte) (application.UnsequencedEvent, error) {
	var value record
	decoder := json.NewDecoder(bytes.NewReader(line))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return application.UnsequencedEvent{}, errors.New("fake harness codec: code=MALFORMED_RECORD")
	}
	if value.Kind == "" || len(value.Payload) == 0 || !json.Valid(value.Payload) {
		return application.UnsequencedEvent{}, errors.New("fake harness codec: code=INVALID_RECORD")
	}
	if value.Sensitivity == "" {
		value.Sensitivity = protocol.SensitivityUserContent
	}
	return application.UnsequencedEvent{Kind: value.Kind, SchemaVersion: 1, Source: protocol.Source{Type: "harness", Adapter: AdapterID, Version: "1.0.0"}, Sensitivity: value.Sensitivity, Payload: append([]byte(nil), value.Payload...)}, nil
}

func fallback(line []byte) []application.UnsequencedEvent {
	raw, _ := json.Marshal(map[string]string{"stream": "stdout", "encoding": "base64", "data": base64.StdEncoding.EncodeToString(line)})
	diagnostic, _ := json.Marshal(map[string]string{"code": "FAKE_MALFORMED_RECORD", "message": "Fake harness emitted an invalid JSON-lines record"})
	return []application.UnsequencedEvent{
		{Kind: "raw.output", SchemaVersion: 1, Source: protocol.Source{Type: "harness", Adapter: AdapterID, Version: "1.0.0"}, Sensitivity: protocol.SensitivityConfidential, Payload: raw},
		{Kind: "diagnostic", SchemaVersion: 1, Source: protocol.Source{Type: "harness", Adapter: AdapterID, Version: "1.0.0"}, Sensitivity: protocol.SensitivityOperational, Payload: diagnostic},
	}
}

func (c *Codec) Encode(message application.InputMessage) ([]application.InputFrame, error) {
	if message.ID == "" {
		return nil, errors.New("fake harness codec: code=INVALID_INPUT field=id")
	}
	data, err := json.Marshal(map[string]string{"id": message.ID, "text": message.Text})
	if err != nil {
		return nil, fmt.Errorf("fake harness codec: code=ENCODE_INPUT: %w", err)
	}
	return []application.InputFrame{{Data: append(data, '\n')}}, nil
}
func (c *Codec) Flush() ([]application.UnsequencedEvent, error) {
	if len(bytes.TrimSpace(c.pending)) != 0 {
		line := append([]byte(nil), c.pending...)
		c.pending = nil
		return fallback(line), nil
	}
	c.pending = nil
	return nil, nil
}

var _ application.HarnessDriver = Driver{}
var _ application.HarnessCodec = (*Codec)(nil)
