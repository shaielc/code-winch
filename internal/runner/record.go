package runner

import "time"

// Record is one runner-local observation of the sandbox session.
type Record struct {
	Ordinal     uint64    `json:"ordinal"`
	Kind        string    `json:"kind"`
	OccurredAt  time.Time `json:"occurredAt"`
	Sensitivity string    `json:"sensitivity"`
	Payload     any       `json:"payload"`
}

type StreamPayload struct {
	Stream   string `json:"stream"`
	Encoding string `json:"encoding"`
	Data     string `json:"data"`
}

type TerminatedPayload struct {
	Outcome  string `json:"outcome"`
	ExitCode *int   `json:"exitCode,omitempty"`
	Signal   string `json:"signal,omitempty"`
}
