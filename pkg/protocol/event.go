// Package protocol defines versioned, provider-neutral wire contracts.
package protocol

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"
)

// Sensitivity controls retention, export, and telemetry handling.
type Sensitivity string

const (
	SensitivityPublic       Sensitivity = "public"
	SensitivityOperational  Sensitivity = "operational"
	SensitivityUserContent  Sensitivity = "user-content"
	SensitivityConfidential Sensitivity = "confidential"
	SensitivitySecret       Sensitivity = "secret"
)

// Source identifies the component that observed an event. Additional fields
// are retained so additive protocol changes can be forwarded losslessly.
type Source struct {
	Type       string
	Adapter    string
	Version    string
	Additional map[string]json.RawMessage
}

// Event is the canonical event envelope. Payload and extensions deliberately
// remain JSON: provider adapters may not leak their types into this package,
// and generic consumers must be able to forward unknown kinds.
type Event struct {
	EventID       string
	RunID         string
	Sequence      uint64
	OccurredAt    time.Time
	Kind          string
	SchemaVersion uint
	Source        Source
	Sensitivity   Sensitivity
	Payload       json.RawMessage
	Extensions    map[string]json.RawMessage
	Additional    map[string]json.RawMessage
}

// ValidationError is safe for logs: it contains a stable code and resource
// IDs, but never payload, extension, or other free-form content.
type ValidationError struct {
	Code    string
	EventID string
	RunID   string
	Field   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("event validation failed: code=%s event_id=%s run_id=%s field=%s", e.Code, e.EventID, e.RunID, e.Field)
}

func decodeObject(data []byte) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&fields); err != nil || fields == nil {
		return nil, errors.New("event decode failed: expected a JSON object")
	}
	return fields, nil
}

func take[T any](fields map[string]json.RawMessage, name string, target *T) error {
	raw, ok := fields[name]
	if !ok {
		return nil
	}
	delete(fields, name)
	return json.Unmarshal(raw, target)
}

func (s *Source) UnmarshalJSON(data []byte) error {
	fields, err := decodeObject(data)
	if err != nil {
		return err
	}
	if err = take(fields, "type", &s.Type); err != nil {
		return err
	}
	if err = take(fields, "adapter", &s.Adapter); err != nil {
		return err
	}
	if err = take(fields, "version", &s.Version); err != nil {
		return err
	}
	s.Additional = fields
	return nil
}

func (s Source) MarshalJSON() ([]byte, error) {
	fields := cloneFields(s.Additional)
	put(fields, "type", s.Type)
	put(fields, "adapter", s.Adapter)
	put(fields, "version", s.Version)
	return json.Marshal(fields)
}

func (e *Event) UnmarshalJSON(data []byte) error {
	fields, err := decodeObject(data)
	if err != nil {
		return err
	}
	if err = take(fields, "eventId", &e.EventID); err != nil {
		return err
	}
	if err = take(fields, "runId", &e.RunID); err != nil {
		return err
	}
	if err = take(fields, "sequence", &e.Sequence); err != nil {
		return err
	}
	if err = take(fields, "occurredAt", &e.OccurredAt); err != nil {
		return err
	}
	if err = take(fields, "kind", &e.Kind); err != nil {
		return err
	}
	if err = take(fields, "schemaVersion", &e.SchemaVersion); err != nil {
		return err
	}
	if err = take(fields, "source", &e.Source); err != nil {
		return err
	}
	if err = take(fields, "sensitivity", &e.Sensitivity); err != nil {
		return err
	}
	if err = take(fields, "payload", &e.Payload); err != nil {
		return err
	}
	if err = take(fields, "extensions", &e.Extensions); err != nil {
		return err
	}
	e.Additional = fields
	return nil
}

func cloneFields(source map[string]json.RawMessage) map[string]json.RawMessage {
	result := make(map[string]json.RawMessage, len(source)+10)
	for key, value := range source {
		result[key] = value
	}
	return result
}

func put(fields map[string]json.RawMessage, name string, value any) {
	encoded, _ := json.Marshal(value)
	fields[name] = encoded
}

func (e Event) MarshalJSON() ([]byte, error) {
	fields := cloneFields(e.Additional)
	put(fields, "eventId", e.EventID)
	put(fields, "runId", e.RunID)
	put(fields, "sequence", e.Sequence)
	put(fields, "occurredAt", e.OccurredAt)
	put(fields, "kind", e.Kind)
	put(fields, "schemaVersion", e.SchemaVersion)
	put(fields, "source", e.Source)
	put(fields, "sensitivity", e.Sensitivity)
	fields["payload"] = e.Payload
	put(fields, "extensions", e.Extensions)
	return json.Marshal(fields)
}

// ValidateEnvelope performs content-safe validation common to known and
// unknown event kinds. Payload validation is defined by the JSON schemas.
func (e Event) ValidateEnvelope() error {
	fail := func(code, field string) error {
		return &ValidationError{Code: code, EventID: e.EventID, RunID: e.RunID, Field: field}
	}
	if e.EventID == "" {
		return fail("EVENT_REQUIRED_FIELD", "eventId")
	}
	if e.RunID == "" {
		return fail("EVENT_REQUIRED_FIELD", "runId")
	}
	if e.Sequence == 0 {
		return fail("EVENT_INVALID_SEQUENCE", "sequence")
	}
	if e.OccurredAt.IsZero() {
		return fail("EVENT_REQUIRED_FIELD", "occurredAt")
	}
	if e.Kind == "" {
		return fail("EVENT_REQUIRED_FIELD", "kind")
	}
	if e.SchemaVersion == 0 {
		return fail("EVENT_UNSUPPORTED_VERSION", "schemaVersion")
	}
	if e.Source.Type == "" {
		return fail("EVENT_REQUIRED_FIELD", "source.type")
	}
	if !e.Sensitivity.Valid() {
		return fail("EVENT_INVALID_SENSITIVITY", "sensitivity")
	}
	if len(e.Payload) == 0 || !json.Valid(e.Payload) {
		return fail("EVENT_INVALID_PAYLOAD", "payload")
	}
	return nil
}

// SortBySequence orders events by their durable run sequence. Timestamps and
// IDs are intentionally never used as ordering tie-breakers.
func SortBySequence(events []Event) {
	sort.SliceStable(events, func(i, j int) bool { return events[i].Sequence < events[j].Sequence })
}
