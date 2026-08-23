package integration

import (
	"context"
	"testing"
	"time"

	"github.com/wyw14/cry-090/internal/bootstrap"
	"github.com/wyw14/cry-090/internal/domain/common"
	"github.com/wyw14/cry-090/internal/domain/event"
	"github.com/wyw14/cry-090/internal/domain/identity"
)

func TestDemoAppHealthAndDomainSeed(t *testing.T) {
	app := bootstrap.NewApp()
	if app.Server.Router() == nil {
		t.Fatal("router missing")
	}
	profile, err := identity.NewProfile("u1", "Mina", 2, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := app.Store.SaveProfile(context.Background(), *profile, 0); err != nil {
		t.Fatal(err)
	}
	price, _ := common.NewMoney(6800, "cny")
	rangeValue, _ := common.NewTimeRange(time.Now().Add(time.Hour), time.Now().Add(3*time.Hour))
	value, err := event.New("e1", "u1", "v1", "Friday Social", []event.Session{{ID: "s1", Window: rangeValue, Styles: []event.SalsaStyle{event.StyleCuban}}}, price, 30, time.Now())
	if err != nil || value.Status != event.StatusDraft {
		t.Fatalf("unexpected event: %v", err)
	}
}
