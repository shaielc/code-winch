// Package fake provides a deterministic harness for contract and end-to-end
// tests. It is deliberately provider-neutral and performs no network access.
package fake

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/pkg/protocol"
)

const AdapterID = "fake"

type Driver struct{}

func (Driver) Describe(context.Context) (application.HarnessDescriptor, error) {
	return application.HarnessDescriptor{ID: AdapterID, Version: "1.0.0", InputModes: []string{"json-lines"}, OutputModes: []string{"json-lines"}, Capabilities: map[string]bool{"structured-events": true}}, nil
}
func (Driver) BuildLaunch(_ context.Context, spec application.RunSpec, _ application.ResolvedCredentials) (application.LaunchSpec, error) {
	if spec.RunID.IsZero() {
		return application.LaunchSpec{}, errors.New("fake harness: code=INVALID_RUN field=run_id")
	}
	return application.LaunchSpec{Command: "fake-harness", Args: []string{"--run-id", spec.RunID.String()}}, nil
}
func (Driver) NewCodec(context.Context, application.RunSpec) (application.HarnessCodec, error) {
	return &Codec{}, nil
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
			return events, err
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
		c.pending = nil
		return nil, errors.New("fake harness codec: code=TRUNCATED_RECORD")
	}
	c.pending = nil
	return nil, nil
}

var _ application.HarnessDriver = Driver{}
var _ application.HarnessCodec = (*Codec)(nil)
