package outbox

import (
	"context"
	"errors"
	"testing"
	"time"
)

type controlledSender struct {
	calls  int
	cancel context.CancelFunc
}

func (s *controlledSender) Send(_ context.Context, _ Message) error {
	s.calls++
	if s.calls == 1 && s.cancel != nil {
		s.cancel()
	}
	return errors.New("offline adapter unavailable")
}

func TestWorkerHonorsCancellationAndMovesExhaustedMessageToDeadLetter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	sender := &controlledSender{cancel: cancel}
	worker := &Worker{MaxAttempts: 2, Backoff: time.Millisecond}
	messages := []Message{{ID: "m1", Topic: "invite"}, {ID: "m2", Topic: "event-cancel"}}
	pending := worker.Run(ctx, messages, sender)
	if sender.calls != 1 || len(pending) != 2 {
		t.Fatalf("cancelled worker consumed more work: calls=%d pending=%d", sender.calls, len(pending))
	}
	ctx2 := context.Background()
	sender2 := &controlledSender{}
	worker2 := &Worker{MaxAttempts: 2, Backoff: time.Millisecond}
	left := worker2.Run(ctx2, []Message{{ID: "m3", Topic: "audit"}}, sender2)
	if len(left) != 0 || sender2.calls != 2 || len(worker2.DeadLetters) != 1 {
		t.Fatalf("retry/dead-letter contract broken: calls=%d pending=%d dead=%d", sender2.calls, len(left), len(worker2.DeadLetters))
	}
	if worker2.DeadLetters[0].Attempts != 2 {
		t.Fatalf("attempt count lost: %+v", worker2.DeadLetters[0])
	}
}
