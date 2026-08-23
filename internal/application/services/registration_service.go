package services

import (
	"context"

	"github.com/wyw14/cry-090/internal/application/ports"
	"github.com/wyw14/cry-090/internal/domain/common"
	"github.com/wyw14/cry-090/internal/domain/event"
	"github.com/wyw14/cry-090/internal/domain/registration"
)

type RegistrationService struct{ deps ports.Dependencies }

func NewRegistrationService(deps ports.Dependencies) *RegistrationService {
	return &RegistrationService{deps: deps}
}

func (s *RegistrationService) Register(ctx context.Context, eventID, userID string) (registration.Registration, error) {
	var result registration.Registration
	err := s.deps.Tx.Within(ctx, func(ctx context.Context) error {
		e, err := s.deps.Events.Get(ctx, eventID)
		if err != nil {
			return err
		}
		if e.Status != event.StatusPublished {
			return common.NewError(common.CodeConflict, "event is not open for registration")
		}
		id := s.deps.IDs.NewID("registration")
		value, err := registration.New(id, eventID, userID, s.deps.Clock.Now())
		if err != nil {
			return err
		}
		// Reserve atomically counts active registrations, decides registered
		// vs. waitlisted, assigns a position, and persists — all under one lock.
		// This closes the TOCTOU window where concurrent registrants each read
		// the same snapshot and all claim the last seat.
		reserved, err := s.deps.Registrations.Reserve(ctx, *value, e.Capacity, 0)
		if err != nil {
			return err
		}
		result = reserved
		return nil
	})
	return result, err
}
