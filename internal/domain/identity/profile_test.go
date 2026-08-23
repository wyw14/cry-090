package identity

import (
	"testing"
	"time"
)

func TestPublicViewHidesSensitiveFields(t *testing.T) {
	p, _ := NewProfile("u", "Dancer", 3, time.Now())
	p.Contact = "+86-000"
	p.PreciseLatitude = 31.2
	p.PreciseLongitude = 121.4
	view := p.PublicView(false)
	if view.Contact != "" || view.Latitude != nil || view.Longitude != nil {
		t.Fatal("sensitive fields leaked")
	}
}
