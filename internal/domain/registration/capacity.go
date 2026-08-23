package registration

import (
	"sort"
	"sync"
	"time"

	"github.com/wyw14/cry-090/internal/domain/common"
)

type CapacityBook struct {
	mu            sync.Mutex
	capacity      int64
	registered    int64
	registrations map[string]*Registration
}

func NewCapacityBook(capacity int) (*CapacityBook, error) {
	if capacity < 1 {
		return nil, common.FieldError("capacity", "must be positive")
	}
	return &CapacityBook{capacity: int64(capacity), registrations: make(map[string]*Registration)}, nil
}

// Register admits reg against the book's capacity. The duplicate check, the
// active-count comparison, the registered-vs-waitlisted decision, position
// assignment, and the map insertion all run under a single lock, so concurrent
// callers can never each read the same count and all claim the last seat.
func (b *CapacityBook) Register(reg *Registration) (Status, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	existing, exists := b.registrations[reg.UserID]
	if exists && existing.Status != StatusCancelled {
		return "", common.NewError(common.CodeConflict, "user already has an active registration")
	}
	if b.registered < b.capacity {
		reg.Status = StatusRegistered
		reg.Position = 0
		b.registered++
	} else {
		reg.Status = StatusWaitlisted
		reg.Position = b.nextWaitlistPositionLocked()
	}
	b.registrations[reg.UserID] = reg
	return reg.Status, nil
}

func (b *CapacityBook) CancelAt(userID string, expectedVersion int64, now func() time.Time) (*Registration, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	reg, ok := b.registrations[userID]
	if !ok {
		return nil, common.NewError(common.CodeNotFound, "registration not found")
	}
	if err := reg.Cancel(expectedVersion, now()); err != nil {
		return nil, err
	}
	if reg.Status == StatusRegistered {
		b.registered--
	}
	return reg, nil
}

func (b *CapacityBook) Snapshot() []*Registration {
	b.mu.Lock()
	defer b.mu.Unlock()
	result := make([]*Registration, 0, len(b.registrations))
	for _, reg := range b.registrations {
		copy := *reg
		result = append(result, &copy)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].UserID < result[j].UserID })
	return result
}

func (b *CapacityBook) activeCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return int(b.registered)
}

// nextWaitlistPositionLocked returns the next 1-based waitlist position. It is
// computed from the current waitlisted registrations rather than a stale count,
// so concurrent waitlisters never receive the same position.
func (b *CapacityBook) nextWaitlistPositionLocked() int {
	max := 0
	for _, reg := range b.registrations {
		if reg.Status == StatusWaitlisted && reg.Position > max {
			max = reg.Position
		}
	}
	return max + 1
}

func (b *CapacityBook) promoteNext(now time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	var candidate *Registration
	for _, reg := range b.registrations {
		if reg.Status == StatusWaitlisted && (candidate == nil || reg.Position < candidate.Position) {
			candidate = reg
		}
	}
	if candidate == nil || b.registered >= b.capacity {
		return
	}
	candidate.Status = StatusRegistered
	candidate.Position = 0
	candidate.Version++
	b.registered++
	_ = now
}
