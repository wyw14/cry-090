package matching

import (
	"testing"
	"time"

	"github.com/wyw14/cry-090/internal/domain/common"
	"github.com/wyw14/cry-090/internal/domain/event"
	"github.com/wyw14/cry-090/internal/domain/identity"
)

// TestMatchHidesSensitiveFieldsUntilConfirmed reproduces the reported privacy
// leak: a member who set dance years, tags and distance bucket to private, and
// whose counterpart has not yet confirmed the match, must not see those fields
// nor contact details / precise coordinates in the match results. Only once the
// counterparty is marked confirmed in the map may the sensitive fields surface.
func TestMatchHidesSensitiveFieldsUntilConfirmed(t *testing.T) {
	now := time.Now().UTC()
	window, _ := common.NewTimeRange(now.Add(time.Hour), now.Add(3*time.Hour))
	expires := now.Add(2 * time.Hour)

	// Owner declares only display name public; dance years, tags and distance
	// are deliberately private.
	candidateProfile, _ := identity.NewProfile("owner-b", "Dancer B", 4, now)
	candidateProfile.Contact = "+86-13800000000"
	candidateProfile.PreciseLatitude = 31.2304
	candidateProfile.PreciseLongitude = 121.4737
	candidateProfile.DistanceBucket = 2
	candidateProfile.Tags = []string{"bachata", "sensual"}
	candidateProfile.Roles = []identity.DanceRole{identity.RoleFollow}
	candidateProfile.Styles = []string{"cuban"}
	candidateProfile.PublicFields = map[identity.PublicField]bool{
		identity.FieldDisplayName: true,
		identity.FieldBio:         true,
		identity.FieldRoles:       true,
		identity.FieldStyles:      true,
	}

	subject, err := NewNeed("need-a", "owner-a", "e1", window,
		[]event.SalsaStyle{event.StyleCuban}, LevelBeginner, identity.RoleFollow,
		[]string{"sensual"}, 3, identity.VisibilityPublic, expires)
	if err != nil {
		t.Fatalf("subject need: %v", err)
	}

	candidateNeed, err := NewNeed("need-b", "owner-b", "e1", window,
		[]event.SalsaStyle{event.StyleCuban}, LevelBeginner, identity.RoleFollow,
		[]string{"sensual"}, 3, identity.VisibilityPublic, expires)
	if err != nil {
		t.Fatalf("candidate need: %v", err)
	}

	candidates := []Candidate{{Need: *candidateNeed, Profile: *candidateProfile}}

	t.Run("unconfirmed match leaks nothing sensitive", func(t *testing.T) {
		results := Matcher{}.Match(*subject, candidates, now, map[string]bool{})
		if len(results) != 1 {
			t.Fatalf("expected one match, got %d", len(results))
		}
		r := results[0]
		if r.User.Contact != "" {
			t.Errorf("contact leaked before confirmation: %q", r.User.Contact)
		}
		if r.User.Latitude != nil || r.User.Longitude != nil {
			t.Errorf("precise coordinates leaked before confirmation: lat=%v lng=%v", r.User.Latitude, r.User.Longitude)
		}
		if r.User.DanceYears != nil {
			t.Errorf("private dance years leaked: %v", *r.User.DanceYears)
		}
		if r.User.Tags != nil {
			t.Errorf("private tags leaked: %v", r.User.Tags)
		}
		if r.User.DistanceBucket != nil {
			t.Errorf("private distance bucket leaked: %v", *r.User.DistanceBucket)
		}
		for _, reason := range r.Reasons {
			if reason.Code == "distance" {
				t.Errorf("distance reason surfaced with private bucket: %s", reason.Message)
			}
			if reason.Code == "tags" {
				t.Errorf("tags reason surfaced with private tags: %s", reason.Message)
			}
		}
	})

	t.Run("confirmed match reveals contact and precise coordinates", func(t *testing.T) {
		results := Matcher{}.Match(*subject, candidates, now, map[string]bool{"owner-b": true})
		if len(results) != 1 {
			t.Fatalf("expected one match, got %d", len(results))
		}
		r := results[0]
		if r.User.Contact != "+86-13800000000" {
			t.Errorf("expected contact after confirmation, got %q", r.User.Contact)
		}
		if r.User.Latitude == nil || *r.User.Latitude != 31.2304 {
			t.Errorf("expected latitude after confirmation, got %v", r.User.Latitude)
		}
		if r.User.Longitude == nil || *r.User.Longitude != 121.4737 {
			t.Errorf("expected longitude after confirmation, got %v", r.User.Longitude)
		}
	})
}

// TestMatchRespectsPublicFieldsForOptIn verifies that once a field is opted into
// the public map, it appears in the match result (and reasons) before, but not
// after, confirmation only applies to contact + precise coordinates.
func TestMatchRespectsPublicFieldsForOptIn(t *testing.T) {
	now := time.Now().UTC()
	window, _ := common.NewTimeRange(now.Add(time.Hour), now.Add(3*time.Hour))
	expires := now.Add(2 * time.Hour)

	candidateProfile, _ := identity.NewProfile("owner-d", "Dancer D", 7, now)
	candidateProfile.Tags = []string{"sensual"}
	candidateProfile.DistanceBucket = 1
	candidateProfile.Roles = []identity.DanceRole{identity.RoleFollow}
	candidateProfile.Styles = []string{"cuban"}
	candidateProfile.Contact = "secret"
	candidateProfile.PreciseLatitude = 10
	candidateProfile.PreciseLongitude = 20
	candidateProfile.PublicFields = map[identity.PublicField]bool{
		identity.FieldDisplayName: true,
		identity.FieldTags:        true,
		identity.FieldDistance:    true,
		identity.FieldRoles:       true,
		identity.FieldStyles:      true,
	}

	subject, _ := NewNeed("need-c", "owner-c", "e1", window,
		[]event.SalsaStyle{event.StyleCuban}, LevelBeginner, identity.RoleFollow,
		[]string{"sensual"}, 3, identity.VisibilityPublic, expires)
	candidateNeed, _ := NewNeed("need-d", "owner-d", "e1", window,
		[]event.SalsaStyle{event.StyleCuban}, LevelBeginner, identity.RoleFollow,
		[]string{"sensual"}, 3, identity.VisibilityPublic, expires)

	results := Matcher{}.Match(*subject, []Candidate{{Need: *candidateNeed, Profile: *candidateProfile}}, now, map[string]bool{})
	if len(results) != 1 {
		t.Fatalf("expected one match, got %d", len(results))
	}
	r := results[0]
	if r.User.Tags == nil || len(r.User.Tags) != 1 || r.User.Tags[0] != "sensual" {
		t.Errorf("public tags should appear: %v", r.User.Tags)
	}
	if r.User.DistanceBucket == nil || *r.User.DistanceBucket != 1 {
		t.Errorf("public distance bucket should appear: %v", r.User.DistanceBucket)
	}
	// Contact and precise coordinates stay hidden until confirmation even when
	// the other public fields are opted in.
	if r.User.Contact != "" || r.User.Latitude != nil || r.User.Longitude != nil {
		t.Errorf("sensitive fields leaked before confirmation: contact=%q lat=%v lng=%v", r.User.Contact, r.User.Latitude, r.User.Longitude)
	}
}
