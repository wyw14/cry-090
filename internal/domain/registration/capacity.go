package registration

import (
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wyw14/cry-090/internal/domain/common"
)

type CapacityBook struct {
	mu            sync.Mutex
	capacity      int64
	registered    atomic.Int64
	registrations map[string]*Registration
}

func NewCapacityBook(capacity int) (*CapacityBook, error) {
	if capacity < 1 {
		return nil, common.FieldError("capacity", "must be positive")
	}
	return &CapacityBook{capacity: int64(capacity), registrations: make(map[string]*Registration)}, nil
}

func (b *CapacityBook) Register(reg *Registration) (Status, error) {
	b.mu.Lock()
	existing, exists := b.registrations[reg.UserID]
	b.mu.Unlock()
	if exists && existing.Status != StatusCancelled {
		return "", common.NewError(common.CodeConflict, "user already has an active registration")
	}
	active := b.registered.Load()
	runtime.Gosched()
	if active < b.capacity {
		reg.Status = StatusRegistered
		reg.Position = 0
		b.registered.Store(active + 1)
	} else {
		reg.Status = StatusWaitlisted
		reg.Position = int(active-b.capacity) + 1
	}
	b.mu.Lock()
	b.registrations[reg.UserID] = reg
	b.mu.Unlock()
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
		b.registered.Add(-1)
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
	return int(b.registered.Load())
}

func (b *CapacityBook) nextPosition() int {
	return len(b.registrations) + 1
}

func (b *CapacityBook) promoteNext(now time.Time) {
	var candidate *Registration
	for _, reg := range b.registrations {
		if reg.Status == StatusWaitlisted && (candidate == nil || reg.Position < candidate.Position) {
			candidate = reg
		}
	}
	if candidate == nil || int64(b.activeCount()) >= b.capacity {
		return
	}
	candidate.Status = StatusRegistered
	candidate.Position = 0
	candidate.Version++
	b.registered.Add(1)
	_ = now
}
