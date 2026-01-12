package tsync

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRWMutexValue_RLock(t *testing.T) {
	mv := NewRWMutexValue(10)

	var got int
	mv.RLock(func(v int) {
		got = v
	})

	if got != 10 {
		t.Fatalf("expected 10, got %d", got)
	}
}

func TestRWMutexValue_Lock(t *testing.T) {
	mv := NewRWMutexValue(1)

	mv.Lock(func(v *int) {
		*v = 42
	})

	var got int
	mv.RLock(func(v int) {
		got = v
	})

	if got != 42 {
		t.Fatalf("expected 42, got %d", got)
	}
}

func TestRWMutexValue_ConcurrentRead(t *testing.T) {
	mv := NewRWMutexValue(0)

	const readers = 10

	var current atomic.Int32
	var max atomic.Int32

	var wg sync.WaitGroup
	wg.Add(readers)

	for i := 0; i < readers; i++ {
		go func() {
			defer wg.Done()

			mv.RLock(func(v int) {
				c := current.Add(1)
				for {
					m := max.Load()
					if c <= m || max.CompareAndSwap(m, c) {
						break
					}
				}

				time.Sleep(10 * time.Millisecond)
				current.Add(-1)
			})
		}()
	}

	wg.Wait()

	if max.Load() <= 1 {
		t.Fatalf("expected concurrent readers, got max=%d", max.Load())
	}
}

func TestRWMutexValue_Lock_IsExclusive(t *testing.T) {
	mv := NewRWMutexValue(0)

	const writers = 5

	var current atomic.Int32
	var max atomic.Int32

	var wg sync.WaitGroup
	wg.Add(writers)

	for i := 0; i < writers; i++ {
		go func() {
			defer wg.Done()

			mv.Lock(func(v *int) {
				c := current.Add(1)
				for {
					m := max.Load()
					if c <= m || max.CompareAndSwap(m, c) {
						break
					}
				}

				time.Sleep(10 * time.Millisecond)
				*v++
				current.Add(-1)
			})
		}()
	}

	wg.Wait()

	if max.Load() != 1 {
		t.Fatalf("expected exclusive writers, got max=%d", max.Load())
	}
}

func TestRWMutexValue_WriteBlocksRead(t *testing.T) {
	mv := NewRWMutexValue(0)

	started := make(chan struct{})
	done := make(chan struct{})

	go func() {
		mv.Lock(func(v *int) {
			close(started)
			time.Sleep(50 * time.Millisecond)
			*v = 1
		})
		close(done)
	}()

	<-started

	readEntered := make(chan struct{})

	go func() {
		mv.RLock(func(v int) {
			close(readEntered)
		})
	}()

	select {
	case <-readEntered:
		t.Fatalf("read should be blocked by write lock")
	case <-time.After(20 * time.Millisecond):
		// expected
	}

	<-done

	select {
	case <-readEntered:
		// ok
	case <-time.After(50 * time.Millisecond):
		t.Fatalf("read was not unblocked after write")
	}
}

func TestRWMutexValue_GenericType(t *testing.T) {
	type data struct {
		n int
	}

	mv := NewRWMutexValue(data{n: 1})

	mv.Lock(func(v *data) {
		v.n = 10
	})

	var got int
	mv.RLock(func(v data) {
		got = v.n
	})

	if got != 10 {
		t.Fatalf("unexpected value: %d", got)
	}
}
