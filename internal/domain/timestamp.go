package domain

import (
	"errors"
	"time"
)

var errInvalidTimestamp = errors.New("invalid timestamp: expected a non-zero RFC3339 timestamp")

// Timestamp is an instant normalized to UTC. It deliberately exposes no
// wall-clock lookup; callers obtain current time through an injected Clock.
type Timestamp struct{ value time.Time }

func NewTimestamp(value time.Time) (Timestamp, error) {
	if value.IsZero() {
		return Timestamp{}, errInvalidTimestamp
	}
	return Timestamp{value: value.UTC()}, nil
}

func ParseTimestamp(value string) (Timestamp, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil || parsed.IsZero() {
		return Timestamp{}, errInvalidTimestamp
	}
	return NewTimestamp(parsed)
}

func (timestamp Timestamp) IsZero() bool    { return timestamp.value.IsZero() }
func (timestamp Timestamp) Time() time.Time { return timestamp.value }
func (timestamp Timestamp) String() string {
	if timestamp.IsZero() {
		return ""
	}
	return timestamp.value.Format(time.RFC3339Nano)
}
