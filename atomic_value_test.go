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

func TestAtomicValue_CompareAndSwap(t *testing.T) {
	a := NewAtomicValue(10)

	swapped := a.CompareAndSwap(10, 20)
	if !swapped {
		t.Fatalf("expected swap to succeed")
	}
	if v := a.Load(); v != 20 {
		t.Fatalf("expected 20, got %d", v)
	}

	swapped = a.CompareAndSwap(10, 30)
	if swapped {
		t.Fatalf("expected swap to fail")
	}
	if v := a.Load(); v != 20 {
		t.Fatalf("expected 20, got %d", v)
	}
}

func TestAtomicValue_ConcurrentCompareAndSwap(t *testing.T) {
	a := NewAtomicValue(0)

	const goroutines = 100
	const iterations = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				old := a.Load()
				newVal := old + 1
				a.CompareAndSwap(old, newVal)
			}
		}()
	}

	wg.Wait()

	finalVal := a.Load()
	if finalVal < 0 || finalVal > goroutines*iterations {
		t.Fatalf("unexpected final value %d", finalVal)
	}
}

func TestAtomicValue_CompareAndSwap_GenericType(t *testing.T) {
	type user struct {
		ID   int
		Name string
	}

	a := NewAtomicValue(user{
		ID:   1,
		Name: "user1",
	})

	oldUser := user{ID: 1, Name: "user1"}
	newUser := user{ID: 2, Name: "user2"}
	swapped := a.CompareAndSwap(oldUser, newUser)
	if !swapped {
		t.Fatalf("expected swap to succeed")
	}

	currentUser := a.Load()
	if currentUser.ID != 2 || currentUser.Name != "user2" {
		t.Fatalf("unexpected value %+v", currentUser)
	}

	swapped = a.CompareAndSwap(oldUser, newUser)
	if swapped {
		t.Fatalf("expected swap to fail")
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
