package services

import (
	"context"

	"github.com/wyw14/cry-090/internal/application/ports"
	"github.com/wyw14/cry-090/internal/domain/common"
	"github.com/wyw14/cry-090/internal/domain/event"
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
		if value.RecipientID != actor {
			return common.NewError(common.CodeForbidden, "only the recipient may accept an invitation")
		}
		accepted, err := s.deps.Invitations.ListAcceptedForUser(ctx, actor)
		if err != nil {
			return err
		}
		conflicts := make([]event.Session, 0)
		for _, other := range accepted {
			if other.Window.Overlaps(value.Window) {
				conflicts = append(conflicts, event.Session{ID: other.ID, Window: other.Window})
			}
		}
		if err := value.Accept(version, s.deps.Clock.Now(), conflicts); err != nil {
			return err
		}
		return s.deps.Invitations.Save(ctx, value, version)
	})
}
