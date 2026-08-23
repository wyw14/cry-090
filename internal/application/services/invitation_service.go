package services

import (
	"context"

	"github.com/wyw14/cry-090/internal/application/ports"
)

type InvitationService struct{ deps ports.Dependencies }

func NewInvitationService(deps ports.Dependencies) *InvitationService {
	return &InvitationService{deps: deps}
}

func (s *InvitationService) Accept(ctx context.Context, id, actor string, version int64) error {
	return s.deps.Tx.Within(ctx, func(ctx context.Context) error {
		value, err := s.deps.Invitations.Get(ctx, id)
		if err != nil {
			return err
		}
		if err := value.Accept(version, s.deps.Clock.Now(), nil); err != nil {
			return err
		}
		return s.deps.Invitations.Save(ctx, value, version)
	})
}
