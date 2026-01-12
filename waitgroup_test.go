package tsync

import (
	"context"
	"sync/atomic"
	"testing"
)

func TestWaitGroup_GoWait(t *testing.T) {
	wg := NewWaitGroup()

	var v atomic.Int32

	wg.Go(func() {
		v.Add(1)
	})

	wg.Wait()

	if v.Load() != 1 {
		t.Fatalf("expected v=1, got %d", v.Load())
	}
}

func TestWaitGroup_MultipleGo(t *testing.T) {
	wg := NewWaitGroup()

	const n = 10
	var v atomic.Int32

	for i := 0; i < n; i++ {
		wg.Go(func() {
			v.Add(1)
		})
	}

	wg.Wait()

	if v.Load() != n {
		t.Fatalf("expected %d, got %d", n, v.Load())
	}
}

func TestWaitGroup_GoCtx_Canceled(t *testing.T) {
	wg := NewWaitGroup()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var called atomic.Bool

	wg.GoCtx(ctx, func(ctx context.Context) {
		called.Store(true)
	})

	wg.Wait()

	if called.Load() {
		t.Fatalf("function should not be called when context is canceled")
	}
}

func TestWaitGroup_GoCtx_Run(t *testing.T) {
	wg := NewWaitGroup()

	ctx := context.Background()
	var called atomic.Bool

	wg.GoCtx(ctx, func(ctx context.Context) {
		called.Store(true)
	})

	wg.Wait()

	if !called.Load() {
		t.Fatalf("function should be called")
	}
}

func TestWaitGroup_PanicRecovery(t *testing.T) {
	var handled atomic.Bool

	wg := NewWaitGroup(
		WithPanicRecovery(func(p any) {
			handled.Store(true)
		}),
	)

	wg.Go(func() {
		panic("boom")
	})

	wg.Wait()

	if !handled.Load() {
		t.Fatalf("panic handler was not called")
	}
}

func TestWaitGroup_PanicHandler_Once(t *testing.T) {
	var count atomic.Int32

	wg := NewWaitGroup(
		WithPanicRecovery(func(p any) {
			count.Add(1)
		}),
	)

	const n = 5
	for i := 0; i < n; i++ {
		wg.Go(func() {
			panic("boom")
		})
	}

	wg.Wait()

	if count.Load() != 1 {
		t.Fatalf("expected handler to be called once, got %d", count.Load())
	}
}

func TestWaitGroup_PanicRecovery_NilHandler(t *testing.T) {
	wg := NewWaitGroup(
		WithPanicRecovery(nil),
	)

	wg.Go(func() {
		panic("boom")
	})

	wg.Wait()
}
