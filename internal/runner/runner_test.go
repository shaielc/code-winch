package runner

import (
	"bytes"
	"encoding/base64"
	"os/exec"
	"strings"
	"syscall"
	"testing"
)

func TestPumpPreservesArbitraryBytesAcrossLargeReads(t *testing.T) {
	prefix := bytes.Repeat([]byte("not-json\n"), readBufferSize/9+100)
	input := append(prefix, 0xff, 0xfe)
	session := NewSession()
	if err := pump(bytes.NewReader(input), session); err != nil {
		t.Fatal(err)
	}
	var got []byte
	for index, record := range session.records {
		if record.Ordinal != uint64(index+1) {
			t.Fatalf("unexpected ordinal %d", record.Ordinal)
		}
		payload := record.Payload.(StreamPayload)
		data := []byte(payload.Data)
		if payload.Encoding == "base64" {
			var err error
			data, err = base64.StdEncoding.DecodeString(payload.Data)
			if err != nil {
				t.Fatal(err)
			}
		}
		got = append(got, data...)
	}
	if !bytes.Equal(got, input) {
		t.Fatalf("output changed: got %d bytes, want %d", len(got), len(input))
	}
}

func TestTerminationMappings(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		outcome string
		code    *int
		signal  string
	}{
		{name: "completed", outcome: "completed", code: intPointer(0)},
		{name: "failed", err: exitError(t, "exit 7"), outcome: "failed", code: intPointer(7)},
		{name: "stopped", err: exitError(t, "kill -TERM $$"), outcome: "stopped", signal: "terminated"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			session := NewSession()
			appendTermination(session, test.err)
			payload := session.records[0].Payload.(TerminatedPayload)
			if payload.Outcome != test.outcome || payload.Signal != test.signal || !equalCode(payload.ExitCode, test.code) {
				t.Fatalf("payload = %#v", payload)
			}
		})
	}
}

func TestTwoSubscribersReceiveEveryRecord(t *testing.T) {
	session := NewSession()
	one, cancelOne := session.Subscribe()
	defer cancelOne()
	two, cancelTwo := session.Subscribe()
	defer cancelTwo()
	session.append("stream.raw", "user-content", StreamPayload{Data: "one"})
	session.append("stream.raw", "user-content", StreamPayload{Data: "two"})
	for index, channel := range []<-chan Record{one, two} {
		for ordinal := uint64(1); ordinal <= 2; ordinal++ {
			if record := <-channel; record.Ordinal != ordinal {
				t.Fatalf("reader %d ordinal = %d", index, record.Ordinal)
			}
		}
	}
}

func exitError(t *testing.T, command string) error {
	t.Helper()
	err := exec.Command("sh", "-c", command).Run()
	if !strings.Contains(command, "exit") {
		if status := err.(*exec.ExitError).Sys().(syscall.WaitStatus); !status.Signaled() {
			t.Fatal("process was not signaled")
		}
	}
	if err == nil {
		t.Fatal("command unexpectedly succeeded")
	}
	return err
}
func intPointer(value int) *int { return &value }
func equalCode(a, b *int) bool  { return a == nil && b == nil || a != nil && b != nil && *a == *b }
