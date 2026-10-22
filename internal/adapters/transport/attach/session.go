package attach

// Posture describes the effective sandbox profile and the controls it does not enforce.
type Posture struct {
	Profile            string   `json:"profile"`
	UnenforcedControls []string `json:"unenforcedControls"`
}

// ChildStatus describes the sandbox's observation of its owned child process.
type ChildStatus string

const (
	// ChildDown means the sandbox composition owns no child process.
	ChildDown ChildStatus = "Down"
)

// Session combines the sandbox's effective posture with its child observation.
type Session struct {
	Posture
	ChildStatus ChildStatus `json:"childStatus"`
}

func DefaultPosture() Posture {
	return Posture{Profile: "container-standard", UnenforcedControls: []string{"network-egress"}}
}
