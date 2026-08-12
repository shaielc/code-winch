package fake_test

import (
	"strings"
	"testing"

	"github.com/shaielc/code-winch/internal/adapters/sandbox/fake"
	"github.com/shaielc/code-winch/internal/application"
	sandboxcontract "github.com/shaielc/code-winch/test/contract/sandbox"
)

func TestSandboxContract(t *testing.T) {
	sandboxcontract.Run(t, func(*testing.T) application.SandboxDriver {
		return fake.New(application.SandboxCapabilities{NetworkPolicy: true, ResourceLimits: true})
	})
}

func TestContractRejectsLyingCapability(t *testing.T) {
	driver := fake.New(application.SandboxCapabilities{NetworkPolicy: true})
	driver.RejectNetworkPolicy = true
	err := sandboxcontract.Validate(driver)
	if err == nil || !strings.Contains(err.Error(), "advertised capability is not usable") {
		t.Fatalf("lying adapter passed contract: %v", err)
	}
}
