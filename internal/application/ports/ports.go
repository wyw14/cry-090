package ports

import (
	"context"
	"time"

	"github.com/wyw14/cry-090/internal/domain/audit"
	"github.com/wyw14/cry-090/internal/domain/common"
	"github.com/wyw14/cry-090/internal/domain/event"
	"github.com/wyw14/cry-090/internal/domain/identity"
	"github.com/wyw14/cry-090/internal/domain/invitation"
	"github.com/wyw14/cry-090/internal/domain/matching"
	"github.com/wyw14/cry-090/internal/domain/registration"
)

type Clock interface{ Now() time.Time }
type IDGenerator interface{ NewID(prefix string) string }
type TxManager interface {
	Within(ctx context.Context, fn func(context.Context) error) error
}

type UserRepository interface {
	GetProfile(ctx context.Context, userID string) (identity.Profile, error)
	SaveProfile(ctx context.Context, profile identity.Profile, expectedVersion int64) error
}
type EventRepository interface {
	Get(ctx context.Context, id string) (event.Event, error)
	Save(ctx context.Context, value event.Event, expectedVersion int64) error
	List(ctx context.Context, page common.PageRequest) (common.Page[event.Event], error)
}
type NeedRepository interface {
	Get(ctx context.Context, id string) (matching.Need, error)
	Save(ctx context.Context, need matching.Need, expectedVersion int64) error
	ListForEvent(ctx context.Context, eventID string) ([]matching.Need, error)
}
type InvitationRepository interface {
	Get(ctx context.Context, id string) (invitation.Invitation, error)
	Save(ctx context.Context, value invitation.Invitation, expectedVersion int64) error
	ListAcceptedForUser(ctx context.Context, userID string) ([]invitation.Invitation, error)
}
type RegistrationRepository interface {
	Get(ctx context.Context, eventID, userID string) (registration.Registration, error)
	Save(ctx context.Context, value registration.Registration, expectedVersion int64) error
	ListForEvent(ctx context.Context, eventID string) ([]registration.Registration, error)
	Promote(ctx context.Context, eventID string) error
}
type AuditRepository interface {
	Append(ctx context.Context, entry audit.Entry) error
	List(ctx context.Context, entityID string) ([]audit.Entry, error)
}
type NotificationSender interface {
	Send(ctx context.Context, userID, subject, body string) error
}

type Dependencies struct {
	Clock         Clock
	IDs           IDGenerator
	Tx            TxManager
	Users         UserRepository
	Events        EventRepository
	Needs         NeedRepository
	Invitations   InvitationRepository
	Registrations RegistrationRepository
	Audits        AuditRepository
	Notifications NotificationSender
}
