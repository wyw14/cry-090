package event

import (
	"reflect"
	"testing"
	"time"

	"github.com/wyw14/cry-090/internal/domain/common"
)

func changeTestEvent(t *testing.T, now time.Time) *Event {
	t.Helper()
	price, err := common.NewMoney(1000, "CNY")
	if err != nil {
		t.Fatal(err)
	}
	value, err := New("event", "organizer", "venue", "Social", []Session{{ID: "morning", Window: common.MustTimeRange(now.Add(time.Hour), now.Add(2*time.Hour)), Styles: []SalsaStyle{StyleLA}}}, price, 20, now)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func TestEventChangeRejectsStaleAndOverlappingSessionsWithoutMutation(t *testing.T) {
	now := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	stale := changeTestEvent(t, now)
	staleBefore := *stale
	staleSessions := []Session{{ID: "afternoon", Window: common.MustTimeRange(now.Add(3*time.Hour), now.Add(4*time.Hour)), Styles: []SalsaStyle{StyleNY}}}
	if err := stale.Change("Stale update", staleSessions, 30, stale.Version-1, now.Add(time.Minute)); err == nil {
		t.Fatal("stale event change was accepted")
	}
	if !reflect.DeepEqual(*stale, staleBefore) {
		t.Fatalf("stale change mutated event: before=%+v after=%+v", staleBefore, *stale)
	}
	overlap := changeTestEvent(t, now)
	overlapBefore := *overlap
	overlappingSessions := []Session{
		{ID: "first", Window: common.MustTimeRange(now.Add(time.Hour), now.Add(3*time.Hour)), Styles: []SalsaStyle{StyleLA}},
		{ID: "second", Window: common.MustTimeRange(now.Add(2*time.Hour), now.Add(4*time.Hour)), Styles: []SalsaStyle{StyleCuban}},
	}
	if err := overlap.Change("Overlapping schedule", overlappingSessions, 20, overlap.Version, now.Add(time.Minute)); err == nil {
		t.Fatal("overlapping event sessions were accepted")
	}
	if !reflect.DeepEqual(*overlap, overlapBefore) {
		t.Fatalf("overlapping change mutated event: before=%+v after=%+v", overlapBefore, *overlap)
	}
}
