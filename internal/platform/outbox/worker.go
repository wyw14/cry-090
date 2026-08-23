package outbox

import (
	"context"
	"time"
)

type Message struct {
	ID          string
	Topic       string
	Payload     []byte
	Attempts    int
	AvailableAt time.Time
}
type Sender interface {
	Send(context.Context, Message) error
}
type Worker struct {
	MaxAttempts int
	Backoff     time.Duration
	DeadLetters []Message
}

func (w *Worker) Run(ctx context.Context, messages []Message, sender Sender) []Message {
	pending := append([]Message(nil), messages...)
	for i := range pending {
		if pending[i].AvailableAt.IsZero() {
			pending[i].AvailableAt = time.Now().UTC()
		}
	}
	for len(pending) > 0 {
		select {
		case <-ctx.Done():
			return pending
		default:
		}
		msg := pending[0]
		pending = pending[1:]
		if err := sender.Send(ctx, msg); err != nil {
			msg.Attempts++
			if msg.Attempts >= w.MaxAttempts {
				w.DeadLetters = append(w.DeadLetters, msg)
			} else {
				msg.AvailableAt = time.Now().UTC().Add(w.Backoff)
				pending = append(pending, msg)
			}
		}
	}
	return nil
}
