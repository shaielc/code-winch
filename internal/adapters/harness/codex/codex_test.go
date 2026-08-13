package codex_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shaielc/code-winch/internal/adapters/harness/codex"
	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
	harnesscontract "github.com/shaielc/code-winch/test/contract/harness"
)

func spec(t *testing.T) application.RunSpec {
	t.Helper()
	id, err := domain.ParseRunID("00000000-0000-0000-0000-000000000014")
	if err != nil {
		t.Fatal(err)
	}
	return application.RunSpec{RunID: id}
}

func driver(t *testing.T) codex.Driver {
	t.Helper()
	d, err := codex.New(codex.Config{})
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestHarnessContract(t *testing.T) {
	harnesscontract.Run(t, func(*testing.T) application.HarnessDriver { return driver(t) }, spec(t), []byte("{\"type\":\"item.completed\",\"item\":{\"type\":\"agent_message\",\"text\":\"hello\"}}\n"))
}

func TestCapabilitiesAndConfiguration(t *testing.T) {
	d := driver(t)
	descriptor, _ := d.Describe(context.Background())
	if !descriptor.Capabilities["structured-events"] || descriptor.Capabilities["incremental-input"] || descriptor.Capabilities["resume"] {
		t.Fatalf("capabilities: %#v", descriptor.Capabilities)
	}
	if _, err := codex.New(codex.Config{Model: "bad\nmodel"}); err == nil {
		t.Fatal("accepted unsafe model")
	}
	if _, err := d.BuildLaunch(context.Background(), spec(t), application.ResolvedCredentials{"token": {Bytes: []byte("secret-canary")}}); err == nil || strings.Contains(err.Error(), "secret-canary") {
		t.Fatalf("credential error was not safe: %v", err)
	}
}

func TestSanitizedTranscriptGolden(t *testing.T) {
	transcript, err := os.ReadFile("testdata/transcript.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	codec, _ := driver(t).NewCodec(context.Background(), spec(t))
	events, err := codec.Consume(application.OutputChunk{Data: transcript})
	if err != nil || len(events) != 4 {
		t.Fatalf("events=%d err=%v", len(events), err)
	}
	if events[2].Kind != "assistant.message.delta" || !bytes.Contains(events[2].Payload, []byte("Fixture response.")) || len(events[2].Extensions[codex.ExtensionNamespace]) == 0 {
		t.Fatalf("normalized message: %#v", events[2])
	}
}

func TestMalformedRecordPreservesEveryByte(t *testing.T) {
	codec, _ := driver(t).NewCodec(context.Background(), spec(t))
	native := []byte("{not-json:\xff}")
	events, err := codec.Consume(application.OutputChunk{Data: append(native, '\n')})
	if err != nil || len(events) != 2 || events[0].Kind != "raw.output" || events[1].Kind != "diagnostic" {
		t.Fatalf("%#v %v", events, err)
	}
	var payload struct {
		Data string `json:"data"`
	}
	if err := json.Unmarshal(events[0].Payload, &payload); err != nil {
		t.Fatal(err)
	}
	decoded, err := base64.StdEncoding.DecodeString(payload.Data)
	if err != nil || !bytes.Equal(decoded, native) {
		t.Fatalf("raw output was changed: %q %v", decoded, err)
	}
	if bytes.Contains(events[1].Payload, native) {
		t.Fatal("diagnostic leaked output content")
	}
}

func TestFakeCLIIntegration(t *testing.T) {
	script, err := filepath.Abs("testdata/fake-codex.sh")
	if err != nil {
		t.Fatal(err)
	}
	d, err := codex.New(codex.Config{Executable: script})
	if err != nil {
		t.Fatal(err)
	}
	launch, err := d.BuildLaunch(context.Background(), spec(t), nil)
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(launch.Command, launch.Args...)
	cmd.Stdin = strings.NewReader("fixture prompt\n")
	output, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	codec, _ := d.NewCodec(context.Background(), spec(t))
	events, err := codec.Consume(application.OutputChunk{Data: output})
	if err != nil || len(events) != 4 {
		t.Fatalf("events=%d err=%v", len(events), err)
	}
}

func FuzzCodecChunkBoundaries(f *testing.F) {
	f.Add([]byte(`{"type":"item.completed","item":{"type":"agent_message","text":"hello"}}`), uint8(7))
	f.Add([]byte("not-json"), uint8(1))
	f.Fuzz(func(t *testing.T, record []byte, width uint8) {
		if len(record) > 4096 {
			t.Skip()
		}
		codec, err := driver(t).NewCodec(context.Background(), spec(t))
		if err != nil {
			t.Fatal(err)
		}
		step := int(width) + 1
		for start := 0; start < len(record); start += step {
			end := min(start+step, len(record))
			if _, err := codec.Consume(application.OutputChunk{Data: record[start:end]}); err != nil {
				t.Fatalf("consume arbitrary bytes: %v", err)
			}
		}
		if _, err := codec.Consume(application.OutputChunk{Data: []byte{'\n'}}); err != nil {
			t.Fatalf("finish record: %v", err)
		}
		if _, err := codec.Flush(); err != nil {
			t.Fatalf("flush: %v", err)
		}
	})
}
