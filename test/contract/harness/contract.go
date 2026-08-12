// Package harness contains the reusable harness adapter contract suite.
package harness

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/shaielc/code-winch/internal/application"
)

type Factory func(t *testing.T) application.HarnessDriver

// Run verifies provider-neutral descriptor, launch, incremental codec, input,
// flush, and content-safe diagnostic behavior.
func Run(t *testing.T, factory Factory, spec application.RunSpec, validRecord []byte) {
	t.Helper()
	driver := factory(t)
	descriptor, err := driver.Describe(context.Background())
	if err != nil || descriptor.ID == "" || descriptor.Version == "" || len(descriptor.InputModes) == 0 || len(descriptor.OutputModes) == 0 {
		t.Fatalf("invalid descriptor: %#v, %v", descriptor, err)
	}
	if _, err = driver.BuildLaunch(context.Background(), spec, nil); err != nil {
		t.Fatalf("BuildLaunch: %v", err)
	}

	for boundary := 0; boundary <= len(validRecord); boundary++ {
		codec, codecErr := driver.NewCodec(context.Background(), spec)
		if codecErr != nil {
			t.Fatalf("NewCodec: %v", codecErr)
		}
		var events []application.UnsequencedEvent
		for _, part := range [][]byte{validRecord[:boundary], validRecord[boundary:]} {
			got, consumeErr := codec.Consume(application.OutputChunk{Data: part})
			if consumeErr != nil {
				t.Fatalf("boundary %d: %v", boundary, consumeErr)
			}
			events = append(events, got...)
		}
		if len(events) != 1 {
			t.Fatalf("boundary %d: got %d events", boundary, len(events))
		}
		if _, flushErr := codec.Flush(); flushErr != nil {
			t.Fatalf("boundary %d flush: %v", boundary, flushErr)
		}
	}

	codec, _ := driver.NewCodec(context.Background(), spec)
	frames, err := codec.Encode(application.InputMessage{ID: "command-1", Text: "a newline\nand unicode ☃"})
	if err != nil || len(frames) != 1 || !bytes.HasSuffix(frames[0].Data, []byte("\n")) {
		t.Fatalf("Encode: %#v, %v", frames, err)
	}
	secret := "do-not-log-this"
	_, err = codec.Consume(application.OutputChunk{Data: []byte("{\"payload\":\"" + secret + "\"}\n")})
	if err == nil || strings.Contains(err.Error(), secret) {
		t.Fatalf("malformed diagnostic must be stable and content-safe: %v", err)
	}
	codec, _ = driver.NewCodec(context.Background(), spec)
	_, _ = codec.Consume(application.OutputChunk{Data: []byte("{")})
	if _, err = codec.Flush(); err == nil {
		t.Fatal("Flush accepted a truncated record")
	}
}
