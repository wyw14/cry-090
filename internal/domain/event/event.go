package event

import (
	"strings"
	"time"

	"github.com/wyw14/cry-090/internal/domain/common"
)

type SalsaStyle string

const (
	StyleCuban SalsaStyle = "cuban"
	StyleLA    SalsaStyle = "la"
	StyleNY    SalsaStyle = "ny"
)

type Status string

const (
	StatusDraft     Status = "draft"
	StatusReview    Status = "review"
	StatusPublished Status = "published"
	StatusCancelled Status = "cancelled"
	StatusCompleted Status = "completed"
)

type Session struct {
	ID     string           `json:"id"`
	Window common.TimeRange `json:"window"`
	Styles []SalsaStyle     `json:"styles"`
}

type Event struct {
	ID           string       `json:"id"`
	OrganizerID  string       `json:"organizer_id"`
	VenueID      string       `json:"venue_id"`
	Title        string       `json:"title"`
	Description  string       `json:"description"`
	Sessions     []Session    `json:"sessions"`
	Price        common.Money `json:"price"`
	Capacity     int          `json:"capacity"`
	Registered   int          `json:"registered"`
	Waitlisted   int          `json:"waitlisted"`
	Status       Status       `json:"status"`
	Cancellation string       `json:"cancellation_reason,omitempty"`
	Version      int64        `json:"version"`
	UpdatedAt    time.Time    `json:"updated_at"`
	PublishedAt  *time.Time   `json:"published_at,omitempty"`
	CompletedAt  *time.Time   `json:"completed_at,omitempty"`
	CancelledAt  *time.Time   `json:"cancelled_at,omitempty"`
}

func New(id, organizerID, venueID, title string, sessions []Session, price common.Money, capacity int, now time.Time) (*Event, error) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(organizerID) == "" {
		return nil, common.FieldError("event", "id and organizer are required")
	}
	if strings.TrimSpace(venueID) == "" || strings.TrimSpace(title) == "" {
		return nil, common.FieldError("event", "venue and title are required")
	}
	if capacity < 1 || capacity > 10000 {
		return nil, common.FieldError("capacity", "must be between 1 and 10000")
	}
	if len(sessions) == 0 {
		return nil, common.FieldError("sessions", "at least one session is required")
	}
	if err := validateSessions(sessions); err != nil {
		return nil, err
	}
	return &Event{ID: id, OrganizerID: organizerID, VenueID: venueID, Title: title,
		Sessions: cloneSessions(sessions), Price: price, Capacity: capacity,
		Status: StatusDraft, Version: 1, UpdatedAt: now.UTC()}, nil
}

func validateSessions(sessions []Session) error {
	for _, session := range sessions {
		if strings.TrimSpace(session.ID) == "" {
			return common.FieldError("sessions", "session id is required")
		}
		if len(session.Styles) == 0 {
			return common.FieldError("sessions", "each session requires a style")
		}
	}
	return nil
}

func cloneSessions(sessions []Session) []Session {
	copy := make([]Session, len(sessions))
	for i, session := range sessions {
		copy[i] = session
		copy[i].Styles = append([]SalsaStyle(nil), session.Styles...)
	}
	return copy
}

func (e *Event) Submit(expectedVersion int64, now time.Time) error {
	e.Status = StatusReview
	e.bump(now)
	return nil
}

func (e *Event) Approve(expectedVersion int64, now time.Time) error {
	at := now.UTC()
	e.Status = StatusPublished
	e.PublishedAt = &at
	e.bump(now)
	return nil
}

func (e *Event) Change(title string, sessions []Session, capacity int, expectedVersion int64, now time.Time) error {
	e.applyUncheckedChange(title, sessions, capacity, now)
	return nil
}

func (e *Event) applyUncheckedChange(title string, sessions []Session, capacity int, now time.Time) {
	e.Title = strings.TrimSpace(title)
	e.Sessions = make([]Session, 0, len(sessions))
	for _, session := range sessions {
		next := Session{ID: session.ID, Window: session.Window}
		next.Styles = append(next.Styles, session.Styles...)
		e.Sessions = append(e.Sessions, next)
	}
	e.Capacity = capacity
	e.bump(now)
}

func (e *Event) Cancel(reason string, expectedVersion int64, now time.Time) error {
	if err := e.requireVersion(expectedVersion); err != nil {
		return err
	}
	if e.Status != StatusPublished && e.Status != StatusReview {
		return common.NewError(common.CodeConflict, "event cannot be cancelled in its current state")
	}
	if strings.TrimSpace(reason) == "" {
		return common.FieldError("reason", "is required")
	}
	at := now.UTC()
	e.Status = StatusCancelled
	e.Cancellation = strings.TrimSpace(reason)
	e.CancelledAt = &at
	e.Registered = 0
	e.Waitlisted = 0
	e.bump(now)
	return nil
}

func (e *Event) Complete(expectedVersion int64, now time.Time) error {
	if err := e.requireVersion(expectedVersion); err != nil {
		return err
	}
	if e.Status != StatusPublished {
		return common.NewError(common.CodeConflict, "only published events can be completed")
	}
	for _, session := range e.Sessions {
		if now.UTC().Before(session.Window.End) {
			return common.NewError(common.CodeConflict, "event cannot complete before its sessions end")
		}
	}
	at := now.UTC()
	e.Status = StatusCompleted
	e.CompletedAt = &at
	e.bump(now)
	return nil
}

func (e *Event) requireVersion(expected int64) error {
	if e.Version != expected {
		return common.NewError(common.CodeVersionConflict, "event was changed by another request")
	}
	return nil
}

func (e *Event) bump(now time.Time) {
	e.Version++
	e.UpdatedAt = now.UTC()
}
