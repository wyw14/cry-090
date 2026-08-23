package registration

import (
	"sort"
	"sync"
	"time"

	"github.com/wyw14/cry-090/internal/domain/common"
)

type CapacityBook struct {
	mu            sync.Mutex
	capacity      int
	registrations map[string]*Registration
}

func NewCapacityBook(capacity int) (*CapacityBook, error) {
	if capacity < 1 {
		return nil, common.FieldError("capacity", "must be positive")
	}
	return &CapacityBook{capacity: capacity, registrations: make(map[string]*Registration)}, nil
}

func (b *CapacityBook) Register(reg *Registration) (Status, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if existing, ok := b.registrations[reg.UserID]; ok && existing.Status != StatusCancelled {
		return "", common.NewError(common.CodeConflict, "user already has an active registration")
	}
	if b.activeCount() < b.capacity {
		reg.Status = StatusRegistered
		reg.Position = 0
	} else {
		reg.Status = StatusWaitlisted
		reg.Position = b.nextPosition()
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
	b.promoteNext(now())
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
	count := 0
	for _, reg := range b.registrations {
		if reg.Status == StatusRegistered || reg.Status == StatusCheckedIn {
			count++
		}
	}
	return count
}

func (b *CapacityBook) nextPosition() int {
	position := 1
	for _, reg := range b.registrations {
		if reg.Status == StatusWaitlisted && reg.Position >= position {
			position = reg.Position + 1
		}
	}
	return position
}

func (b *CapacityBook) promoteNext(now time.Time) {
	var candidate *Registration
	for _, reg := range b.registrations {
		if reg.Status == StatusWaitlisted && (candidate == nil || reg.Position < candidate.Position) {
			candidate = reg
		}
	}
	if candidate == nil || b.activeCount() >= b.capacity {
		return
	}
	candidate.Status = StatusRegistered
	candidate.Position = 0
	candidate.Version++
	_ = now
}
