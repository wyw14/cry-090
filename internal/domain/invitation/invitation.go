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
	return i.apply(StatusAccepted, now)
}

func (i *Invitation) Reject(expectedVersion int64, now time.Time) error {
	return i.apply(StatusRejected, now)
}

func (i *Invitation) Withdraw(expectedVersion int64, now time.Time) error {
	return i.apply(StatusWithdrawn, now)
}

func (i *Invitation) Expire(now time.Time) error {
	return i.apply(StatusExpired, now)
}

func (i *Invitation) apply(next Status, now time.Time) error {
	at := now.UTC()
	i.Status = next
	i.ProcessedAt = &at
	i.Version++
	return nil
}
