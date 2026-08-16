// Package sandbox contains the reusable sandbox adapter contract suite.
package sandbox

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/shaielc/code-winch/internal/application"
)

type Factory func(t *testing.T) application.SandboxDriver

// Run checks the complete lifecycle, repeated stop/cleanup, and every optional
// control the implementation advertises.
func Run(t *testing.T, factory Factory) {
	t.Helper()
	if err := Validate(factory(t)); err != nil {
		t.Fatal(err)
	}
}

// Validate is the non-testing form used to assert that deliberately dishonest
// test doubles are rejected by this same suite.
func Validate(driver application.SandboxDriver) error {
	ctx := context.Background()
	caps := driver.Capabilities(ctx)
	spec := application.SandboxSpec{ID: "contract-sandbox", DisableNetwork: caps.NetworkPolicy}
	if caps.ResourceLimits {
		spec.MemoryBytes = 1024
	}
	prepared, err := driver.Prepare(ctx, spec)
	if err != nil {
		return fmt.Errorf("advertised capability is not usable: %w", err)
	}
	handle, err := driver.Start(ctx, prepared, application.LaunchSpec{Command: "contract-command"})
	if err != nil {
		return fmt.Errorf("start: %w", err)
	}
	var stream io.ReadWriteCloser
	if caps.Attach {
		stream, err = driver.Attach(ctx, handle)
		if err != nil || stream == nil {
			return fmt.Errorf("advertised attach is not usable: stream=%v: %w", stream, err)
		}
		if caps.AttachSingleUse {
			if second, secondErr := driver.Attach(ctx, handle); secondErr == nil || second != nil {
				return fmt.Errorf("single-use attach accepted a second attachment")
			}
		}
		payload := []byte("contract read-after-write\n")
		if _, err = stream.Write(payload); err != nil {
			return fmt.Errorf("attached write: %w", err)
		}
		setReadDeadline(stream, time.Now().Add(time.Second))
		got := make([]byte, len(payload)+1) // PTYs commonly translate LF to CRLF.
		n, readErr := stream.Read(got)
		if readErr != nil || strings.TrimSpace(string(got[:n])) != strings.TrimSpace(string(payload)) {
			return fmt.Errorf("attached read-after-write: %q: %v", got, err)
		}
	}
	observed, err := driver.Inspect(ctx, handle)
	if err != nil || !observed.Running {
		return fmt.Errorf("inspect running: %#v: %v", observed, err)
	}
	if caps.Resize {
		if err = driver.Resize(ctx, handle, application.TerminalSize{Rows: 31, Cols: 97}); err != nil {
			return fmt.Errorf("advertised resize is not usable: %w", err)
		}
	}
	for i := 0; i < 2; i++ {
		if err = driver.Stop(ctx, handle, application.StopPolicy{}); err != nil {
			return fmt.Errorf("stop #%d: %w", i+1, err)
		}
	}
	observed, err = driver.Inspect(ctx, handle)
	if err != nil || observed.Running || observed.ExitCode == nil {
		return fmt.Errorf("inspect stopped: %#v: %v", observed, err)
	}
	if stream != nil {
		setReadDeadline(stream, time.Now().Add(time.Second))
		var one [1]byte
		if _, readErr := stream.Read(one[:]); readErr == nil {
			return fmt.Errorf("attached read after exit did not terminate")
		}
		if _, writeErr := stream.Write(one[:]); writeErr == nil {
			return fmt.Errorf("attached write after exit succeeded")
		}
		if err = stream.Close(); err != nil {
			return fmt.Errorf("attached close: %w", err)
		}
		_ = stream.Close() // repeated close must be safe, but may report closed
	}
	for i := 0; i < 2; i++ {
		if err = driver.Cleanup(ctx, prepared); err != nil {
			return fmt.Errorf("cleanup #%d: %w", i+1, err)
		}
	}
	return nil
}

func setReadDeadline(stream io.ReadWriteCloser, deadline time.Time) {
	if value, ok := stream.(interface{ SetReadDeadline(time.Time) error }); ok {
		_ = value.SetReadDeadline(deadline)
	}
}
