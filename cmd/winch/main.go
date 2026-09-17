// Command winch provides local operator and development commands.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	harnessfake "github.com/shaielc/code-winch/internal/adapters/harness/fake"
	sandboxlocal "github.com/shaielc/code-winch/internal/adapters/sandbox/local"
	"github.com/shaielc/code-winch/internal/application"
	runnerlocal "github.com/shaielc/code-winch/internal/runner/local"
	"github.com/shaielc/code-winch/pkg/protocol"
)

const devRunID = "00000000-0000-0000-0000-000000000001"

func main() {
	// An asked-for usage text is not a usage error: deployments/README.md opens
	// with `winch --help`, so it reports success on stdout rather than failure.
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help" || os.Args[1] == "help") {
		printUsage(os.Stdout)
		return
	}
	if len(os.Args) >= 3 && os.Args[1] == "dev" && os.Args[2] == "run" {
		devRun()
		return
	}
	if len(os.Args) >= 3 && os.Args[1] == "run" {
		switch os.Args[2] {
		case "create":
			runCreate()
			return
		case "get":
			runGet()
			return
		case "start":
			runStart()
			return
		case "events":
			runEvents()
			return
		}
	}
	printUsage(os.Stderr)
	os.Exit(2)
}

func printUsage(out io.Writer) {
	_, _ = fmt.Fprintln(out, "usage: winch run {create|get|start|events} | winch dev run")
}

func devRun() {
	fs := flag.NewFlagSet("dev run", flag.ExitOnError)
	harness := fs.String("harness", "fake", "harness driver")
	sandbox := fs.String("sandbox", "local", "sandbox driver")
	stopAfter := fs.Duration("stop-after", 0, "stop the harness after this duration")
	fakeBinary := fs.String("fake-binary", "", "path to the fake-harness executable (default: resolve on PATH)")
	fakeTranscript := fs.String("fake-transcript", "", "path to a scripted fake-harness transcript")
	fakeDelay := fs.Duration("fake-delay", 0, "delay each scripted fake-harness action")
	fakeFailure := fs.Bool("fake-force-failure", false, "make the fake harness exit unsuccessfully")
	fakeMalformed := fs.Bool("fake-malformed-line", false, "make the fake harness emit an invalid JSON-lines record")
	fakeEarlyExit := fs.Bool("fake-early-exit", false, "make the fake harness exit before reading interactive input")
	_ = fs.Parse(os.Args[3:])
	if *harness != "fake" || *sandbox != "local" {
		fmt.Fprintln(os.Stderr, "winch dev run: code=UNSUPPORTED_DRIVER")
		os.Exit(2)
	}
	ctx := context.Background()
	runner := runnerlocal.New(sandboxlocal.New(), harnessfake.Driver{Config: harnessfake.Config{
		Executable: *fakeBinary, Transcript: *fakeTranscript, Delay: *fakeDelay,
		ForceFailure: *fakeFailure, MalformedLine: *fakeMalformed, EarlyExit: *fakeEarlyExit,
	}})
	executionID := "dev-execution"
	lease := "dev-lease"
	var command atomic.Uint64
	send := func(kind string, payload any) error {
		n := command.Add(1)
		data, _ := json.Marshal(payload)
		return runner.Send(ctx, protocol.RunnerMessage{Version: protocol.RunnerVersion{Major: 1}, Kind: kind, CommandID: fmt.Sprintf("dev-%d", n), ExecutionID: executionID, LeaseToken: lease, Payload: data})
	}
	if err := send("prepare", protocol.PreparePayload{WorkspaceID: devRunID}); err != nil {
		fatal(err)
	}
	if err := send("start", protocol.StartPayload{LaunchProfile: "fake"}); err != nil {
		fatal(err)
	}
	if *stopAfter > 0 {
		go func() { time.Sleep(*stopAfter); _ = send("stop", protocol.StopPayload{GraceMilliseconds: 100}) }()
	}
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		i := 0
		for scanner.Scan() {
			i++
			if err := send("input", protocol.InputPayload{InputID: fmt.Sprintf("input-%d", i), Text: scanner.Text()}); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}()
	exitCode := 0
	for observation := range runner.Observations() {
		switch observation.Type {
		case "start":
			fmt.Println("[started]")
		case "output":
			printEvent(observation.Event)
		case "exit":
			fmt.Printf("[exit] successful=%t code=%s\n", observation.Exit.Successful, observation.Exit.Code)
			if !observation.Exit.Successful {
				exitCode = 1
			}
			_ = runner.Cleanup(ctx, executionID)
			// Close ends the loop by closing the channel, so the range drains
			// anything already queued instead of abandoning it.
			runner.Close()
		}
	}
	// An unsuccessful harness exit is the command's own result, so it reaches the
	// shell as one: `--fake-force-failure` and a failing transcript are only
	// observable controls if the exit status carries them.
	if exitCode != 0 {
		os.Exit(exitCode)
	}
}

type apiRun struct {
	ID             string `json:"id"`
	State          string `json:"state"`
	Version        int64  `json:"version"`
	LastSequence   int64  `json:"lastSequence"`
	WorkspacePath  string `json:"workspacePath"`
	HarnessProfile string `json:"harnessProfile"`
	SandboxProfile string `json:"sandboxProfile"`
}

