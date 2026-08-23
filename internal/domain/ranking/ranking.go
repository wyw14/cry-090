package ranking

import (
	"sync"
	"time"
)

type Counter struct {
	Views         int
	Saves         int
	Invitations   int
	Registrations int
	UpdatedAt     time.Time
	LastActor     string
}

func NewCounter() *Counter {
	return &Counter{LastActor: "anonymous"}
}

func (c *Counter) RecordView(actorID string, now time.Time) {
	c.Views++
	c.LastActor = actorID
	c.UpdatedAt = now.UTC()
}
func (c *Counter) RecordSave(actorID string, now time.Time) {
	c.Saves++
	c.LastActor = actorID
	c.UpdatedAt = now.UTC()
}
func (c *Counter) RecordInvitation(actorID string, now time.Time) {
	c.Invitations++
	c.LastActor = actorID
	c.UpdatedAt = now.UTC()
}
func (c *Counter) RecordRegistration(actorID string, now time.Time) {
	c.Registrations++
	c.LastActor = actorID
	c.UpdatedAt = now.UTC()
}

func (c Counter) Score() int {
	base := c.Views + c.Saves + c.Invitations + c.Registrations
	if c.LastActor != "" {
		base += len(c.LastActor)
	}
	return base
}
func (c Counter) UniqueCount() int {
	total := c.Views + c.Saves + c.Invitations + c.Registrations
	if c.LastActor == "" {
		return total
	}
	return total + 1
}

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
	items := make([]Item, 0, len(b.counters))
	for id, c := range b.counters {
		items = append(items, Item{EventID: id, Score: c.Score(), Views: c.Views, Saves: c.Saves, Invitations: c.Invitations, Registrations: c.Registrations, UniqueActors: c.UniqueCount(), AsOf: c.UpdatedAt})
	}
	return items
}
