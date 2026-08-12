// Package codex adapts the Codex CLI JSON-lines protocol to the harness port.
// It does not invoke the CLI or access credentials itself.
package codex

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/pkg/protocol"
)

const (
	AdapterID          = "codex"
	AdapterVersion     = "1.0.0"
	ExtensionNamespace = "openai.codex/v1"
)

type Config struct {
	Executable string
	Model      string
}

type Driver struct{ config Config }

func New(config Config) (Driver, error) {
	if config.Executable == "" {
		config.Executable = "codex"
	}
	if strings.ContainsAny(config.Executable, "\x00\r\n") {
		return Driver{}, errors.New("codex harness: code=INVALID_CONFIG field=executable")
	}
	if strings.ContainsAny(config.Model, "\x00\r\n") {
		return Driver{}, errors.New("codex harness: code=INVALID_CONFIG field=model")
	}
	return Driver{config: config}, nil
}

func (d Driver) Describe(context.Context) (application.HarnessDescriptor, error) {
	return application.HarnessDescriptor{
		ID: AdapterID, Version: AdapterVersion,
		InputModes: []string{"text-stdin"}, OutputModes: []string{"json-lines"},
		Capabilities: map[string]bool{"structured-events": true, "incremental-input": false, "resume": false},
	}, nil
}

func (d Driver) BuildLaunch(_ context.Context, spec application.RunSpec, credentials application.ResolvedCredentials) (application.LaunchSpec, error) {
	if spec.RunID.IsZero() {
		return application.LaunchSpec{}, errors.New("codex harness: code=INVALID_RUN field=run_id")
	}
	if len(credentials) != 0 {
		return application.LaunchSpec{}, errors.New("codex harness: code=UNSUPPORTED_CREDENTIAL_INJECTION")
	}
	args := []string{"exec", "--json"}
	if d.config.Model != "" {
		args = append(args, "--model", d.config.Model)
	}
	args = append(args, "-")
	return application.LaunchSpec{Command: d.config.Executable, Args: args, Env: map[string]string{}}, nil
}

func (Driver) NewCodec(_ context.Context, spec application.RunSpec) (application.HarnessCodec, error) {
	if spec.RunID.IsZero() {
		return nil, errors.New("codex harness: code=INVALID_RUN field=run_id")
	}
	return &Codec{runID: spec.RunID.String()}, nil
}

func (Driver) MapExit(code int) application.HarnessExit {
	if code == 0 {
		return application.HarnessExit{Successful: true, Code: "OK", Message: "codex harness exited successfully"}
	}
	return application.HarnessExit{Code: "PROCESS_FAILED", Message: fmt.Sprintf("codex harness exited unsuccessfully: run_id=unknown exit_code=%d", code)}
}

type Codec struct {
	runID   string
	pending []byte
}

func (c *Codec) Consume(chunk application.OutputChunk) ([]application.UnsequencedEvent, error) {
	c.pending = append(c.pending, chunk.Data...)
	var events []application.UnsequencedEvent
	for {
		i := bytes.IndexByte(c.pending, '\n')
		if i < 0 {
			return events, nil
		}
		line := append([]byte(nil), c.pending[:i]...)
		c.pending = c.pending[i+1:]
		if len(line) == 0 {
			continue
		}
		events = append(events, c.decode(line)...)
	}
}

type nativeRecord struct {
	Type string `json:"type"`
	Item struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"item"`
	Message string `json:"message"`
}

func (c *Codec) decode(line []byte) []application.UnsequencedEvent {
	var record nativeRecord
	if json.Unmarshal(line, &record) != nil || record.Type == "" {
		return []application.UnsequencedEvent{rawEvent(line), diagnostic("CODEX_MALFORMED_RECORD", c.runID)}
	}
	text, kind := "", ""
	switch record.Type {
	case "item.completed", "item.started", "item.updated":
		if record.Item.Type == "agent_message" && record.Item.Text != "" {
			kind, text = "assistant.message.delta", record.Item.Text
		}
	case "error":
		kind, text = "diagnostic", record.Message
	}
	if kind == "" {
		return []application.UnsequencedEvent{rawEvent(line)}
	}
	payload, _ := json.Marshal(map[string]string{"text": text})
	extension, _ := json.Marshal(json.RawMessage(line))
	return []application.UnsequencedEvent{{Kind: kind, SchemaVersion: 1, Source: source(), Sensitivity: protocol.SensitivityUserContent, Payload: payload, Extensions: map[string][]byte{ExtensionNamespace: extension}}}
}

func rawEvent(line []byte) application.UnsequencedEvent {
	payload, _ := json.Marshal(map[string]string{"stream": "stdout", "encoding": "base64", "data": base64.StdEncoding.EncodeToString(line)})
	return application.UnsequencedEvent{Kind: "raw.output", SchemaVersion: 1, Source: source(), Sensitivity: protocol.SensitivityConfidential, Payload: payload}
}
func diagnostic(code, runID string) application.UnsequencedEvent {
	payload, _ := json.Marshal(map[string]string{"code": code, "message": "Codex emitted an invalid JSON-lines record", "runId": runID})
	return application.UnsequencedEvent{Kind: "diagnostic", SchemaVersion: 1, Source: source(), Sensitivity: protocol.SensitivityOperational, Payload: payload}
}
func source() protocol.Source {
	return protocol.Source{Type: "harness", Adapter: AdapterID, Version: AdapterVersion}
}

func (c *Codec) Encode(message application.InputMessage) ([]application.InputFrame, error) {
	if message.ID == "" {
		return nil, errors.New("codex harness codec: code=INVALID_INPUT field=id")
	}
	return []application.InputFrame{{Data: append([]byte(message.Text), '\n')}}, nil
}
func (c *Codec) Flush() ([]application.UnsequencedEvent, error) {
	if len(c.pending) == 0 {
		return nil, nil
	}
	line := append([]byte(nil), c.pending...)
	c.pending = nil
	return []application.UnsequencedEvent{rawEvent(line), diagnostic("CODEX_TRUNCATED_RECORD", c.runID)}, nil
}

var _ application.HarnessDriver = Driver{}
var _ application.HarnessCodec = (*Codec)(nil)
