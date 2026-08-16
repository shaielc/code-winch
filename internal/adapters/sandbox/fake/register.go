package fake

import "github.com/shaielc/code-winch/internal/application"

func init() {
	application.DefaultDrivers.RegisterSandbox("fake", New(application.SandboxCapabilities{Isolation: "in-memory", Attach: true}))
}
