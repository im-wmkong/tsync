package tsync

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestCond_WaitUntil(t *testing.T) {
	c := NewCond()

	var ready atomic.Bool

	go func() {
		time.Sleep(50 * time.Millisecond)
		ready.Store(true)
		c.Broadcast()
	}()

	c.WaitUntil(func() bool {
		return ready.Load()
	})
}

func TestCond_WaitUntilCtx_Cancel(t *testing.T) {
	c := NewCond()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := c.WaitUntilCtx(ctx, func() bool {
		return false
	})

	if err == nil {
		t.Fatalf("expected context cancellation error")
	}
}

func TestCond_Signal(t *testing.T) {
	c := NewCond()

	var count atomic.Int32

	go func() {
		c.WaitUntil(func() bool {
			return count.Load() == 1
		})
		count.Add(1)
	}()

	time.Sleep(20 * time.Millisecond)
	count.Store(1)
	c.Signal()

	time.Sleep(20 * time.Millisecond)

	if count.Load() != 2 {
		t.Fatalf("signal did not wake waiter")
	}
}
