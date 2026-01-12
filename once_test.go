package tsync

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestOnceValue_Get(t *testing.T) {
	ov := NewOnceValue(func() int {
		return 42
	})

	v := ov.Get()
	if v != 42 {
		t.Fatalf("expected 42, got %d", v)
	}
}

func TestOnceValue_Get_MultipleTimes(t *testing.T) {
	var calls atomic.Int32

	ov := NewOnceValue(func() int {
		calls.Add(1)
		return 100
	})

	for i := 0; i < 10; i++ {
		if v := ov.Get(); v != 100 {
			t.Fatalf("unexpected value: %d", v)
		}
	}

	if calls.Load() != 1 {
		t.Fatalf("init function called %d times, want 1", calls.Load())
	}
}

func TestOnceValue_Get_Concurrent(t *testing.T) {
	var calls atomic.Int32

	ov := NewOnceValue(func() int {
		calls.Add(1)
		return 7
	})

	const goroutines = 20

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			if v := ov.Get(); v != 7 {
				t.Errorf("unexpected value: %d", v)
			}
		}()
	}

	wg.Wait()

	if calls.Load() != 1 {
		t.Fatalf("init function called %d times, want 1", calls.Load())
	}
}

func TestOnceValue_GenericType(t *testing.T) {
	type data struct {
		x int
	}

	ov := NewOnceValue(func() data {
		return data{x: 10}
	})

	v := ov.Get()
	if v.x != 10 {
		t.Fatalf("unexpected value: %+v", v)
	}
}

func TestOnceValue_New_NilFuncPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic for nil init function")
		}
	}()

	_ = NewOnceValue[int](nil)
}
