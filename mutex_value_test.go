package tsync

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestMutexValue_Load(t *testing.T) {
	mv := NewMutexValue(10)

	if v := mv.Load(); v != 10 {
		t.Fatalf("expected 10, got %d", v)
	}
}

func TestMutexValue_Lock_Modify(t *testing.T) {
	mv := NewMutexValue(1)

	mv.Lock(func(v *int) {
		*v = 42
	})

	if v := mv.Load(); v != 42 {
		t.Fatalf("expected 42, got %d", v)
	}
}

func TestMutexValue_ConcurrentUpdate(t *testing.T) {
	mv := NewMutexValue(0)

	const goroutines = 20

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			mv.Lock(func(v *int) {
				*v++
			})
		}()
	}

	wg.Wait()

	if v := mv.Load(); v != goroutines {
		t.Fatalf("expected %d, got %d", goroutines, v)
	}
}

func TestMutexValue_Lock_IsExclusive(t *testing.T) {
	mv := NewMutexValue(0)

	var maxConcurrent atomic.Int32
	var current atomic.Int32

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			mv.Lock(func(v *int) {
				c := current.Add(1)
				for {
					m := maxConcurrent.Load()
					if c <= m || maxConcurrent.CompareAndSwap(m, c) {
						break
					}
				}

				// 保持一小段时间，放大并发冲突窗口
				*v++
				current.Add(-1)
			})
		}()
	}

	wg.Wait()

	if maxConcurrent.Load() != 1 {
		t.Fatalf("expected exclusive execution, got %d", maxConcurrent.Load())
	}
}

func TestMutexValue_GenericType(t *testing.T) {
	type data struct {
		n int
	}

	mv := NewMutexValue(data{n: 1})

	mv.Lock(func(v *data) {
		v.n = 5
	})

	if v := mv.Load(); v.n != 5 {
		t.Fatalf("unexpected value: %+v", v)
	}
}
