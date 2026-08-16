package application

import (
	"fmt"
	"sort"
	"sync"
)

// DriverRegistry is populated by adapter-owned registration files. Registration
// is append-only: duplicate names panic at process startup rather than silently
// replacing a security-sensitive execution adapter.
type DriverRegistry struct {
	mu      sync.RWMutex
	harness map[string]HarnessDriver
	sandbox map[string]SandboxDriver
}

func NewDriverRegistry() *DriverRegistry {
	return &DriverRegistry{harness: map[string]HarnessDriver{}, sandbox: map[string]SandboxDriver{}}
}
func (r *DriverRegistry) RegisterHarness(name string, d HarnessDriver) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if name == "" || d == nil || r.harness[name] != nil {
		panic("application: duplicate or invalid harness driver: " + name)
	}
	r.harness[name] = d
}
func (r *DriverRegistry) RegisterSandbox(name string, d SandboxDriver) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if name == "" || d == nil || r.sandbox[name] != nil {
		panic("application: duplicate or invalid sandbox driver: " + name)
	}
	r.sandbox[name] = d
}
func (r *DriverRegistry) Harness(name string) (HarnessDriver, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d := r.harness[name]
	if d == nil {
		return nil, fmt.Errorf("%w: harness_profile", ErrUnknownProfile)
	}
	return d, nil
}
func (r *DriverRegistry) Sandbox(name string) (SandboxDriver, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d := r.sandbox[name]
	if d == nil {
		return nil, fmt.Errorf("%w: sandbox_profile", ErrUnknownProfile)
	}
	return d, nil
}
func (r *DriverRegistry) Names() (harness, sandbox []string) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for n := range r.harness {
		harness = append(harness, n)
	}
	for n := range r.sandbox {
		sandbox = append(sandbox, n)
	}
	sort.Strings(harness)
	sort.Strings(sandbox)
	return
}

var ErrUnknownProfile = fmt.Errorf("unknown driver profile")

// DefaultDrivers is the process registry populated by adapter init functions.
var DefaultDrivers = NewDriverRegistry()
