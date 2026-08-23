package registration

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestCapacityBookConcurrentRegistrationKeepsSingleSeat(t *testing.T) {
	book, err := NewCapacityBook(1)
	if err != nil {
		t.Fatal(err)
	}
	const actors = 64
	start := make(chan struct{})
	var ready sync.WaitGroup
	var done sync.WaitGroup
	ready.Add(actors)
	done.Add(actors)
	for i := 0; i < actors; i++ {
		go func(index int) {
			defer done.Done()
			ready.Done()
			<-start
			value, createErr := New(fmt.Sprintf("r-%d", index), "event", fmt.Sprintf("u-%d", index), time.Now())
			if createErr != nil {
				t.Errorf("new: %v", createErr)
				return
			}
			if _, registerErr := book.Register(value); registerErr != nil {
				t.Errorf("register: %v", registerErr)
			}
		}(i)
	}
	ready.Wait()
	close(start)
	done.Wait()
	registered, waitlisted := 0, 0
	for _, value := range book.Snapshot() {
		if value.Status == StatusRegistered {
			registered++
		}
		if value.Status == StatusWaitlisted {
			waitlisted++
		}
	}
	if registered != 1 || waitlisted != actors-1 {
		t.Fatalf("capacity was oversubscribed: registered=%d waitlisted=%d", registered, waitlisted)
	}
}
