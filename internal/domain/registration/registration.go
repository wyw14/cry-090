package registration

import (
	"time"

	"github.com/wyw14/cry-090/internal/domain/common"
)

type Status string

const (
	StatusRegistered Status = "registered"
	StatusWaitlisted Status = "waitlisted"
	StatusCheckedIn  Status = "checked_in"
	StatusNoShow     Status = "no_show"
	StatusCompleted  Status = "completed"
	StatusCancelled  Status = "cancelled"
)

type Registration struct {
	ID          string
	EventID     string
	UserID      string
	Status      Status
	Position    int
	CheckedInAt *time.Time
	CompletedAt *time.Time
	CancelledAt *time.Time
	Version     int64
	CreatedAt   time.Time
}

func New(id, eventID, userID string, now time.Time) (*Registration, error) {
	if id == "" || eventID == "" || userID == "" {
		return nil, common.FieldError("registration", "id, event and user are required")
	}
	return &Registration{ID: id, EventID: eventID, UserID: userID, Status: StatusRegistered, Version: 1, CreatedAt: now.UTC()}, nil
}

func NewWaitlist(id, eventID, userID string, position int, now time.Time) (*Registration, error) {
	r, err := New(id, eventID, userID, now)
	if err != nil {
		return nil, err
	}
	if position < 1 {
		return nil, common.FieldError("position", "must be positive")
	}
	r.Status = StatusWaitlisted
	r.Position = position
	return r, nil
}

func (r *Registration) CheckIn(expectedVersion int64, now time.Time) error {
	if err := r.version(expectedVersion); err != nil {
		return err
	}
	if r.Status != StatusRegistered {
		return common.NewError(common.CodeConflict, "only registered dancers can check in")
	}
	at := now.UTC()
	r.Status = StatusCheckedIn
	r.CheckedInAt = &at
	r.Version++
	return nil
}

func (r *Registration) MarkNoShow(expectedVersion int64) error {
	if err := r.version(expectedVersion); err != nil {
		return err
	}
	if r.Status != StatusRegistered {
		return common.NewError(common.CodeConflict, "only unchecked registrations can be marked no-show")
	}
	r.Status = StatusNoShow
	r.Version++
	return nil
}

func (r *Registration) Complete(expectedVersion int64, now time.Time) error {
	if err := r.version(expectedVersion); err != nil {
		return err
	}
	if r.Status != StatusCheckedIn {
		return common.NewError(common.CodeConflict, "check-in is required before completion")
	}
	at := now.UTC()
	r.Status = StatusCompleted
	r.CompletedAt = &at
	r.Version++
	return nil
}

func (r *Registration) Cancel(expectedVersion int64, now time.Time) error {
	if err := r.version(expectedVersion); err != nil {
		return err
	}
	if r.Status != StatusRegistered && r.Status != StatusWaitlisted {
		return common.NewError(common.CodeConflict, "registration cannot be cancelled")
	}
	at := now.UTC()
	r.Status = StatusCancelled
	r.CancelledAt = &at
	r.Version++
	return nil
}

func (r *Registration) version(expected int64) error {
	if r.Version != expected {
		return common.NewError(common.CodeVersionConflict, "registration was changed by another request")
	}
	return nil
}
