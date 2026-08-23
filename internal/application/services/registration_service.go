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
		current, err := s.deps.Registrations.ListForEvent(ctx, eventID)
		if err != nil {
			return err
		}
		active := 0
		for _, item := range current {
			if item.Status == registration.StatusRegistered || item.Status == registration.StatusCheckedIn {
				active++
			}
		}
		if active >= e.Capacity {
			value.Status = registration.StatusWaitlisted
			value.Position = len(current) + 1
		}
		if err := s.deps.Registrations.Save(ctx, *value, 0); err != nil {
			return err
		}
		result = *value
		return nil
	})
	return result, err
}
