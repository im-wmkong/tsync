package tsync

import (
	"context"
	"sync"
)

type WaitGroup struct {
	wg           sync.WaitGroup
	onPanic      PanicHandler
	recoverPanic bool
	panicOnce    sync.Once
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

func (wg *WaitGroup) Go(f func()) {
	wg.wg.Add(1)
	go func() {
		defer wg.wg.Done()
		wg.run(f)
	}()
}

func (wg *WaitGroup) GoCtx(ctx context.Context, f func(ctx context.Context)) {
	wg.wg.Add(1)
	go func() {
		defer wg.wg.Done()

		select {
		case <-ctx.Done():
			return
		default:
			wg.run(func() {
				f(ctx)
			})
		}
	}()
}

func (wg *WaitGroup) Wait() {
	wg.wg.Wait()
}

func (wg *WaitGroup) run(f func()) {
	if !wg.recoverPanic {
		f()
		return
	}

	defer func() {
		if p := recover(); p != nil {
			if wg.onPanic != nil {
				wg.panicOnce.Do(func() {
					wg.onPanic(p)
				})
			}
		}
	}()

	f()
}
