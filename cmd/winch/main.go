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
	"strings"
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
			fmt.Fprintln(os.Stderr, "winch run start: not implemented; owner=P0-008")
			os.Exit(1)
		}
	}
	fmt.Fprintln(os.Stderr, "usage: winch run {create|get|start} | winch dev run")
	os.Exit(2)
}

func devRun() {
	fs := flag.NewFlagSet("dev run", flag.ExitOnError)
	harness := fs.String("harness", "fake", "harness driver")
	sandbox := fs.String("sandbox", "local", "sandbox driver")
	stopAfter := fs.Duration("stop-after", 0, "stop the harness after this duration")
	_ = fs.Parse(os.Args[3:])
	if *harness != "fake" || *sandbox != "local" {
		fmt.Fprintln(os.Stderr, "winch dev run: code=UNSUPPORTED_DRIVER")
		os.Exit(2)
	}
	ctx := context.Background()
	runner := runnerlocal.New(sandboxlocal.New(), harnessfake.Driver{})
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
	for observation := range runner.Observations() {
		switch observation.Type {
		case "start":
			fmt.Println("[started]")
		case "output":
			printEvent(observation.Event)
		case "exit":
			fmt.Printf("[exit] successful=%t code=%s\n", observation.Exit.Successful, observation.Exit.Code)
			_ = runner.Cleanup(ctx, executionID)
			// Close ends the loop by closing the channel, so the range drains
			// anything already queued instead of abandoning it.
			runner.Close()
		}
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
	_ = fs.Parse(os.Args[3:])
	body, _ := json.Marshal(map[string]string{"workspacePath": *workspace, "harnessProfile": *harness, "sandboxProfile": *sandbox})
	var run apiRun
	requestAPI(http.MethodPost, "/api/v1/runs", body, &run)
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
	requestAPI(http.MethodGet, "/api/v1/runs/"+fs.Arg(0), nil, &run)
	data, _ := json.MarshalIndent(run, "", "  ")
	fmt.Println(string(data))
}
func requestAPI(method, path string, body []byte, target any) {
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
