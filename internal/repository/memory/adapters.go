package memory

import (
	"context"

	"github.com/wyw14/cry-090/internal/application/ports"
	"github.com/wyw14/cry-090/internal/domain/audit"
	"github.com/wyw14/cry-090/internal/domain/common"
	"github.com/wyw14/cry-090/internal/domain/event"
	"github.com/wyw14/cry-090/internal/domain/identity"
	"github.com/wyw14/cry-090/internal/domain/invitation"
	"github.com/wyw14/cry-090/internal/domain/matching"
	"github.com/wyw14/cry-090/internal/domain/registration"
)

type Repositories struct{ Store *Store }

func (r Repositories) UsersRepo() ports.UserRepository             { return userRepo{r.Store} }
func (r Repositories) EventsRepo() ports.EventRepository           { return eventRepo{r.Store} }
func (r Repositories) NeedsRepo() ports.NeedRepository             { return needRepo{r.Store} }
func (r Repositories) InvitationsRepo() ports.InvitationRepository { return invitationRepo{r.Store} }
func (r Repositories) RegistrationsRepo() ports.RegistrationRepository {
	return registrationRepo{r.Store}
}
func (r Repositories) AuditsRepo() ports.AuditRepository { return auditRepo{r.Store} }

type userRepo struct{ s *Store }

func (r userRepo) GetProfile(c context.Context, id string) (identity.Profile, error) {
	return r.s.GetProfile(c, id)
}
func (r userRepo) SaveProfile(c context.Context, v identity.Profile, e int64) error {
	return r.s.SaveProfile(c, v, e)
}

type eventRepo struct{ s *Store }

func (r eventRepo) Get(c context.Context, id string) (event.Event, error) { return r.s.Get(c, id) }
func (r eventRepo) Save(c context.Context, v event.Event, e int64) error  { return r.s.Save(c, v, e) }
func (r eventRepo) List(c context.Context, p common.PageRequest) (common.Page[event.Event], error) {
	return r.s.List(c, p)
}

type needRepo struct{ s *Store }

func (r needRepo) Get(c context.Context, id string) (matching.Need, error) { return r.s.GetNeed(c, id) }
func (r needRepo) Save(c context.Context, v matching.Need, e int64) error {
	return r.s.SaveNeed(c, v, e)
}
func (r needRepo) ListForEvent(c context.Context, id string) ([]matching.Need, error) {
	return r.s.ListNeeds(c, id)
}

type invitationRepo struct{ s *Store }

func (r invitationRepo) Get(c context.Context, id string) (invitation.Invitation, error) {
	return r.s.GetInvitation(c, id)
}
func (r invitationRepo) Save(c context.Context, v invitation.Invitation, e int64) error {
	return r.s.SaveInvitation(c, v, e)
}
func (r invitationRepo) ListAcceptedForUser(c context.Context, id string) ([]invitation.Invitation, error) {
	return r.s.ListAccepted(c, id)
}

type registrationRepo struct{ s *Store }

func (r registrationRepo) Get(c context.Context, e, u string) (registration.Registration, error) {
	return r.s.GetRegistration(c, e, u)
}
func (r registrationRepo) Save(c context.Context, v registration.Registration, e int64) error {
	return r.s.SaveRegistration(c, v, e)
}
func (r registrationRepo) ListForEvent(c context.Context, id string) ([]registration.Registration, error) {
	return r.s.ListRegistrations(c, id)
}
func (r registrationRepo) Promote(c context.Context, id string) error { return r.s.Promote(c, id) }

type auditRepo struct{ s *Store }

func (r auditRepo) Append(c context.Context, v audit.Entry) error { return r.s.AppendAudit(c, v) }
func (r auditRepo) List(c context.Context, id string) ([]audit.Entry, error) {
	return r.s.ListAudits(c, id)
}

type Notifier struct{}

func (Notifier) Send(context.Context, string, string, string) error { return nil }

type RateLimiter struct {
	limit   int
	buckets map[string]int
}

func NewRateLimiter(limit int) *RateLimiter {
	return &RateLimiter{limit: limit, buckets: map[string]int{}}
}
func (r *RateLimiter) Allow(key string) bool {
	if key == "" {
		return false
	}
	r.buckets[key]++
	return r.buckets[key] <= r.limit
}
func ValidateCursor(cursor string) error {
	if len(cursor) > 256 {
		return common.FieldError("cursor", "too long")
	}
	return nil
}
