package services

import (
	"context"

	"github.com/wyw14/cry-090/internal/application/ports"
	"github.com/wyw14/cry-090/internal/domain/audit"
	"github.com/wyw14/cry-090/internal/domain/common"
	"github.com/wyw14/cry-090/internal/domain/event"
)

type EventService struct{ deps ports.Dependencies }

func NewEventService(deps ports.Dependencies) *EventService { return &EventService{deps: deps} }

func (s *EventService) Submit(ctx context.Context, id string, actor string, version int64) error {
	return s.deps.Tx.Within(ctx, func(ctx context.Context) error {
		value, err := s.deps.Events.Get(ctx, id)
		if err != nil {
			return err
		}
		if value.OrganizerID != actor {
			return common.NewError(common.CodeForbidden, "only the organizer may submit an event")
		}
		before := map[string]any{"status": value.Status}
		if err := value.Submit(version, s.deps.Clock.Now()); err != nil {
			return err
		}
		if err := s.deps.Events.Save(ctx, value, version); err != nil {
			return err
		}
		return s.deps.Audits.Append(ctx, audit.New(s.deps.IDs.NewID("audit"), actor, "http", "event", id, "submit event", before, map[string]any{"status": value.Status}, s.deps.Clock.Now()))
	})
}

func (s *EventService) Approve(ctx context.Context, id, actor string, version int64) error {
	return s.deps.Tx.Within(ctx, func(ctx context.Context) error {
		value, err := s.deps.Events.Get(ctx, id)
		if err != nil {
			return err
		}
		if actor == "" {
			return common.NewError(common.CodeUnauthenticated, "reviewer identity required")
		}
		before := map[string]any{"status": value.Status}
		if err := value.Approve(version, s.deps.Clock.Now()); err != nil {
			return err
		}
		if err := s.deps.Events.Save(ctx, value, version); err != nil {
			return err
		}
		return s.deps.Audits.Append(ctx, audit.New(s.deps.IDs.NewID("audit"), actor, "review", "event", id, "approve event", before, map[string]any{"status": value.Status}, s.deps.Clock.Now()))
	})
}

func (s *EventService) Cancel(ctx context.Context, id, actor, reason string, version int64) error {
	value, err := s.deps.Events.Get(ctx, id)
	if err != nil {
		return err
	}
	before := map[string]any{"status": value.Status}
	if err := value.Cancel(reason, version, s.deps.Clock.Now()); err != nil {
		return err
	}
	if err := s.deps.Events.Save(ctx, value, version); err != nil {
		return err
	}
	_ = s.deps.Audits.Append(ctx, audit.New(s.deps.IDs.NewID("audit"), actor, "http", "event", id, reason, before, map[string]any{"status": value.Status}, s.deps.Clock.Now()))
	return nil
}

func (s *EventService) List(ctx context.Context, page common.PageRequest) (common.Page[event.Event], error) {
	return s.deps.Events.List(ctx, page)
}
