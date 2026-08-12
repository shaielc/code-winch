package protocol

import (
	"bytes"
	"encoding/json"
	"fmt"
)

const (
	RunnerProtocolMajor          = uint16(1)
	RunnerProtocolMinor          = uint16(0)
	DefaultMaxRunnerMessageBytes = 1 << 20
	DefaultMaxRunnerBufferBytes  = 8 << 20
)

// RunnerVersion is negotiated before commands are assigned. Minor versions are
// additive; unknown fields are retained when a message is forwarded.
type RunnerVersion struct {
	Major      uint16
	Minor      uint16
	Additional map[string]json.RawMessage
}

func (v *RunnerVersion) UnmarshalJSON(data []byte) error {
	fields, err := decodeObject(data)
	if err != nil {
		return err
	}
	if err = take(fields, "major", &v.Major); err != nil {
		return err
	}
	if err = take(fields, "minor", &v.Minor); err != nil {
		return err
	}
	v.Additional = fields
	return nil
}

func (v RunnerVersion) MarshalJSON() ([]byte, error) {
	fields := cloneFields(v.Additional)
	put(fields, "major", v.Major)
	put(fields, "minor", v.Minor)
	return json.Marshal(fields)
}

type RunnerLimits struct {
	MaxMessageBytes        uint64 `json:"maxMessageBytes"`
	MaxBufferBytes         uint64 `json:"maxBufferBytes"`
	MaxOutstandingCommands uint32 `json:"maxOutstandingCommands"`
}

// Command payloads contain portable resource descriptions, never opaque OS or
// container handles. Raw bytes are base64 encoded by encoding/json.
type PreparePayload struct {
	WorkspaceID string `json:"workspaceId"`
}
type StartPayload struct {
	LaunchProfile string `json:"launchProfile"`
}
type InputPayload struct {
	InputID string `json:"inputId"`
	Text    string `json:"text,omitempty"`
	Bytes   []byte `json:"bytes,omitempty"`
}
type ResizePayload struct {
	Columns uint16 `json:"columns"`
	Rows    uint16 `json:"rows"`
}
type StopPayload struct {
	GraceMilliseconds uint32 `json:"graceMilliseconds"`
}
type InspectPayload struct{}
type CleanupPayload struct{}
type ObservationPayload struct {
	Type  string `json:"type"`
	Bytes []byte `json:"bytes,omitempty"`
	State string `json:"state,omitempty"`
}

type RunnerProtocolOffer struct {
	Major    uint16       `json:"major"`
	MinMinor uint16       `json:"minMinor"`
	MaxMinor uint16       `json:"maxMinor"`
	Limits   RunnerLimits `json:"limits"`
}

// RunnerProtocolError contains identifiers and bounded field names only. It is
// safe for telemetry and never includes message payloads or lease tokens.
type RunnerProtocolError struct{ Code, CommandID, ExecutionID, Field string }

func (e *RunnerProtocolError) Error() string {
	return fmt.Sprintf("runner protocol failed: code=%s command_id=%s execution_id=%s field=%s", e.Code, e.CommandID, e.ExecutionID, e.Field)
}

// NegotiateRunnerProtocol selects the highest common minor version. Version 1
// deliberately rejects major-version fallback.
func NegotiateRunnerProtocol(local, remote RunnerProtocolOffer) (RunnerVersion, RunnerLimits, error) {
	if local.Major != remote.Major || local.Major != RunnerProtocolMajor {
		return RunnerVersion{}, RunnerLimits{}, &RunnerProtocolError{Code: "RUNNER_MAJOR_VERSION_MISMATCH", Field: "major"}
	}
	low, high := local.MinMinor, local.MaxMinor
	if remote.MinMinor > low {
		low = remote.MinMinor
	}
	if remote.MaxMinor < high {
		high = remote.MaxMinor
	}
	if low > high {
		return RunnerVersion{}, RunnerLimits{}, &RunnerProtocolError{Code: "RUNNER_MINOR_VERSION_MISMATCH", Field: "minor"}
	}
	limits := RunnerLimits{
		MaxMessageBytes:        min64(local.Limits.MaxMessageBytes, remote.Limits.MaxMessageBytes),
		MaxBufferBytes:         min64(local.Limits.MaxBufferBytes, remote.Limits.MaxBufferBytes),
		MaxOutstandingCommands: min32(local.Limits.MaxOutstandingCommands, remote.Limits.MaxOutstandingCommands),
	}
	if limits.MaxMessageBytes == 0 || limits.MaxBufferBytes < limits.MaxMessageBytes || limits.MaxOutstandingCommands == 0 {
		return RunnerVersion{}, RunnerLimits{}, &RunnerProtocolError{Code: "RUNNER_INVALID_LIMITS", Field: "limits"}
	}
	return RunnerVersion{Major: local.Major, Minor: high}, limits, nil
}

