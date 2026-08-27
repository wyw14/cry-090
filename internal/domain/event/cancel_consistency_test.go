package event

import (
	"reflect"
	"testing"
	"time"

	"github.com/wyw14/cry-090/internal/domain/common"
)

func cancellableEvent(t *testing.T, now time.Time) *Event {
	t.Helper()
	price, err := common.NewMoney(1000, "CNY")
	if err != nil {
		t.Fatal(err)
	}
	value, err := New("event", "organizer", "venue", "Social", []Session{{ID: "session", Window: common.MustTimeRange(now.Add(time.Hour), now.Add(2*time.Hour)), Styles: []SalsaStyle{StyleLA}}}, price, 3, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := value.Submit(value.Version, now); err != nil {
		t.Fatal(err)
	}
	if err := value.Approve(value.Version, now); err != nil {
		t.Fatal(err)
	}
	value.Registered = 2
	value.Waitlisted = 1
	return value
}

func TestCancelReleasesCapacityWithoutPartialState(t *testing.T) {
	now := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	stale := cancellableEvent(t, now)
	before := *stale
	if err := stale.Cancel("weather", stale.Version-1, now.Add(time.Minute)); err == nil {
		t.Fatal("stale cancellation was accepted")
	}
	if !reflect.DeepEqual(*stale, before) {
		t.Fatalf("failed cancellation left partial state: before=%+v after=%+v", before, *stale)
	}
	valid := cancellableEvent(t, now)
	if err := valid.Cancel("weather", valid.Version, now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if valid.Status != StatusCancelled || valid.Registered != 0 || valid.Waitlisted != 0 {
		t.Fatalf("cancelled event retained capacity: %+v", valid)
	}
}
