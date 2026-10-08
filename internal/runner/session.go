package runner

import (
	"sync"
	"time"
)

// Session assigns local ordinals and fans records out to all attached readers.
type Session struct {
	mu          sync.Mutex
	records     []Record
	subscribers map[chan Record]struct{}
	now         func() time.Time
}

func NewSession() *Session {
	return &Session{subscribers: make(map[chan Record]struct{}), now: time.Now}
}

func (s *Session) append(kind, sensitivity string, payload any) Record {
	s.mu.Lock()
	defer s.mu.Unlock()
	record := Record{Ordinal: uint64(len(s.records) + 1), Kind: kind, OccurredAt: s.now().UTC(), Sensitivity: sensitivity, Payload: payload}
	s.records = append(s.records, record)
	for subscriber := range s.subscribers {
		subscriber <- record
	}
	return record
}

// Subscribe returns the current snapshot followed by every newly appended record.
func (s *Session) Subscribe() (<-chan Record, func()) {
	s.mu.Lock()
	channel := make(chan Record, len(s.records)+256)
	for _, record := range s.records {
		channel <- record
	}
	s.subscribers[channel] = struct{}{}
	s.mu.Unlock()
	return channel, func() {
		s.mu.Lock()
		delete(s.subscribers, channel)
		s.mu.Unlock()
	}
}
