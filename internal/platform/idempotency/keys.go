package idempotency

import (
	"context"
	"sync"
	"time"
)

type Entry struct {
	Response  []byte
	CreatedAt time.Time
}
type Store struct {
	mu      sync.Mutex
	entries map[string]Entry
	ttl     time.Duration
}

func New(ttl time.Duration) *Store { return &Store{entries: map[string]Entry{}, ttl: ttl} }
func (s *Store) Replay(_ context.Context, key string, now time.Time) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[key]
	if !ok || now.Sub(e.CreatedAt) > s.ttl {
		return nil, false
	}
	return append([]byte(nil), e.Response...), true
}
func (s *Store) Put(_ context.Context, key string, response []byte, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[key] = Entry{Response: append([]byte(nil), response...), CreatedAt: now}
}
