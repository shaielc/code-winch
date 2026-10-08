package runner

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"os/exec"
	"syscall"
	"unicode/utf8"
)

const readBufferSize = 32 * 1024

type HarnessConfig struct {
	Executable string
	Args       []string
}

// RunHarness starts and owns one harness process until it exits.
func RunHarness(ctx context.Context, session *Session, config HarnessConfig) error {
	command := exec.CommandContext(ctx, config.Executable, config.Args...)
	command.Cancel = func() error { return command.Process.Signal(syscall.SIGTERM) }
	stdout, err := command.StdoutPipe()
	if err != nil {
		return err
	}
	stdin, err := command.StdinPipe()
	if err != nil {
		return err
	}
	if err := command.Start(); err != nil {
		return err
	}
	// Input arrives in a later stage. Keeping stdin open keeps an interactive
	// harness alive while its output is observed.
	defer func() { _ = stdin.Close() }()
	readErr := pump(stdout, session)
	waitErr := command.Wait()
	appendTermination(session, waitErr)
	if readErr != nil {
		return readErr
	}
	return nil
}

func pump(source io.Reader, session *Session) error {
	buffer := make([]byte, readBufferSize)
	for {
		n, err := source.Read(buffer)
		if n > 0 {
			chunk := buffer[:n]
			encoding, data := "utf-8", string(chunk)
			if !utf8.Valid(chunk) {
				encoding, data = "base64", base64.StdEncoding.EncodeToString(chunk)
			}
			session.append("stream.raw", "user-content", StreamPayload{Stream: "stdout", Encoding: encoding, Data: data})
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
	}
}

func appendTermination(session *Session, waitErr error) {
	payload := TerminatedPayload{Outcome: "completed"}
	if waitErr != nil {
		payload.Outcome = "failed"
		var exitError *exec.ExitError
		if errors.As(waitErr, &exitError) {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok && status.Signaled() {
				payload.Outcome = "stopped"
				payload.Signal = status.Signal().String()
			} else {
				code := exitError.ExitCode()
				payload.ExitCode = &code
			}
		}
	} else {
		code := 0
		payload.ExitCode = &code
	}
	session.append("session.terminated", "operational", payload)
}
