// Command fake-harness is a controllable stand-in for a coding-agent CLI.
package main

import (
	"bufio"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	mathrand "math/rand"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/shaielc/code-winch/internal/adapters/harness/fake"
)

type record struct {
	Kind        string `json:"kind"`
	Payload     any    `json:"payload"`
	Sensitivity string `json:"sensitivity,omitempty"`
}
type streamPayload struct {
	Stream   string `json:"stream"`
	Encoding string `json:"encoding"`
	Data     string `json:"data"`
}
type command struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type options struct {
	latency                                     time.Duration
	failOnInput, truncateRecord, oversizedBytes int
	earlyExit                                   bool
	ignoreSIGTERM                               time.Duration
	rng                                         *mathrand.Rand
}

const usage = `fake-harness accepts JSON commands or bare text. Commands: help, echo <text>, fail, exit.`

func main() {
	runID := flag.String("run-id", "manual", "run identifier")
	scenarioPath := flag.String("scenario", "", "scenario JSON file")
	latency := flag.Duration("latency", 0, "latency added before every step")
	failInput := flag.Int("fail-on-input", 0, "exit unsuccessfully on the Nth input")
	truncate := flag.Int("truncate-record", 0, "truncate the Nth output record")
	oversized := flag.Int("oversized-record-bytes", 0, "pad every emitted record to at least this size")
	early := flag.Bool("early-exit", false, "exit without flushing after the first output")
	ignoreTerm := flag.Duration("ignore-sigterm", 0, "ignore SIGTERM for this bounded interval")
	seed := flag.String("seed", "1", "random seed, or off for nondeterministic seeding")
	flag.Parse()
	if *latency < 0 || *failInput < 0 || *truncate < 0 || *oversized < 0 || *ignoreTerm < 0 {
		fmt.Fprintln(os.Stderr, "fake harness: fault values must be non-negative")
		os.Exit(2)
	}
	rng, err := seeded(*seed)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	opt := options{*latency, *failInput, *truncate, *oversized, *early, *ignoreTerm, rng}
	if *scenarioPath != "" {
		scenario, err := fake.LoadScenario(*scenarioPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "fake harness:", err)
			os.Exit(2)
		}
		if configured := time.Duration(scenario.IgnoreSIGTERMMS) * time.Millisecond; configured > opt.ignoreSIGTERM {
			opt.ignoreSIGTERM = configured
		}
		installSignals(os.Stdout, opt.ignoreSIGTERM)
		os.Exit(runScenario(scenario, os.Stdin, os.Stdout, opt))
	}
	installSignals(os.Stdout, opt.ignoreSIGTERM)
	os.Exit(runBare(*runID, os.Stdin, os.Stdout, opt))
}

func seeded(value string) (*mathrand.Rand, error) {
	var seed int64
	if value == "off" {
		if err := binary.Read(rand.Reader, binary.LittleEndian, &seed); err != nil {
			return nil, fmt.Errorf("random seed: %w", err)
		}
	} else {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid --seed: use an integer or off")
		}
		seed = parsed
	}
	return mathrand.New(mathrand.NewSource(seed)), nil
}

func installSignals(out io.Writer, ignore time.Duration) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-signals
		if sig == syscall.SIGTERM && ignore > 0 {
			time.Sleep(ignore)
		}
		writeRecord(out, operational("fake harness received a termination signal"), 0)
		os.Exit(0)
	}()
}

func runScenario(s fake.Scenario, in io.Reader, out io.Writer, opt options) int {
	scanner := newScanner(in)
	inputs := 0
	records := 0
	for _, step := range s.Steps {
		if opt.latency > 0 {
			time.Sleep(opt.latency)
		}
		switch step.Kind {
		case "wait_input":
			re := regexp.MustCompile(step.Pattern)
			for {
				if !scanner.Scan() {
					return 0
				}
				inputs++
				if opt.failOnInput == inputs {
					return 1
				}
				text, ok := parse(scanner.Bytes())
				if ok && re.MatchString(text) {
					break
				}
			}
		case "sleep":
			time.Sleep(step.Duration())
		case "exit":
			return step.Code
		case "malformed":
			records++
			if writeBytes(out, []byte(expand(step.Data, opt.rng)), records, opt) {
				return 0
			}
		case "emit":
			var payload any
			_ = json.Unmarshal(step.Payload, &payload)
			payload = expandValue(payload, opt.rng)
			records++
			if writeRecordFault(out, record{Kind: step.EventKind, Payload: payload}, records, opt) {
				return 0
			}
		}
	}
	return 0
}

