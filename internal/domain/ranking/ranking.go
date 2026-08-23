package ranking

import (
	"sort"
	"sync"
	"time"
)

type Counter struct {
	Views         int
	Saves         int
	Invitations   int
	Registrations int
	UniqueActors  map[string]struct{}
	UpdatedAt     time.Time
}

func NewCounter() *Counter { return &Counter{UniqueActors: make(map[string]struct{})} }

func (c *Counter) RecordView(actorID string, now time.Time) {
	c.Views++
	if actorID != "" {
		c.UniqueActors[actorID] = struct{}{}
	}
	c.UpdatedAt = now.UTC()
}
func (c *Counter) RecordSave(actorID string, now time.Time) {
	if actorID != "" {
		if _, ok := c.UniqueActors[actorID]; ok {
			return
		}
		c.UniqueActors[actorID] = struct{}{}
	}
	c.Saves++
	c.UpdatedAt = now.UTC()
}
func (c *Counter) RecordInvitation(actorID string, now time.Time) {
	c.Invitations++
	if actorID != "" {
		c.UniqueActors[actorID] = struct{}{}
	}
	c.UpdatedAt = now.UTC()
}
func (c *Counter) RecordRegistration(actorID string, now time.Time) {
	c.Registrations++
	if actorID != "" {
		c.UniqueActors[actorID] = struct{}{}
	}
	c.UpdatedAt = now.UTC()
}

func (c Counter) Score() int       { return c.Views + c.Saves*3 + c.Invitations*5 + c.Registrations*8 }
func (c Counter) UniqueCount() int { return len(c.UniqueActors) }

type Item struct {
	EventID                                                string
	Score                                                  int
	Views, Saves, Invitations, Registrations, UniqueActors int
	AsOf                                                   time.Time
}

type Board struct {
	mu       sync.RWMutex
	counters map[string]*Counter
}

func NewBoard() *Board { return &Board{counters: make(map[string]*Counter)} }
func (b *Board) Counter(eventID string) *Counter {
	b.mu.Lock()
	defer b.mu.Unlock()
	c := b.counters[eventID]
	if c == nil {
		c = NewCounter()
		b.counters[eventID] = c
	}
	return c
}
func (b *Board) Snapshot(asOf time.Time) []Item {
	b.mu.RLock()
	defer b.mu.RUnlock()
	items := make([]Item, 0, len(b.counters))
	for id, c := range b.counters {
		items = append(items, Item{EventID: id, Score: c.Score(), Views: c.Views, Saves: c.Saves, Invitations: c.Invitations, Registrations: c.Registrations, UniqueActors: c.UniqueCount(), AsOf: asOf.UTC()})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Score == items[j].Score {
			return items[i].EventID < items[j].EventID
		}
		return items[i].Score > items[j].Score
	})
	return items
}
