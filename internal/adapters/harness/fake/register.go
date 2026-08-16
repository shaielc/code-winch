package fake

import "github.com/shaielc/code-winch/internal/application"

func init() { application.DefaultDrivers.RegisterHarness(AdapterID, Driver{}) }
