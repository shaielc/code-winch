// Package fake contains an in-memory sandbox test double. It enforces rather
// than merely advertises optional capabilities.
package fake

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/shaielc/code-winch/internal/application"
)

type Driver struct {
	CapabilitiesValue application.SandboxCapabilities
	// RejectNetworkPolicy and RejectResourceLimits intentionally allow contract
	// tests to prove that an implementation lying about support is rejected.
	RejectNetworkPolicy  bool
	RejectResourceLimits bool
	mu                   sync.Mutex
	prepared             map[string]bool
	executions           map[string]application.ObservedExecution
}

func New(capabilities application.SandboxCapabilities) *Driver {
	return &Driver{CapabilitiesValue: capabilities, prepared: map[string]bool{}, executions: map[string]application.ObservedExecution{}}
}
func (d *Driver) Capabilities(context.Context) application.SandboxCapabilities {
	return d.CapabilitiesValue
}
func (d *Driver) Prepare(_ context.Context, spec application.SandboxSpec) (application.PreparedSandbox, error) {
	if spec.ID == "" {
		return application.PreparedSandbox{}, errors.New("fake sandbox: code=INVALID_SPEC field=sandbox_id")
	}
	if spec.DisableNetwork && (!d.CapabilitiesValue.NetworkPolicy || d.RejectNetworkPolicy) {
		return application.PreparedSandbox{}, fmt.Errorf("fake sandbox: code=UNSUPPORTED_CAPABILITY sandbox_id=%s capability=network_policy", spec.ID)
	}
	if spec.MemoryBytes > 0 && (!d.CapabilitiesValue.ResourceLimits || d.RejectResourceLimits) {
		return application.PreparedSandbox{}, fmt.Errorf("fake sandbox: code=UNSUPPORTED_CAPABILITY sandbox_id=%s capability=resource_limits", spec.ID)
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.prepared[spec.ID] = true
	return application.PreparedSandbox{ID: spec.ID}, nil
}
func (d *Driver) Start(_ context.Context, prepared application.PreparedSandbox, _ application.LaunchSpec) (application.ExecutionHandle, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.prepared[prepared.ID] {
		return application.ExecutionHandle{}, fmt.Errorf("fake sandbox: code=NOT_PREPARED sandbox_id=%s", prepared.ID)
	}
	h := application.ExecutionHandle{ID: "execution-" + prepared.ID}
	d.executions[h.ID] = application.ObservedExecution{Running: true}
	return h, nil
}
func (d *Driver) Inspect(_ context.Context, handle application.ExecutionHandle) (application.ObservedExecution, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	value, ok := d.executions[handle.ID]
	if !ok {
		return application.ObservedExecution{}, fmt.Errorf("fake sandbox: code=NOT_FOUND execution_id=%s", handle.ID)
	}
	return value, nil
}
func (d *Driver) Resize(_ context.Context, handle application.ExecutionHandle, size application.TerminalSize) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.CapabilitiesValue.Resize {
		return errors.New("fake sandbox: code=UNSUPPORTED_CAPABILITY capability=resize")
	}
	if _, ok := d.executions[handle.ID]; !ok {
		return fmt.Errorf("fake sandbox: code=NOT_FOUND execution_id=%s", handle.ID)
	}
	if size.Rows == 0 || size.Cols == 0 {
		return errors.New("fake sandbox: code=INVALID_SIZE")
	}
	return nil
}
func (d *Driver) Stop(_ context.Context, handle application.ExecutionHandle, _ application.StopPolicy) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	value, ok := d.executions[handle.ID]
	if !ok {
		return nil
	}
	code := 0
	value.Running = false
	value.ExitCode = &code
	d.executions[handle.ID] = value
	return nil
}
func (d *Driver) Cleanup(_ context.Context, prepared application.PreparedSandbox) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.prepared, prepared.ID)
	delete(d.executions, "execution-"+prepared.ID)
	return nil
}

var _ application.SandboxDriver = (*Driver)(nil)
