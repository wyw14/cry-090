package invitation

import (
	"errors"
	"testing"
	"time"

	"github.com/wyw14/cry-090/internal/domain/common"
	"github.com/wyw14/cry-090/internal/domain/event"
)

func newPending(now time.Time) *Invitation {
	win := common.MustTimeRange(now.Add(time.Hour), now.Add(2*time.Hour))
	i, _ := New("inv-1", "sender", "recipient", "need-1", "event-1", "", win, now.Add(30*time.Minute), now)
	return i
}

func TestAcceptRequiresPending(t *testing.T) {
	now := time.Date(2026, 8, 24, 10, 0, 0, 0, time.UTC)
	i := newPending(now)
	if err := i.Accept(1, now, nil); err != nil {
		t.Fatalf("first accept: %v", err)
	}
	// A second accept (even with the stale version) must be rejected.
	if err := i.Accept(1, now, nil); err == nil {
		t.Fatal("accepted an invitation that was already processed")
	}
	// Rejection after acceptance must also fail.
	if err := i.Reject(2, now); err == nil {
		t.Fatal("rejected an invitation that was already accepted")
	}
}

func TestRejectBlocksAccept(t *testing.T) {
	now := time.Date(2026, 8, 24, 10, 0, 0, 0, time.UTC)
	i := newPending(now)
	if err := i.Reject(1, now); err != nil {
		t.Fatalf("reject: %v", err)
	}
	if err := i.Accept(2, now, nil); err == nil {
		t.Fatal("accepted an invitation that was already rejected")
	}
}

func TestExpiredCannotBeAccepted(t *testing.T) {
	now := time.Date(2026, 8, 24, 10, 0, 0, 0, time.UTC)
	i := newPending(now)
	expired := now.Add(31 * time.Minute) // past ExpiresAt
	err := i.Accept(1, expired, nil)
	if err == nil {
		t.Fatal("accepted an expired invitation")
	}
	if got := common.CodeOf(err); got != common.CodeExpired {
		t.Fatalf("want expired code, got %s", got)
	}
}

func TestVersionMismatchRejected(t *testing.T) {
	now := time.Date(2026, 8, 24, 10, 0, 0, 0, time.UTC)
	i := newPending(now)
	err := i.Accept(2, now, nil)
	if err == nil {
		t.Fatal("accepted with a stale version")
	}
	if got := common.CodeOf(err); got != common.CodeVersionConflict {
		t.Fatalf("want version_conflict code, got %s", got)
	}
}

func TestAcceptRejectsOverlappingConfirmedPartner(t *testing.T) {
	now := time.Date(2026, 8, 24, 10, 0, 0, 0, time.UTC)
	i := newPending(now)
	// An already-accepted invitation whose window overlaps the pending one.
	overlapping := event.Session{
		ID:     "inv-other",
		Window: common.MustTimeRange(now.Add(90*time.Minute), now.Add(3*time.Hour)),
	}
	err := i.Accept(1, now, []event.Session{overlapping})
	if err == nil {
		t.Fatal("accepted despite an overlapping confirmed partner")
	}
	if got := common.CodeOf(err); got != common.CodeConflict {
		t.Fatalf("want conflict code, got %s", got)
	}
	// A non-overlapping accepted partner must not block acceptance.
	other := newPending(now)
	adjacent := event.Session{
		ID:     "inv-other",
		Window: common.MustTimeRange(now.Add(2*time.Hour), now.Add(3*time.Hour)),
	}
	if err := other.Accept(1, now, []event.Session{adjacent}); err != nil {
		t.Fatalf("adjacent partner blocked accept: %v", err)
	}
}

// guard against accidental future changes that drop the error sentinel.
var _ = errors.New
