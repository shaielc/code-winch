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
	// ChildRunning means an owned child started and its exit has not been observed.
	ChildRunning ChildStatus = "Running"
	// ChildExited means the owned child's exit has been observed.
	ChildExited ChildStatus = "Exited"
)

// Session combines the sandbox's effective posture with its child observation.
type Session struct {
	Posture
	ChildStatus       ChildStatus        `json:"childStatus"`
	ChildStatusSource func() ChildStatus `json:"-"`
}

func (s Session) current() Session {
	if s.ChildStatusSource != nil {
		s.ChildStatus = s.ChildStatusSource()
	}
	return s
}

func DefaultPosture() Posture {
	return Posture{Profile: "container-standard", UnenforcedControls: []string{"network-egress"}}
}
