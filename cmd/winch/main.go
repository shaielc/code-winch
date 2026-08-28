// Command winch provides local operator and development commands.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	harnessfake "github.com/shaielc/code-winch/internal/adapters/harness/fake"
	sandboxlocal "github.com/shaielc/code-winch/internal/adapters/sandbox/local"
	"github.com/shaielc/code-winch/internal/application"
	runnerlocal "github.com/shaielc/code-winch/internal/runner/local"
	"github.com/shaielc/code-winch/pkg/protocol"
)

const devRunID = "00000000-0000-0000-0000-000000000001"

func main() {
	if len(os.Args) < 3 || os.Args[1] != "dev" || os.Args[2] != "run" {
		fmt.Fprintln(os.Stderr, "usage: winch dev run --harness fake --sandbox local [fake controls]")
		os.Exit(2)
	}
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
	os.Exit(exitCode)
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
