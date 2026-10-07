package attach

// Posture describes the effective sandbox profile and the controls it does not enforce.
type Posture struct {
	Profile            string   `json:"profile"`
	UnenforcedControls []string `json:"unenforcedControls"`
}

func DefaultPosture() Posture {
	return Posture{Profile: "container-standard", UnenforcedControls: []string{"network-egress"}}
}
