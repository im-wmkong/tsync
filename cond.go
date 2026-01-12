package tsync

import (
	"context"
	"sync"
)

type Cond struct {
	mu   sync.Mutex
	cond *sync.Cond
}

func NewCond() *Cond {
	c := &Cond{}
	c.cond = sync.NewCond(&c.mu)
	return c
}

func (c *Cond) WaitUntil(predicate func() bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for !predicate() {
		c.cond.Wait()
	}
}

func (c *Cond) WaitUntilCtx(
	ctx context.Context,
	predicate func() bool,
) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	done := make(chan struct{})

	// 用一个 goroutine 监听 ctx.Done()
	go func() {
		select {
		case <-ctx.Done():
			c.mu.Lock()
			c.cond.Broadcast()
			c.mu.Unlock()
		case <-done:
		}
	}()

	defer close(done)

	for !predicate() {
		if err := ctx.Err(); err != nil {
			return err
		}
		c.cond.Wait()
	}

	return nil
}

func (c *Cond) Signal() {
	c.mu.Lock()
	c.cond.Signal()
	c.mu.Unlock()
}

func (c *Cond) Broadcast() {
	c.mu.Lock()
	c.cond.Broadcast()
	c.mu.Unlock()
}
