package runner

import "time"

// MaxRecordBytes bounds the JSON encoding of every record the runner emits.
// Transports write, and readers accept, messages of at least this size, so a
// record is never split, dropped or refused for its size.
const MaxRecordBytes = 64 * 1024

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
