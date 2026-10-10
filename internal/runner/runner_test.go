package runner

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
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
	if got := streamedBytes(t, session); !bytes.Equal(got, input) {
		t.Fatalf("output changed: got %d bytes, want %d", len(got), len(input))
	}
}

// Every chunk class JSON-escapes or re-encodes differently, so each is pumped at
// full buffer size: the largest record the runner can emit must still fit the
// bound that the transport and the CLI are built to carry.
func TestPumpRecordsFitMaxRecordBytes(t *testing.T) {
	tests := []struct {
		name  string
		unit  []byte
		wider bool
	}{
		{name: "plain ASCII", unit: []byte("a")},
		{name: "control bytes escape to six bytes each", unit: []byte{0x01}, wider: true},
		{name: "HTML characters escape to six bytes each", unit: []byte("<"), wider: true},
		{name: "quotes escape to two bytes each", unit: []byte(`"`)},
		{name: "line separator escapes to six bytes for three", unit: []byte(" ")},
		{name: "invalid UTF-8 is sent as base64", unit: []byte{0xff}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := bytes.Repeat(test.unit, 3*readBufferSize/len(test.unit))
			session := NewSession()
			if err := pump(bytes.NewReader(input), session); err != nil {
				t.Fatal(err)
			}
			largest := 0
			for _, record := range session.records {
				encoded, err := json.Marshal(record)
				if err != nil {
					t.Fatal(err)
				}
				largest = max(largest, len(encoded))
			}
			if largest > MaxRecordBytes {
				t.Fatalf("largest record is %d bytes, over the %d byte bound", largest, MaxRecordBytes)
			}
			if test.wider && largest <= MaxRecordBytes/2 {
				t.Fatalf("largest record is %d bytes: the worst case is no longer exercised", largest)
			}
			if got := streamedBytes(t, session); !bytes.Equal(got, input) {
				t.Fatalf("output changed: got %d bytes, want %d", len(got), len(input))
			}
		})
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

// streamedBytes concatenates the decoded data of every record in order, and
// fails if the ordinals are not contiguous from 1.
func streamedBytes(t *testing.T, session *Session) []byte {
	t.Helper()
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
	return got
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