func runBare(runID string, in io.Reader, out io.Writer, opt options) int {
	writeRecord(out, operational(fmt.Sprintf("fake harness ready: run_id=%s", runID)), opt.oversizedBytes)
	writeRecord(out, operational(usage), opt.oversizedBytes)
	scanner := newScanner(in)
	inputs := 0
	for scanner.Scan() {
		inputs++
		if opt.failOnInput == inputs {
			return 1
		}
		text, ok := parse(scanner.Bytes())
		if !ok {
			writeRecord(out, diagnostic("FAKE_HARNESS_MALFORMED_COMMAND"), opt.oversizedBytes)
			continue
		}
		switch field, rest := split(text); field {
		case "":
			continue
		case "help":
			writeRecord(out, operational(usage), opt.oversizedBytes)
		case "fail":
			return 1
		case "exit", "quit":
			return 0
		case "echo":
			writeRecord(out, output(rest), opt.oversizedBytes)
		default:
			writeRecord(out, output(text), opt.oversizedBytes)
		}
	}
	if scanner.Err() != nil {
		return 1
	}
	return 0
}

func newScanner(in io.Reader) *bufio.Scanner {
	s := bufio.NewScanner(in)
	s.Buffer(make([]byte, 1024), 2<<20)
	return s
}
func split(text string) (string, string) {
	field, rest, _ := strings.Cut(strings.TrimSpace(text), " ")
	return field, strings.TrimSpace(rest)
}
func parse(line []byte) (string, bool) {
	trimmed := strings.TrimRight(string(line), "\r")
	if !strings.HasPrefix(strings.TrimSpace(trimmed), "{") {
		return trimmed, true
	}
	var v command
	if json.Unmarshal([]byte(trimmed), &v) != nil || v.ID == "" {
		return "", false
	}
	return v.Text, true
}
func output(text string) record {
	return record{"stream.raw", streamPayload{"stdout", "utf-8", text + "\r\n"}, ""}
}
func operational(text string) record { v := output(text); v.Sensitivity = "operational"; return v }
func diagnostic(code string) record {
	return record{"diagnostic", map[string]string{"code": code, "message": "fake harness could not interpret a command"}, "operational"}
}
func expand(s string, rng *mathrand.Rand) string {
	return strings.ReplaceAll(s, "{{random}}", fmt.Sprintf("%016x", rng.Uint64()))
}
func expandValue(v any, rng *mathrand.Rand) any {
	switch x := v.(type) {
	case string:
		return expand(x, rng)
	case []any:
		for i := range x {
			x[i] = expandValue(x[i], rng)
		}
	case map[string]any:
		for k := range x {
			x[k] = expandValue(x[k], rng)
		}
	}
	return v
}
func writeRecord(out io.Writer, value record, oversized int) {
	data, _ := json.Marshal(value)
	if oversized > len(data) {
		data = append(data, bytesOf(' ', oversized-len(data))...)
	}
	_, _ = out.Write(append(data, '\n'))
}
func writeRecordFault(out io.Writer, value record, n int, opt options) bool {
	data, _ := json.Marshal(value)
	return writeBytes(out, data, n, opt)
}
func writeBytes(out io.Writer, data []byte, n int, opt options) bool {
	if opt.oversizedBytes > len(data) {
		data = append(data, bytesOf(' ', opt.oversizedBytes-len(data))...)
	}
	if opt.truncateRecord == n && len(data) > 1 {
		data = data[:len(data)/2]
	}
	_, _ = out.Write(append(data, '\n'))
	return opt.earlyExit
}
func bytesOf(b byte, n int) []byte {
	result := make([]byte, n)
	for i := range result {
		result[i] = b
	}
	return result
}
