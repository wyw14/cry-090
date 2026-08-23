package identity

import (
	"testing"
	"time"
)

func TestPublicViewKeepsUnconfirmedMatchFieldsPrivate(t *testing.T) {
	profile, err := NewProfile("u-private", "Mina", 7, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	profile.Bio = "private practice notes"
	profile.Roles = []DanceRole{RoleFollow}
	profile.Styles = []string{"ny"}
	profile.Tags = []string{"musicality"}
	profile.DistanceBucket = 2
	profile.Contact = "+86-13800000000"
	profile.PreciseLatitude = 31.214
	profile.PreciseLongitude = 121.452
	profile.PublicFields = map[PublicField]bool{FieldDisplayName: true}

	view := profile.PublicView(false)
	if view.DisplayName != "Mina" {
		t.Fatalf("display name missing: %+v", view)
	}
	if view.Bio != "" || view.DanceYears != nil || len(view.Roles) != 0 || len(view.Styles) != 0 || len(view.Tags) != 0 {
		t.Fatalf("non-public profile fields leaked: %+v", view)
	}
	if view.DistanceBucket != nil || view.Contact != "" || view.Latitude != nil || view.Longitude != nil {
		t.Fatalf("unconfirmed sensitive fields leaked: %+v", view)
	}
}
