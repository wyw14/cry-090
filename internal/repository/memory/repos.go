package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/wyw14/cry-090/internal/domain/audit"
	"github.com/wyw14/cry-090/internal/domain/common"
	"github.com/wyw14/cry-090/internal/domain/event"
	"github.com/wyw14/cry-090/internal/domain/identity"
	"github.com/wyw14/cry-090/internal/domain/invitation"
	"github.com/wyw14/cry-090/internal/domain/matching"
	"github.com/wyw14/cry-090/internal/domain/registration"
)

type Store struct {
	mu            sync.RWMutex
	Users         map[string]identity.Profile
	Events        map[string]event.Event
	Needs         map[string]matching.Need
	Invitations   map[string]invitation.Invitation
	Registrations map[string]registration.Registration
	Audits        []audit.Entry
}

func NewStore() *Store {
	return &Store{Users: map[string]identity.Profile{}, Events: map[string]event.Event{}, Needs: map[string]matching.Need{}, Invitations: map[string]invitation.Invitation{}, Registrations: map[string]registration.Registration{}}
}

func (s *Store) GetProfile(_ context.Context, id string) (identity.Profile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.Users[id]
	if !ok {
		return identity.Profile{}, common.NewError(common.CodeNotFound, "profile not found")
	}
	return v, nil
}
func (s *Store) SaveProfile(_ context.Context, v identity.Profile, expected int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if old, ok := s.Users[v.UserID]; ok && expected != 0 && old.Version != expected {
		return common.NewError(common.CodeVersionConflict, "profile version conflict")
	}
	s.Users[v.UserID] = v
	return nil
}
func (s *Store) Get(_ context.Context, id string) (event.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.Events[id]
	if !ok {
		return event.Event{}, common.NewError(common.CodeNotFound, "event not found")
	}
	return v, nil
}
func (s *Store) Save(_ context.Context, v event.Event, expected int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if old, ok := s.Events[v.ID]; ok && expected != 0 && old.Version != expected {
		return common.NewError(common.CodeVersionConflict, "event version conflict")
	}
	s.Events[v.ID] = v
	return nil
}
func (s *Store) List(_ context.Context, p common.PageRequest) (common.Page[event.Event], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]event.Event, 0, len(s.Events))
	for _, v := range s.Events {
		items = append(items, v)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].UpdatedAt.After(items[j].UpdatedAt) })
	if p.Limit == 0 || p.Limit > len(items) {
		p.Limit = len(items)
	}
	start := p.Page * p.Limit
	if start > len(items) {
		start = len(items)
	}
	end := start + p.Limit
	if end > len(items) {
		end = len(items)
	}
	return common.Page[event.Event]{Items: items[start:end], Page: p.Page, Total: len(items)}, nil
}
func (s *Store) GetNeed(_ context.Context, id string) (matching.Need, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.Needs[id]
	if !ok {
		return matching.Need{}, common.NewError(common.CodeNotFound, "partner need not found")
	}
	return v, nil
}
func (s *Store) SaveNeed(_ context.Context, v matching.Need, expected int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if old, ok := s.Needs[v.ID]; ok && expected != 0 && old.Version != expected {
		return common.NewError(common.CodeVersionConflict, "need version conflict")
	}
	s.Needs[v.ID] = v
	return nil
}
func (s *Store) ListNeeds(_ context.Context, eventID string) ([]matching.Need, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []matching.Need{}
	for _, v := range s.Needs {
		if v.EventID == eventID {
			result = append(result, v)
		}
	}
	return result, nil
}
func (s *Store) GetInvitation(_ context.Context, id string) (invitation.Invitation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.Invitations[id]
	if !ok {
		return invitation.Invitation{}, common.NewError(common.CodeNotFound, "invitation not found")
	}
	return v, nil
}
func (s *Store) SaveInvitation(_ context.Context, v invitation.Invitation, expected int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if old, ok := s.Invitations[v.ID]; ok && expected != 0 && old.Version != expected {
		return common.NewError(common.CodeVersionConflict, "invitation version conflict")
	}
	s.Invitations[v.ID] = v
	return nil
}
func (s *Store) ListAccepted(_ context.Context, userID string) ([]invitation.Invitation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []invitation.Invitation{}
	for _, v := range s.Invitations {
		if v.RecipientID == userID && v.Status == invitation.StatusAccepted {
			result = append(result, v)
		}
	}
	return result, nil
}
func (s *Store) GetRegistration(_ context.Context, eventID, userID string) (registration.Registration, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, v := range s.Registrations {
		if v.EventID == eventID && v.UserID == userID {
			return v, nil
		}
	}
	return registration.Registration{}, common.NewError(common.CodeNotFound, "registration not found")
}
func (s *Store) SaveRegistration(_ context.Context, v registration.Registration, expected int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := v.EventID + ":" + v.UserID
	if old, ok := s.Registrations[key]; ok && expected != 0 && old.Version != expected {
		return common.NewError(common.CodeVersionConflict, "registration version conflict")
	}
	s.Registrations[key] = v
	return nil
}

// Reserve atomically admits value against capacity. The check-then-act
// sequence — reject duplicates, count active registrations, decide registered
// vs. waitlisted, assign a position, and persist — runs entirely under the
// store lock so concurrent reservations can never oversell the last seat or
// hand out colliding waitlist positions.
func (s *Store) Reserve(_ context.Context, value registration.Registration, capacity int, expectedVersion int64) (registration.Registration, error) {
	if capacity < 1 {
		return registration.Registration{}, common.FieldError("capacity", "must be positive")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	key := value.EventID + ":" + value.UserID
	if existing, ok := s.Registrations[key]; ok && existing.Status != registration.StatusCancelled {
		return registration.Registration{}, common.NewError(common.CodeConflict, "user already has an active registration")
	}
	if old, ok := s.Registrations[key]; ok && expectedVersion != 0 && old.Version != expectedVersion {
		return registration.Registration{}, common.NewError(common.CodeVersionConflict, "registration version conflict")
	}

	active := 0
	waitlisted := 0
	for _, reg := range s.Registrations {
		if reg.EventID != value.EventID {
			continue
		}
		switch reg.Status {
		case registration.StatusRegistered, registration.StatusCheckedIn:
			active++
		case registration.StatusWaitlisted:
			waitlisted++
		}
	}

	if active < capacity {
		value.Status = registration.StatusRegistered
		value.Position = 0
	} else {
		value.Status = registration.StatusWaitlisted
		value.Position = waitlisted + 1
	}
	s.Registrations[key] = value
	return value, nil
}
func (s *Store) ListRegistrations(_ context.Context, eventID string) ([]registration.Registration, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []registration.Registration{}
	for _, v := range s.Registrations {
		if v.EventID == eventID {
			result = append(result, v)
		}
	}
	return result, nil
}
func (s *Store) Promote(_ context.Context, eventID string) error { return nil }
func (s *Store) AppendAudit(_ context.Context, v audit.Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Audits = append(s.Audits, v)
	return nil
}
func (s *Store) ListAudits(_ context.Context, entityID string) ([]audit.Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []audit.Entry{}
	for _, v := range s.Audits {
		if v.EntityID == entityID {
			result = append(result, v)
		}
	}
	return result, nil
}
