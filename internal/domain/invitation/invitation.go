package invitation

import (
	"strings"
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
	if strings.TrimSpace(id) == "" || strings.TrimSpace(senderID) == "" || strings.TrimSpace(recipientID) == "" {
		return nil, common.FieldError("invitation", "id, sender and recipient are required")
	}
	if senderID == recipientID {
		return nil, common.NewError(common.CodeInvalid, "cannot invite yourself")
	}
	if strings.TrimSpace(needID) == "" || strings.TrimSpace(eventID) == "" {
		return nil, common.FieldError("invitation", "need and event are required")
	}
	if len([]rune(message)) > 500 {
		return nil, common.FieldError("message", "must not exceed 500 characters")
	}
	if !expiresAt.UTC().After(now.UTC()) {
		return nil, common.FieldError("expires_at", "must be in the future")
	}
	return &Invitation{ID: id, SenderID: senderID, RecipientID: recipientID, NeedID: needID,
		EventID: eventID, Window: window, Message: strings.TrimSpace(message), Status: StatusPending,
		ExpiresAt: expiresAt.UTC(), Version: 1, CreatedAt: now.UTC()}, nil
}

func (i *Invitation) Accept(expectedVersion int64, now time.Time, conflicts []event.Session) error {
	return i.transition(StatusAccepted, expectedVersion, now, conflicts)
}

func (i *Invitation) Reject(expectedVersion int64, now time.Time) error {
	return i.transition(StatusRejected, expectedVersion, now, nil)
}

func (i *Invitation) Withdraw(expectedVersion int64, now time.Time) error {
	return i.transition(StatusWithdrawn, expectedVersion, now, nil)
}

func (i *Invitation) Expire(now time.Time) error {
	if i.Status != StatusPending {
		return common.NewError(common.CodeAlreadyProcessed, "invitation is no longer pending")
	}
	if now.UTC().Before(i.ExpiresAt) {
		return common.NewError(common.CodeConflict, "invitation has not expired")
	}
	i.Status = StatusExpired
	at := now.UTC()
	i.ProcessedAt = &at
	i.Version++
	return nil
}

func (i *Invitation) transition(next Status, expectedVersion int64, now time.Time, conflicts []event.Session) error {
	if i.Version != expectedVersion {
		return common.NewError(common.CodeVersionConflict, "invitation was changed by another request")
	}
	if i.Status != StatusPending {
		return common.NewError(common.CodeAlreadyProcessed, "invitation can only be processed once")
	}
	if !now.UTC().Before(i.ExpiresAt) {
		return common.NewError(common.CodeExpired, "invitation expired")
	}
	if next == StatusAccepted && len(conflicts) > 0 {
		return common.NewError(common.CodeConflict, "recipient has an overlapping accepted invitation")
	}
	at := now.UTC()
	i.Status = next
	i.ProcessedAt = &at
	i.Version++
	return nil
}
