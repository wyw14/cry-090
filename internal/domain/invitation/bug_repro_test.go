package invitation

import (
	"testing"
	"time"

	"github.com/wyw14/cry-090/internal/domain/common"
	"github.com/wyw14/cry-090/internal/domain/event"
)

func TestInvitationProcessingIsSingleUseAndConflictAware(t *testing.T) {
	now := time.Date(2026, 8, 23, 8, 0, 0, 0, time.UTC)
	window := common.MustTimeRange(now.Add(time.Hour), now.Add(3*time.Hour))
	value, err := New("invite-1", "sender", "recipient", "need", "event", "dance?", window, now.Add(30*time.Minute), now)
	if err != nil {
		t.Fatal(err)
	}
	conflicts := []event.Session{{ID: "accepted-other", Window: window}}
	if err := value.Accept(1, now.Add(time.Minute), conflicts); common.CodeOf(err) != common.CodeConflict {
		t.Fatalf("overlap should be rejected, got %v", err)
	}
	if value.Status != StatusPending || value.Version != 1 {
		t.Fatalf("failed accept changed state: %+v", value)
	}
	if err := value.Accept(1, now.Add(time.Minute), nil); err != nil {
		t.Fatal(err)
	}
	if err := value.Reject(2, now.Add(2*time.Minute)); common.CodeOf(err) != common.CodeAlreadyProcessed {
		t.Fatalf("second processing should fail, got %v", err)
	}
	expired, _ := New("invite-2", "sender", "recipient", "need", "event", "dance?", window, now.Add(time.Minute), now)
	if err := expired.Accept(1, now.Add(2*time.Minute), nil); common.CodeOf(err) != common.CodeExpired {
		t.Fatalf("expired invitation should fail, got %v", err)
	}
}
