package protocol

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"
)

const (
	// DefaultMaxEventBytes bounds one canonical event before persistence.
	DefaultMaxEventBytes = 256 * 1024
	// InvalidEventExtension is the non-authoritative evidence retained when a
	// provider record cannot be represented by a known canonical family.
	InvalidEventExtension = "code-winch.invalid/v1"
)

var extensionName = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]*/v[1-9][0-9]*$`)

// NormalizationPolicy contains ingestion limits. A zero limit selects the safe
// default; callers may lower, but not accidentally disable, the limit.
type NormalizationPolicy struct{ MaxEventBytes int }

// NormalizationResult always contains a persistable event. Degraded is true
// when malformed input was converted to a confidential diagnostic fallback.
type NormalizationResult struct {
	Event    Event
	Degraded bool
}

// NormalizeEvent validates and classifies a provider-produced event. Unknown
// kinds and valid namespaced extensions are forwarded losslessly. Malformed
// input becomes a diagnostic instead of terminating a run. The original bytes
// are retained when they fit the event limit; oversized evidence is represented
// by a length and SHA-256 digest so logs and diagnostics never contain content.
func NormalizeEvent(event Event, policy NormalizationPolicy) NormalizationResult {
	if event.Sensitivity == "" || !event.Sensitivity.Valid() {
		event.Sensitivity = SensitivityConfidential
		return degradedEvent(event, "EVENT_INVALID_SENSITIVITY", policy)
	}
	if err := event.ValidateEnvelope(); err != nil {
		code := "EVENT_INVALID_ENVELOPE"
		if validation, ok := err.(*ValidationError); ok {
			code = validation.Code
		}
		return degradedEvent(event, code, policy)
	}
	for name, value := range event.Extensions {
		if !extensionName.MatchString(name) || !validJSONObject(value) {
			return degradedEvent(event, "EVENT_INVALID_EXTENSION", policy)
		}
	}
	if code := validateKnownPayload(event.Kind, event.Payload); code != "" {
		return degradedEvent(event, code, policy)
	}
	encoded, err := json.Marshal(event)
	if err != nil || len(encoded) > maxEventBytes(policy) {
		return degradedEvent(event, "EVENT_SIZE_EXCEEDED", policy)
	}
	return NormalizationResult{Event: event}
}

func (s Sensitivity) Valid() bool {
	switch s {
	case SensitivityPublic, SensitivityOperational, SensitivityUserContent, SensitivityConfidential, SensitivitySecret:
		return true
	default:
		return false
	}
}

func maxEventBytes(policy NormalizationPolicy) int {
	if policy.MaxEventBytes <= 0 {
		return DefaultMaxEventBytes
	}
	return policy.MaxEventBytes
}

func validJSONObject(raw json.RawMessage) bool {
	var object map[string]json.RawMessage
	return json.Unmarshal(raw, &object) == nil && object != nil
}

func validateKnownPayload(kind string, raw json.RawMessage) string {
	var payload map[string]json.RawMessage
	if json.Unmarshal(raw, &payload) != nil || payload == nil {
		return "EVENT_INVALID_PAYLOAD"
	}
	required := map[string][]string{
		"run.lifecycle": {"state"}, "stream.raw": {"stream", "encoding", "data"},
		"message.created": {"messageId", "role", "content"}, "tool.call": {"toolCallId", "name", "arguments"},
		"tool.result": {"toolCallId", "status", "result"}, "approval.requested": {"approvalId", "action", "summary"},
		"approval.resolved": {"approvalId", "decision"}, "file.changed": {"changeId", "operation", "path"},
		"artifact.published": {"artifactId", "mediaType", "size", "digest"}, "usage.recorded": {"unit", "quantity"},
		"diagnostic.emitted": {"code", "severity"}, "workflow.linked": {"workflowId", "stepId", "relation"},
	}
	fields, known := required[kind]
	if !known {
		return ""
	}
	for _, field := range fields {
		if value, ok := payload[field]; !ok || string(value) == "null" {
			return "EVENT_INVALID_PAYLOAD"
		}
	}
	return ""
}

func degradedEvent(event Event, code string, policy NormalizationPolicy) NormalizationResult {
	raw, _ := json.Marshal(event)
	digest := sha256.Sum256(raw)
	payload, _ := json.Marshal(map[string]any{
		"code": code, "severity": "warning", "retryable": false,
		"evidenceBytes": len(raw), "evidenceSha256": hex.EncodeToString(digest[:]),
	})
	event.Kind = "diagnostic.emitted"
	event.SchemaVersion = 1
	event.Sensitivity = SensitivityConfidential
	event.Payload = payload
	event.Extensions = nil
	if len(raw) > 0 {
		evidence, _ := json.Marshal(map[string]json.RawMessage{"event": raw})
		event.Extensions = map[string]json.RawMessage{InvalidEventExtension: evidence}
		encoded, _ := json.Marshal(event)
		if len(encoded) > maxEventBytes(policy) {
			event.Extensions = nil
		}
	}
	return NormalizationResult{Event: event, Degraded: true}
}
