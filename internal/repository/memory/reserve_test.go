package memory

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/wyw14/cry-090/internal/domain/common"
	"github.com/wyw14/cry-090/internal/domain/event"
	"github.com/wyw14/cry-090/internal/domain/registration"
)

// newPublishedEvent seeds an event the way the rest of the store does, bypassing
// the service layer so the repository can be exercised in isolation.
func newPublishedEvent(t *testing.T, s *Store, id, venueID string, capacity int) event.Event {
	t.Helper()
	money, err := common.NewMoney(1000, "CNY")
	if err != nil {
		t.Fatal(err)
	}
	window, err := common.NewTimeRange(time.Now().Add(time.Hour), time.Now().Add(3*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	e, err := event.New(id, "organizer", venueID, "Social",
		[]event.Session{{ID: "s1", Window: window, Styles: []event.SalsaStyle{event.StyleCuban}}},
		money, capacity, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	e.Status = event.StatusPublished
	e.Version = 1
	if err := s.Save(context.Background(), *e, 0); err != nil {
		t.Fatal(err)
	}
	return *e
}

// TestReserveConcurrentNoOversell fires capacity+surplus concurrent
// registrations against a single-remaining seat and asserts that exactly
// `capacity` land as registered and the rest land as waitlisted with unique,
// contiguous positions. The previous read-then-save Register path oversold the
// last seat here; Reserve performs the decision and insert under one lock.
func TestReserveConcurrentNoOversell(t *testing.T) {
	s := NewStore()
	capacity := 1
	e := newPublishedEvent(t, s, "e1", "v1", capacity)

	// Pre-fill the event to capacity-1 so only one seat remains.
	for i := 0; i < capacity-1; i++ {
		reg, err := registration.New("seed", e.ID, "seed-"+itoa(i), time.Now())
		if err != nil {
			t.Fatal(err)
		}
		if _, err := s.Reserve(context.Background(), *reg, capacity, 0); err != nil {
			t.Fatal(err)
		}
	}

	const contenders = 25
	var wg sync.WaitGroup
	results := make([]registration.Registration, contenders)
	errs := make([]error, contenders)
	start := make(chan struct{})
	wg.Add(contenders)
	for i := 0; i < contenders; i++ {
		i := i
		go func() {
			defer wg.Done()
			<-start
			reg, err := registration.New("registration", e.ID, "user-"+itoa(i), time.Now())
			if err != nil {
				errs[i] = err
				return
			}
			results[i], errs[i] = s.Reserve(context.Background(), *reg, capacity, 0)
		}()
	}
	close(start)
	wg.Wait()

	registered := 0
	waitlisted := 0
	positions := map[int]bool{}
	for i := 0; i < contenders; i++ {
		if errs[i] != nil {
			t.Fatalf("reserve %d failed: %v", i, errs[i])
		}
		switch results[i].Status {
		case registration.StatusRegistered:
			registered++
		case registration.StatusWaitlisted:
			waitlisted++
			if positions[results[i].Position] {
				t.Fatalf("duplicate waitlist position %d", results[i].Position)
			}
			positions[results[i].Position] = true
		default:
			t.Fatalf("unexpected status %q for result %d", results[i].Status, i)
		}
	}

	if registered != capacity {
		t.Fatalf("expected exactly %d registered, got %d (oversold)", capacity, registered)
	}
	if got := waitlisted; got != contenders-capacity {
		t.Fatalf("expected %d waitlisted, got %d", contenders-capacity, got)
	}
	for i := 1; i <= waitlisted; i++ {
		if !positions[i] {
			t.Fatalf("waitlist position %d missing (positions not contiguous)", i)
		}
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
