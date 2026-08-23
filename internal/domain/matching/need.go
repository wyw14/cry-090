package matching

import (
	"strings"
	"time"

	"github.com/wyw14/cry-090/internal/domain/common"
	"github.com/wyw14/cry-090/internal/domain/event"
	"github.com/wyw14/cry-090/internal/domain/identity"
)

type Level string

const (
	LevelBeginner     Level = "beginner"
	LevelIntermediate Level = "intermediate"
	LevelAdvanced     Level = "advanced"
)

type Need struct {
	ID                string
	OwnerID           string
	EventID           string
	Window            common.TimeRange
	Styles            []event.SalsaStyle
	Level             Level
	WantedRole        identity.DanceRole
	Tags              []string
	MaxDistanceBucket int
	Visibility        identity.Visibility
	Active            bool
	ExpiresAt         time.Time
	Version           int64
}

func NewNeed(id, ownerID, eventID string, window common.TimeRange, styles []event.SalsaStyle, level Level,
	wantedRole identity.DanceRole, tags []string, maxDistance int, visibility identity.Visibility, expiresAt time.Time) (*Need, error) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(ownerID) == "" || strings.TrimSpace(eventID) == "" {
		return nil, common.FieldError("need", "id, owner and event are required")
	}
	if len(styles) == 0 {
		return nil, common.FieldError("styles", "at least one style is required")
	}
	if maxDistance < 0 || maxDistance > 5 {
		return nil, common.FieldError("max_distance_bucket", "must be between 0 and 5")
	}
	if !expiresAt.UTC().After(window.Start) {
		return nil, common.FieldError("expires_at", "must be after the need opens")
	}
	return &Need{ID: id, OwnerID: ownerID, EventID: eventID, Window: window,
		Styles: append([]event.SalsaStyle(nil), styles...), Level: level, WantedRole: wantedRole,
		Tags: normalizeTags(tags), MaxDistanceBucket: maxDistance, Visibility: visibility,
		Active: true, ExpiresAt: expiresAt.UTC(), Version: 1}, nil
}

func normalizeTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		result = append(result, tag)
	}
	return result
}

func (n *Need) Close(expectedVersion int64) error {
	if n.Version != expectedVersion {
		return common.NewError(common.CodeVersionConflict, "partner need was changed")
	}
	if !n.Active {
		return common.NewError(common.CodeAlreadyProcessed, "partner need is already closed")
	}
	n.Active = false
	n.Version++
	return nil
}