func min64(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
func min32(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

// RunnerMessage represents every command and observation without exposing a
// process/container handle. Payload is interpreted according to Kind.
type RunnerMessage struct {
	Version     RunnerVersion
	Kind        string
	CommandID   string
	ExecutionID string
	LeaseToken  string
	Ordinal     uint64
	Payload     json.RawMessage
	Additional  map[string]json.RawMessage
}

func (m *RunnerMessage) UnmarshalJSON(data []byte) error {
	fields, err := decodeObject(data)
	if err != nil {
		return err
	}
	if err = take(fields, "version", &m.Version); err != nil {
		return err
	}
	if err = take(fields, "kind", &m.Kind); err != nil {
		return err
	}
	if err = take(fields, "commandId", &m.CommandID); err != nil {
		return err
	}
	if err = take(fields, "executionId", &m.ExecutionID); err != nil {
		return err
	}
	if err = take(fields, "leaseToken", &m.LeaseToken); err != nil {
		return err
	}
	if err = take(fields, "ordinal", &m.Ordinal); err != nil {
		return err
	}
	if err = take(fields, "payload", &m.Payload); err != nil {
		return err
	}
	m.Additional = fields
	return nil
}

func (m RunnerMessage) MarshalJSON() ([]byte, error) {
	fields := cloneFields(m.Additional)
	put(fields, "version", m.Version)
	put(fields, "kind", m.Kind)
	put(fields, "commandId", m.CommandID)
	put(fields, "executionId", m.ExecutionID)
	put(fields, "leaseToken", m.LeaseToken)
	if m.Ordinal != 0 {
		put(fields, "ordinal", m.Ordinal)
	}
	fields["payload"] = m.Payload
	return json.Marshal(fields)
}

func DecodeRunnerMessage(data []byte, maxBytes uint64) (RunnerMessage, error) {
	if maxBytes == 0 {
		maxBytes = DefaultMaxRunnerMessageBytes
	}
	if uint64(len(data)) > maxBytes {
		return RunnerMessage{}, &RunnerProtocolError{Code: "RUNNER_MESSAGE_TOO_LARGE", Field: "message"}
	}
	var m RunnerMessage
	dec := json.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&m); err != nil {
		return m, &RunnerProtocolError{Code: "RUNNER_INVALID_JSON", Field: "message"}
	}
	if err := m.Validate(); err != nil {
		return m, err
	}
	return m, nil
}

// EncodeRunnerMessage enforces the same negotiated boundary on producers.
func EncodeRunnerMessage(message RunnerMessage, maxBytes uint64) ([]byte, error) {
	if err := message.Validate(); err != nil {
		return nil, err
	}
	if maxBytes == 0 {
		maxBytes = DefaultMaxRunnerMessageBytes
	}
	data, err := json.Marshal(message)
	if err != nil {
		return nil, &RunnerProtocolError{Code: "RUNNER_INVALID_PAYLOAD", CommandID: message.CommandID, ExecutionID: message.ExecutionID, Field: "message"}
	}
	if uint64(len(data)) > maxBytes {
		return nil, &RunnerProtocolError{Code: "RUNNER_MESSAGE_TOO_LARGE", CommandID: message.CommandID, ExecutionID: message.ExecutionID, Field: "message"}
	}
	return data, nil
}

func (m RunnerMessage) Validate() error {
	fail := func(code, field string) error {
		return &RunnerProtocolError{Code: code, CommandID: m.CommandID, ExecutionID: m.ExecutionID, Field: field}
	}
	if m.Version.Major != RunnerProtocolMajor {
		return fail("RUNNER_MAJOR_VERSION_MISMATCH", "version.major")
	}
	if m.Kind == "" {
		return fail("RUNNER_REQUIRED_FIELD", "kind")
	}
	if m.CommandID == "" {
		return fail("RUNNER_REQUIRED_FIELD", "commandId")
	}
	if m.ExecutionID == "" {
		return fail("RUNNER_REQUIRED_FIELD", "executionId")
	}
	if m.LeaseToken == "" {
		return fail("RUNNER_REQUIRED_FIELD", "leaseToken")
	}
	if len(m.Payload) == 0 || !json.Valid(m.Payload) {
		return fail("RUNNER_INVALID_PAYLOAD", "payload")
	}
	known := map[string]bool{"prepare": true, "start": true, "input": true, "resize": true, "stop": true, "inspect": true, "cleanup": true, "observation": true}
	if !known[m.Kind] {
		return fail("RUNNER_UNKNOWN_MESSAGE_KIND", "kind")
	}
	if m.Kind == "observation" && m.Ordinal == 0 {
		return fail("RUNNER_REQUIRED_FIELD", "ordinal")
	}
	return nil
}
