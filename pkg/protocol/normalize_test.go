package protocol

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func normalizationEvent(kind, payload string) Event {
	return Event{EventID: "evt-21", RunID: "run-21", Sequence: 1,
		OccurredAt: time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC), Kind: kind, SchemaVersion: 1,
		Source: Source{Type: "harness", Adapter: "fixture", Version: "1"}, Sensitivity: SensitivityUserContent,
		Payload: json.RawMessage(payload), Extensions: map[string]json.RawMessage{}}
}

func TestNormalizeEveryKnownFamily(t *testing.T) {
	valid := map[string]string{
		"run.lifecycle": `{"state":"running"}`, "stream.raw": `{"stream":"stdout","encoding":"utf-8","data":"ok"}`,
		"message.created": `{"messageId":"m","role":"assistant","content":"ok"}`,
		"tool.call":       `{"toolCallId":"t","name":"read","arguments":{}}`, "tool.result": `{"toolCallId":"t","status":"succeeded","result":{}}`,
		"approval.requested": `{"approvalId":"a","action":"write","summary":"change"}`, "approval.resolved": `{"approvalId":"a","decision":"approved"}`,
		"file.changed":       `{"changeId":"c","operation":"modified","path":"a.go"}`,
		"artifact.published": `{"artifactId":"a","mediaType":"text/plain","size":0,"digest":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`,
		"usage.recorded":     `{"unit":"input-token","quantity":1}`, "diagnostic.emitted": `{"code":"NOTICE","severity":"info"}`,
		"workflow.linked": `{"workflowId":"w","stepId":"s","relation":"started"}`,
	}
	for kind, payload := range valid {
		t.Run(kind+"/valid", func(t *testing.T) {
			if result := NormalizeEvent(normalizationEvent(kind, payload), NormalizationPolicy{}); result.Degraded {
				t.Fatalf("valid family degraded: %s", result.Event.Payload)
			}
		})
		t.Run(kind+"/invalid", func(t *testing.T) {
			result := NormalizeEvent(normalizationEvent(kind, `{}`), NormalizationPolicy{})
			if !result.Degraded || result.Event.Kind != "diagnostic.emitted" || result.Event.Sensitivity != SensitivityConfidential {
				t.Fatalf("invalid family did not safely degrade: %+v", result)
			}
			if _, ok := result.Event.Extensions[InvalidEventExtension]; !ok {
				t.Fatal("raw evidence was not retained")
			}
		})
	}
}

func TestNormalizePreservesUnknownKindAndExtension(t *testing.T) {
	event := normalizationEvent("vendor.future", `{"new":true}`)
	event.Extensions["example.vendor/v2"] = json.RawMessage(`{"opaque":[1,2,3]}`)
	before, _ := json.Marshal(event)
	result := NormalizeEvent(event, NormalizationPolicy{})
	after, _ := json.Marshal(result.Event)
	if result.Degraded || !bytes.Equal(before, after) {
		t.Fatalf("compatible data changed\nbefore=%s\nafter=%s", before, after)
	}
}

func TestNormalizeSensitivityAndExtensionFailuresAreContentSafe(t *testing.T) {
	secret := "do-not-report-me"
	event := normalizationEvent("message.created", `{"messageId":"m","role":"assistant","content":"`+secret+`"}`)
	event.Sensitivity = ""
	result := NormalizeEvent(event, NormalizationPolicy{})
	if !result.Degraded || result.Event.Sensitivity != SensitivityConfidential {
		t.Fatal("missing sensitivity was not confidential")
	}
	if strings.Contains(string(result.Event.Payload), secret) {
		t.Fatal("diagnostic payload leaked content")
	}
	event = normalizationEvent("vendor.future", `{}`)
	event.Extensions["not namespaced"] = json.RawMessage(`{"secret":"` + secret + `"}`)
	if result = NormalizeEvent(event, NormalizationPolicy{}); !result.Degraded {
		t.Fatal("invalid extension namespace was accepted")
	}
}

func TestNormalizeEventSizeBoundary(t *testing.T) {
	base := normalizationEvent("vendor.future", `{}`)
	encoded, _ := json.Marshal(base)
	if got := NormalizeEvent(base, NormalizationPolicy{MaxEventBytes: len(encoded)}); got.Degraded {
		t.Fatal("event at limit degraded")
	}
	if got := NormalizeEvent(base, NormalizationPolicy{MaxEventBytes: len(encoded) - 1}); !got.Degraded {
		t.Fatal("event above limit accepted")
	} else if got.Event.Extensions != nil {
		t.Fatal("oversized evidence should be represented only by digest")
	}
}

func FuzzNormalizeExtensionDecoding(f *testing.F) {
	f.Add("example.vendor/v1", `{"ok":true}`)
	f.Add("bad namespace", `"secret"`)
	f.Fuzz(func(t *testing.T, name, value string) {
		event := normalizationEvent("vendor.future", `{}`)
		event.Extensions = map[string]json.RawMessage{name: json.RawMessage(value)}
		result := NormalizeEvent(event, NormalizationPolicy{MaxEventBytes: 4096})
		if err := result.Event.ValidateEnvelope(); err != nil {
			t.Fatalf("normalizer returned invalid envelope: %v", err)
		}
	})
}
