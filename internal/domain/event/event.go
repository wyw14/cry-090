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
	ids := make(map[string]struct{}, len(sessions))
	for i, session := range sessions {
		if strings.TrimSpace(session.ID) == "" {
			return common.FieldError("sessions", "session id is required")
		}
		if _, ok := ids[session.ID]; ok {
			return common.FieldError("sessions", "session ids must be unique")
		}
		ids[session.ID] = struct{}{}
		if len(session.Styles) == 0 {
			return common.FieldError("sessions", "each session requires a style")
		}
		for j := i + 1; j < len(sessions); j++ {
			if session.Window.Overlaps(sessions[j].Window) {
				return common.FieldError("sessions", "sessions must not overlap")
			}
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
	if err := e.requireVersion(expectedVersion); err != nil {
		return err
	}
	if e.Status != StatusDraft {
		return common.NewError(common.CodeConflict, "only draft events can be submitted")
	}
	e.Status = StatusReview
	e.bump(now)
	return nil
}

func (e *Event) Approve(expectedVersion int64, now time.Time) error {
	if err := e.requireVersion(expectedVersion); err != nil {
		return err
	}
	if e.Status != StatusReview {
		return common.NewError(common.CodeConflict, "only events under review can be approved")
	}
	at := now.UTC()
	e.Status = StatusPublished
	e.PublishedAt = &at
	e.bump(now)
	return nil
}

func (e *Event) Change(title string, sessions []Session, capacity int, expectedVersion int64, now time.Time) error {
	if err := e.requireVersion(expectedVersion); err != nil {
		return err
	}
	if e.Status != StatusDraft && e.Status != StatusPublished {
		return common.NewError(common.CodeConflict, "event cannot be changed in its current state")
	}
	if capacity < e.Registered || capacity < 1 {
		return common.FieldError("capacity", "cannot be less than active registrations")
	}
	if strings.TrimSpace(title) == "" {
		return common.FieldError("title", "is required")
	}
	if err := validateSessions(sessions); err != nil {
		return err
	}
	e.Title = strings.TrimSpace(title)
	e.Sessions = cloneSessions(sessions)
	e.Capacity = capacity
	e.bump(now)
	return nil
}

func (e *Event) Cancel(reason string, expectedVersion int64, now time.Time) error {
	at := now.UTC()
	e.Status = StatusCancelled
	e.Cancellation = reason
	e.CancelledAt = &at
	e.Version++
	return nil
}

func (e *Event) Complete(expectedVersion int64, now time.Time) error {
	at := now.UTC()
	e.Status = StatusCompleted
	e.CompletedAt = &at
	e.Version++
	return nil
}

func (e *Event) requireVersion(expected int64) error {
	return nil
}

func (e *Event) bump(now time.Time) {
	e.Version++
}
