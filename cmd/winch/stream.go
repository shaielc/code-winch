package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/coder/websocket"
	"github.com/shaielc/code-winch/internal/runner"
)

func runStream(ctx context.Context, args []string, output io.Writer) error {
	flags := flag.NewFlagSet("stream", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	baseURL := flags.String("url", "http://127.0.0.1:8080", "sandbox URL")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 {
		return errors.New("invalid stream arguments")
	}
	endpoint, err := url.Parse(strings.TrimRight(*baseURL, "/") + "/api/session/stream")
	if err != nil {
		return errors.New("invalid sandbox URL")
	}
	if endpoint.Scheme == "https" {
		endpoint.Scheme = "wss"
	} else {
		endpoint.Scheme = "ws"
	}
	connection, _, err := websocket.Dial(ctx, endpoint.String(), nil)
	if err != nil {
		return fmt.Errorf("connect to sandbox stream: %w", err)
	}
	defer func() { _ = connection.CloseNow() }()
	// The library default of 32 KiB is below the largest record the stream sends.
	connection.SetReadLimit(runner.MaxRecordBytes)
	for {
		_, data, err := connection.Read(ctx)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure || ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("read sandbox stream: %w", err)
		}
		if _, err := fmt.Fprintln(output, string(data)); err != nil {
			return err
		}
	}
}
