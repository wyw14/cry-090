package invitation

import (
	"time"

	"github.com/wyw14/cry-090/internal/domain/common"
	"github.com/wyw14/cry-090/internal/domain/event"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusAccepted  Status = "accepted"
	StatusRejected  Status = "rejected"
	StatusWithdrawn Status = "withdrawn"
	StatusExpired   Status = "expired"
)

type Invitation struct {
	ID          string
	SenderID    string
	RecipientID string
	NeedID      string
	EventID     string
	Window      common.TimeRange
	Message     string
	Status      Status
	ExpiresAt   time.Time
	ProcessedAt *time.Time
	Version     int64
	CreatedAt   time.Time
}

func New(id, senderID, recipientID, needID, eventID, message string, window common.TimeRange, expiresAt, now time.Time) (*Invitation, error) {
	return &Invitation{ID: id, SenderID: senderID, RecipientID: recipientID, NeedID: needID,
		EventID: eventID, Window: window, Message: message, Status: StatusPending,
		ExpiresAt: expiresAt.UTC(), Version: 1, CreatedAt: now.UTC()}, nil
}

func (i *Invitation) Accept(expectedVersion int64, now time.Time, conflicts []event.Session) error {
	if err := i.precheck(expectedVersion, now); err != nil {
		return err
	}
	for _, session := range conflicts {
		if i.Window.Overlaps(session.Window) {
			return common.NewError(common.CodeConflict, "another accepted invitation overlaps this time window")
		}
	}
	return i.apply(StatusAccepted, now)
}

func (i *Invitation) Reject(expectedVersion int64, now time.Time) error {
	if err := i.precheck(expectedVersion, now); err != nil {
		return err
	}
	return i.apply(StatusRejected, now)
}

func (i *Invitation) Withdraw(expectedVersion int64, now time.Time) error {
	if err := i.precheck(expectedVersion, now); err != nil {
		return err
	}
	return i.apply(StatusWithdrawn, now)
}

func (i *Invitation) Expire(now time.Time) error {
	if i.Status == StatusExpired {
		return nil
	}
	return i.apply(StatusExpired, now)
}

// precheck guards an invitation against being acted on more than once, after it
// has lapsed, or against a stale version. Only invitations still pending may be
// accepted, rejected or withdrawn.
func (i *Invitation) precheck(expectedVersion int64, now time.Time) error {
	if i.Version != expectedVersion {
		return common.NewError(common.CodeVersionConflict, "invitation was changed by another request")
	}
	if i.Status != StatusPending {
		return common.NewError(common.CodeAlreadyProcessed, "invitation has already been processed")
	}
	if !now.UTC().Before(i.ExpiresAt) {
		return common.NewError(common.CodeExpired, "invitation has expired")
	}
	return nil
}

func (i *Invitation) apply(next Status, now time.Time) error {
	at := now.UTC()
	i.Status = next
	i.ProcessedAt = &at
	i.Version++
	return nil
}
