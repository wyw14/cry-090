package event

import (
	"github.com/wyw14/cry-090/internal/domain/common"
	"testing"
	"time"
)

func TestEventStateMachine(t *testing.T) {
	now := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	tr := common.MustTimeRange(now.Add(time.Hour), now.Add(3*time.Hour))
	money, _ := common.NewMoney(1000, "CNY")
	e, err := New("e", "u", "v", "Social", []Session{{ID: "s", Window: tr, Styles: []SalsaStyle{StyleLA}}}, money, 2, now)
	if err != nil {
		t.Fatal(err)
	}
	if err = e.Submit(1, now); err != nil {
		t.Fatal(err)
	}
	if err = e.Approve(2, now); err != nil {
		t.Fatal(err)
	}
	if err = e.Cancel("weather", 3, now); err != nil {
		t.Fatal(err)
	}
	if e.Status != StatusCancelled || e.Registered != 0 || e.Waitlisted != 0 {
		t.Fatalf("cancel did not release capacity: %+v", e)
	}
}
