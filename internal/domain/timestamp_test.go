package domain

import (
	"testing"
	"time"
)

func TestTimestampNormalizesToUTC(t *testing.T) {
	timestamp, err := ParseTimestamp("2026-08-11T12:30:45.123+02:00")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := timestamp.String(), "2026-08-11T10:30:45.123Z"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
	if timestamp.Time().Location() != time.UTC {
		t.Fatalf("location = %v, want UTC", timestamp.Time().Location())
	}
}

func TestTimestampRejectsInvalidAndZeroValues(t *testing.T) {
	if _, err := ParseTimestamp("yesterday"); err == nil {
		t.Fatal("invalid timestamp was accepted")
	}
	if _, err := NewTimestamp(time.Time{}); err == nil {
		t.Fatal("zero timestamp was accepted")
	}
	if !(Timestamp{}).IsZero() {
		t.Fatal("zero Timestamp was accepted")
	}
}
