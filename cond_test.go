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

func TestCond_WaitUntilCtx_ReturnNil(t *testing.T) {
	c := NewCond()
	var ready atomic.Bool

	// 创建一个不会被取消的上下文
	ctx := context.Background()

	go func() {
		time.Sleep(50 * time.Millisecond)
		ready.Store(true)
		c.Broadcast()
	}()

	err := c.WaitUntilCtx(ctx, func() bool {
		return ready.Load()
	})

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
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

	// 使用公共的Signal方法而不是直接调用内部的cond.Signal()
	time.Sleep(20 * time.Millisecond) // 确保goroutine已经开始等待
	count.Store(1)
	c.Signal()

	time.Sleep(20 * time.Millisecond)

	if count.Load() != 2 {
		t.Fatalf("signal did not wake waiter")
	}
}