func apiSettings() (string, string, string, string) {
	base := os.Getenv("WINCH_API_URL")
	if base == "" {
		base = "http://localhost:8080"
	}
	return strings.TrimRight(base, "/"), os.Getenv("WINCH_TOKEN"), os.Getenv("WINCH_CSRF_TOKEN"), base
}
func runCreate() {
	fs := flag.NewFlagSet("run create", flag.ExitOnError)
	workspace := fs.String("workspace", "", "workspace path")
	harness := fs.String("harness", "", "harness profile")
	sandbox := fs.String("sandbox", "", "sandbox profile")
	idempotencyKey := fs.String("idempotency-key", uuid.NewString(), "request idempotency key")
	_ = fs.Parse(os.Args[3:])
	body, _ := json.Marshal(map[string]string{"workspacePath": *workspace, "harnessProfile": *harness, "sandboxProfile": *sandbox})
	var run apiRun
	requestAPI(http.MethodPost, "/api/v1/runs", map[string]string{"Idempotency-Key": *idempotencyKey}, body, &run)
	fmt.Println(run.ID)
}
func runGet() {
	fs := flag.NewFlagSet("run get", flag.ExitOnError)
	_ = fs.Parse(os.Args[3:])
	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: winch run get RUN_ID")
		os.Exit(2)
	}
	var run apiRun
	requestAPI(http.MethodGet, "/api/v1/runs/"+fs.Arg(0), nil, nil, &run)
	data, _ := json.MarshalIndent(run, "", "  ")
	fmt.Println(string(data))
}

// runStart reads the run first because start is conditional: the daemon refuses
// a command that does not carry the run's current ETag, so the operator does
// not have to quote one by hand.
func runStart() {
	fs := flag.NewFlagSet("run start", flag.ExitOnError)
	idempotencyKey := fs.String("idempotency-key", uuid.NewString(), "request idempotency key")
	_ = fs.Parse(os.Args[3:])
	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: winch run start RUN_ID")
		os.Exit(2)
	}
	var current apiRun
	requestAPI(http.MethodGet, "/api/v1/runs/"+fs.Arg(0), nil, nil, &current)
	var started apiRun
	headers := map[string]string{"Idempotency-Key": *idempotencyKey, "If-Match": fmt.Sprintf("%q", strconv.FormatInt(current.Version, 10))}
	requestAPI(http.MethodPost, "/api/v1/runs/"+fs.Arg(0)+"/start", headers, nil, &started)
	data, _ := json.MarshalIndent(started, "", "  ")
	fmt.Println(string(data))
}

type apiEventPage struct {
	Events []struct {
		EventID     string          `json:"eventId"`
		Sequence    int64           `json:"sequence"`
		Kind        string          `json:"kind"`
		Sensitivity string          `json:"sensitivity"`
		OccurredAt  string          `json:"occurredAt"`
		Payload     json.RawMessage `json:"payload"`
	} `json:"events"`
	NextAfterSequence int64 `json:"nextAfterSequence"`
	HasMore           bool  `json:"hasMore"`
}

// runEvents polls the durable event page. It follows hasMore so one invocation
// prints the whole history rather than the first page of it.
func runEvents() {
	fs := flag.NewFlagSet("run events", flag.ExitOnError)
	after := fs.Int64("after-sequence", 0, "return events after this sequence")
	limit := fs.Int("limit", 50, "events per request")
	_ = fs.Parse(os.Args[3:])
	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: winch run events RUN_ID")
		os.Exit(2)
	}
	cursor := *after
	for {
		var page apiEventPage
		requestAPI(http.MethodGet, fmt.Sprintf("/api/v1/runs/%s/events?after_sequence=%d&limit=%d", fs.Arg(0), cursor, *limit), nil, nil, &page)
		for _, event := range page.Events {
			fmt.Printf("%d\t%s\t%s\t%s\n", event.Sequence, event.Kind, event.Sensitivity, event.Payload)
		}
		if !page.HasMore {
			return
		}
		cursor = page.NextAfterSequence
	}
}

func requestAPI(method, path string, headers map[string]string, body []byte, target any) {
	base, token, csrf, origin := apiSettings()
	req, err := http.NewRequest(method, base+path, strings.NewReader(string(body)))
	if err != nil {
		fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if method != http.MethodGet {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF-Token", csrf)
		req.Header.Set("Origin", origin)
	}
	for name, value := range headers {
		req.Header.Set(name, value)
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		fatal(err)
	}
	defer func() { _ = response.Body.Close() }()
	data, _ := io.ReadAll(response.Body)
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		fatal(fmt.Errorf("api status=%d: %s", response.StatusCode, strings.TrimSpace(string(data))))
	}
	if err = json.Unmarshal(data, target); err != nil {
		fatal(err)
	}
}
func printEvent(event *application.UnsequencedEvent) {
	var payload struct {
		Data string `json:"data"`
	}
	if json.Unmarshal(event.Payload, &payload) == nil && payload.Data != "" {
		fmt.Print(payload.Data)
	} else {
		fmt.Printf("[%s] %s\n", event.Kind, event.Payload)
	}
}
func fatal(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
