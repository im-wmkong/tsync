package tsync

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestAtomicValue_LoadStore(t *testing.T) {
	a := NewAtomicValue(10)

	if v := a.Load(); v != 10 {
		t.Fatalf("expected 10, got %d", v)
	}

	a.Store(20)

	if v := a.Load(); v != 20 {
		t.Fatalf("expected 20, got %d", v)
	}
}

func TestAtomicValue_Swap(t *testing.T) {
	a := NewAtomicValue(1)

	old := a.Swap(2)
	if old != 1 {
		t.Fatalf("expected old=1, got %d", old)
	}

	if v := a.Load(); v != 2 {
		t.Fatalf("expected 2, got %d", v)
	}
}

func TestAtomicValue_ConcurrentLoad(t *testing.T) {
	a := NewAtomicValue(42)

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			if v := a.Load(); v != 42 {
				t.Errorf("unexpected value %d", v)
			}
		}()
	}

	wg.Wait()
}

func TestAtomicValue_ConcurrentStoreLoad(t *testing.T) {
	a := NewAtomicValue(0)

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(v int) {
			defer wg.Done()
			a.Store(v)
			_ = a.Load()
		}(i)
	}

	wg.Wait()
}

func TestAtomicValue_GenericType(t *testing.T) {
	type config struct {
		Version int
		Name    string
	}

	a := NewAtomicValue(config{
		Version: 1,
		Name:    "v1",
	})

	cfg := a.Load()
	if cfg.Version != 1 || cfg.Name != "v1" {
		t.Fatalf("unexpected value %+v", cfg)
	}

	a.Store(config{
		Version: 2,
		Name:    "v2",
	})

	cfg = a.Load()
	if cfg.Version != 2 || cfg.Name != "v2" {
		t.Fatalf("unexpected value %+v", cfg)
	}
}

func TestAtomicValue_ConcurrentSwap(t *testing.T) {
	a := NewAtomicValue(0)

	const goroutines = 10
	var swaps atomic.Int32

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(v int) {
			defer wg.Done()
			_ = a.Swap(v)
			swaps.Add(1)
		}(i)
	}

	wg.Wait()

	if swaps.Load() != goroutines {
		t.Fatalf("expected %d swaps, got %d", goroutines, swaps.Load())
	}
}

func TestAtomicValue_Load_Uninitialized_Panic(t *testing.T) {
	var a AtomicValue[int]

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic")
		}
	}()

	_ = a.Load()
}
