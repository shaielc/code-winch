package local

import "github.com/shaielc/code-winch/internal/application"

func init() { application.DefaultDrivers.RegisterSandbox("local", New()) }
