// Command fake-harness is a deterministic stand-in for a coding-agent CLI. It
// speaks the newline-delimited JSON dialect that the fake harness adapter
// encodes and decodes, so a run can be driven end to end without a vendor
// account. It is a development and test fixture, not a production harness.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// record mirrors the fake harness codec's wire format. The codec rejects
// unknown fields, so this struct must stay a subset of its own.
type record struct {
	Kind        string `json:"kind"`
	Payload     any    `json:"payload"`
	Sensitivity string `json:"sensitivity,omitempty"`
}

// streamPayload matches schemas/events/v1/fixtures/raw-stream.json. Text is
// carried as utf-8 rather than base64 so the terminal projection can render it
// without decoding.
type streamPayload struct {
	Stream   string `json:"stream"`
	Encoding string `json:"encoding"`
	Data     string `json:"data"`
}

// command is one line encoded by the fake harness codec's Encode.
type command struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

const usage = `fake-harness accepts one JSON command per line: {"id":"<id>","text":"<text>"}
  help          print this message
  echo <text>   emit <text> back as terminal output
  fail          exit with a nonzero status
  exit, quit    exit successfully
Any other text is echoed.`

func main() {
	runID := flag.String("run-id", "", "run identifier supplied by the harness adapter")
	flag.Parse()
	if *runID == "" {
		fmt.Fprintln(os.Stderr, "fake harness: code=INVALID_RUN field=run_id")
		os.Exit(2)
	}
	os.Exit(run(*runID, os.Stdin, os.Stdout))
}

func run(runID string, in *os.File, out *os.File) int {
	// A terminated harness must still produce a final observation, otherwise a
	// stop looks identical to a crash from the run's event history.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-signals
		emit(out, operational("fake harness received a termination signal"))
		os.Exit(0)
	}()

	emit(out, operational(fmt.Sprintf("fake harness ready: run_id=%s", runID)))
	emit(out, operational(usage))

	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 0, 64*1024), 1<<20)
	for scanner.Scan() {
		text, ok := parse(scanner.Bytes())
		if !ok {
			emit(out, diagnostic("FAKE_HARNESS_MALFORMED_COMMAND"))
			continue
		}
		code, done := respond(out, text)
		if done {
			return code
		}
	}
	if err := scanner.Err(); err != nil && !errors.Is(err, os.ErrClosed) {
		emit(out, diagnostic("FAKE_HARNESS_READ_FAILED"))
		return 1
	}
	emit(out, operational("fake harness reached end of input"))
	return 0
}

// respond reports the exit code and whether the harness should stop reading.
func respond(out *os.File, text string) (int, bool) {
	switch field, rest := split(text); field {
	case "":
		return 0, false
	case "help":
		emit(out, operational(usage))
	case "fail":
		emit(out, operational("fake harness exiting with a failure"))
		return 1, true
	case "exit", "quit":
		emit(out, operational("fake harness exiting"))
		return 0, true
	case "echo":
		emit(out, output(rest))
	default:
		emit(out, output(text))
	}
	return 0, false
}

func split(text string) (string, string) {
	field, rest, _ := strings.Cut(strings.TrimSpace(text), " ")
	return field, strings.TrimSpace(rest)
}

// parse accepts the codec's JSON command and also bare text, so the binary
// stays usable when a human is attached directly to its PTY.
func parse(line []byte) (string, bool) {
	trimmed := strings.TrimRight(string(line), "\r")
	if !strings.HasPrefix(strings.TrimSpace(trimmed), "{") {
		return trimmed, true
	}
	var value command
	if json.Unmarshal([]byte(trimmed), &value) != nil || value.ID == "" {
		return "", false
	}
	return value.Text, true
}

func output(text string) record {
	return record{Kind: "stream.raw", Payload: streamPayload{Stream: "stdout", Encoding: "utf-8", Data: text + "\r\n"}}
}

func operational(text string) record {
	value := output(text)
	value.Sensitivity = "operational"
	return value
}

func diagnostic(code string) record {
	return record{Kind: "diagnostic", Payload: map[string]string{"code": code, "message": "fake harness could not interpret a command"}, Sensitivity: "operational"}
}

// emit writes one record per line. Write failures are unrecoverable here: the
// harness has no other channel on which to report them.
func emit(out *os.File, value record) {
	data, err := json.Marshal(value)
	if err != nil {
		return
	}
	_, _ = out.Write(append(data, '\n'))
}
