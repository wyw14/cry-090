package memory

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

type Clock struct{ fixed atomic.Int64 }

func NewClock(now time.Time) *Clock { c := &Clock{}; c.fixed.Store(now.UnixNano()); return c }
func (c *Clock) Now() time.Time     { return time.Unix(0, c.fixed.Load()).UTC() }
func (c *Clock) Set(now time.Time)  { c.fixed.Store(now.UnixNano()) }

type IDs struct{ next atomic.Uint64 }

func (i *IDs) NewID(prefix string) string { return fmt.Sprintf("%s-%08d", prefix, i.next.Add(1)) }

type Tx struct{}

func (Tx) Within(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }
