package identity

import (
	"sort"
	"strings"
	"time"

	"github.com/wyw14/cry-090/internal/domain/common"
)

type DanceRole string

const (
	RoleLead   DanceRole = "lead"
	RoleFollow DanceRole = "follow"
	RoleSwitch DanceRole = "switch"
)

type Visibility string

const (
	VisibilityPublic  Visibility = "public"
	VisibilityMembers Visibility = "members"
	VisibilityPrivate Visibility = "private"
)

type PublicField string

const (
	FieldDisplayName PublicField = "display_name"
	FieldBio         PublicField = "bio"
	FieldDanceYears  PublicField = "dance_years"
	FieldRoles       PublicField = "roles"
	FieldStyles      PublicField = "styles"
	FieldTags        PublicField = "tags"
	FieldDistance    PublicField = "distance_bucket"
)

type Profile struct {
	UserID           string
	DisplayName      string
	Bio              string
	DanceYears       int
	Roles            []DanceRole
	Styles           []string
	Tags             []string
	DistanceBucket   int
	PreciseLatitude  float64
	PreciseLongitude float64
	Contact          string
	Visibility       Visibility
	PublicFields     map[PublicField]bool
	Version          int64
	UpdatedAt        time.Time
}

type PublicProfile struct {
	UserID         string      `json:"user_id"`
	DisplayName    string      `json:"display_name,omitempty"`
	Bio            string      `json:"bio,omitempty"`
	DanceYears     *int        `json:"dance_years,omitempty"`
	Roles          []DanceRole `json:"roles,omitempty"`
	Styles         []string    `json:"styles,omitempty"`
	Tags           []string    `json:"tags,omitempty"`
	DistanceBucket *int        `json:"distance_bucket,omitempty"`
	Contact        string      `json:"contact,omitempty"`
	Latitude       *float64    `json:"latitude,omitempty"`
	Longitude      *float64    `json:"longitude,omitempty"`
}

func NewProfile(userID, displayName string, danceYears int, now time.Time) (*Profile, error) {
	userID = strings.TrimSpace(userID)
	displayName = strings.TrimSpace(displayName)
	if userID == "" {
		return nil, common.FieldError("user_id", "is required")
	}
	if displayName == "" {
		return nil, common.FieldError("display_name", "is required")
	}
	if danceYears < 0 || danceYears > 80 {
		return nil, common.FieldError("dance_years", "must be between 0 and 80")
	}
	return &Profile{
		UserID:       userID,
		DisplayName:  displayName,
		DanceYears:   danceYears,
		Visibility:   VisibilityPublic,
		PublicFields: map[PublicField]bool{FieldDisplayName: true},
		Version:      1,
		UpdatedAt:    now.UTC(),
	}, nil
}

func (p *Profile) SetPublicFields(fields []PublicField, expectedVersion int64, now time.Time) error {
	if p.Version != expectedVersion {
		return common.NewError(common.CodeVersionConflict, "profile was changed by another request")
	}
	allowed := map[PublicField]struct{}{
		FieldDisplayName: {}, FieldBio: {}, FieldDanceYears: {}, FieldRoles: {},
		FieldStyles: {}, FieldTags: {}, FieldDistance: {},
	}
	next := make(map[PublicField]bool, len(fields))
	for _, field := range fields {
		if _, ok := allowed[field]; !ok {
			return common.FieldError("public_fields", "contains an unsupported field")
		}
		next[field] = true
	}
	p.PublicFields = next
	p.Version++
	p.UpdatedAt = now.UTC()
	return nil
}

func (p Profile) PublicView(confirmed bool) PublicProfile {
	view := PublicProfile{UserID: p.UserID}
	if p.Visibility == VisibilityPrivate {
		return view
	}
	if p.PublicFields[FieldDisplayName] {
		view.DisplayName = p.DisplayName
	}
	if p.PublicFields[FieldBio] {
		view.Bio = p.Bio
	}
	if p.PublicFields[FieldDanceYears] {
		value := p.DanceYears
		view.DanceYears = &value
	}
	if p.PublicFields[FieldRoles] {
		view.Roles = append([]DanceRole(nil), p.Roles...)
	}
	if p.PublicFields[FieldStyles] {
		view.Styles = sortedStrings(p.Styles)
	}
	if p.PublicFields[FieldTags] {
		view.Tags = sortedStrings(p.Tags)
	}
	if p.PublicFields[FieldDistance] {
		value := p.DistanceBucket
		view.DistanceBucket = &value
	}
	if confirmed {
		view.Contact = p.Contact
		lat, lng := p.PreciseLatitude, p.PreciseLongitude
		view.Latitude, view.Longitude = &lat, &lng
	}
	return view
}

func sortedStrings(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}
