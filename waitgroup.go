package tsync

import (
	"context"
	"sync"
)

type WaitGroup struct {
	wg           sync.WaitGroup
	onPanic      PanicHandler
	recoverPanic bool
}

type PanicHandler func(p any)

type WaitGroupOption func(*WaitGroup)

func WithPanicRecovery(handler PanicHandler) WaitGroupOption {
	return func(wg *WaitGroup) {
		wg.recoverPanic = true
		wg.onPanic = handler
	}
}

func NewWaitGroup(opts ...WaitGroupOption) *WaitGroup {
	wg := &WaitGroup{}
	for _, opt := range opts {
		opt(wg)
	}
	return wg
}

func (wg *WaitGroup) Go(fn func()) {
	wg.wg.Add(1)
	go func() {
		defer wg.wg.Done()
		wg.run(fn)
	}()
}

func (wg *WaitGroup) GoCtx(ctx context.Context, fn func(ctx context.Context)) {
	if ctx.Err() != nil {
		return
	}

	wg.wg.Add(1)
	go func() {
		defer wg.wg.Done()

		select {
		case <-ctx.Done():
			return
		default:
			wg.run(func() {
				fn(ctx)
			})
		}
	}()
}

func (wg *WaitGroup) Wait() {
	wg.wg.Wait()
}

func (wg *WaitGroup) run(fn func()) {
	if !wg.recoverPanic {
		fn()
		return
	}

	defer func() {
		if p := recover(); p != nil {
			if wg.onPanic != nil {
				wg.onPanic(p)
			}
		}
	}()

	fn()
}
