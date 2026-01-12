package tsync

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestMap_StoreLoad(t *testing.T) {
	var m Map[string, int]

	m.Store("a", 1)

	v, ok := m.Load("a")
	if !ok {
		t.Fatalf("expected key to exist")
	}
	if v != 1 {
		t.Fatalf("expected 1, got %d", v)
	}
}

func TestMap_Load_NotFound(t *testing.T) {
	var m Map[string, int]

	v, ok := m.Load("missing")
	if ok {
		t.Fatalf("expected not found")
	}
	if v != 0 {
		t.Fatalf("expected zero value, got %d", v)
	}
}

func TestMap_Delete(t *testing.T) {
	var m Map[string, int]

	m.Store("a", 1)
	m.Delete("a")

	_, ok := m.Load("a")
	if ok {
		t.Fatalf("expected key to be deleted")
	}
}

func TestMap_LoadOrStore_New(t *testing.T) {
	var m Map[string, int]

	v, loaded := m.LoadOrStore("a", 10)
	if loaded {
		t.Fatalf("expected not loaded")
	}
	if v != 10 {
		t.Fatalf("expected 10, got %d", v)
	}
}

func TestMap_LoadOrStore_Existing(t *testing.T) {
	var m Map[string, int]

	m.Store("a", 1)

	v, loaded := m.LoadOrStore("a", 2)
	if !loaded {
		t.Fatalf("expected loaded")
	}
	if v != 1 {
		t.Fatalf("expected existing value 1, got %d", v)
	}
}

func TestMap_MustLoad(t *testing.T) {
	var m Map[string, int]

	m.Store("a", 5)

	if v := m.MustLoad("a"); v != 5 {
		t.Fatalf("unexpected value %d", v)
	}
}

func TestMap_Range(t *testing.T) {
	var m Map[string, int]

	m.Store("a", 1)
	m.Store("b", 2)

	sum := 0
	m.Range(func(k string, v int) bool {
		sum += v
		return true
	})

	if sum != 3 {
		t.Fatalf("expected sum=3, got %d", sum)
	}
}

func TestMap_Range_Stop(t *testing.T) {
	var m Map[string, int]

	m.Store("a", 1)
	m.Store("b", 2)

	count := 0
	m.Range(func(k string, v int) bool {
		count++
		return false
	})

	if count != 1 {
		t.Fatalf("expected range to stop early")
	}
}

func TestMap_LoadOrInit(t *testing.T) {
	var m Map[string, int]

	var called atomic.Int32

	init := func() int {
		called.Add(1)
		return 42
	}

	v1, loaded1 := m.LoadOrInit("k", init)
	v2, loaded2 := m.LoadOrInit("k", init)

	if loaded1 {
		t.Fatalf("first call should not be loaded")
	}
	if !loaded2 {
		t.Fatalf("second call should be loaded")
	}

	if v1 != 42 || v2 != 42 {
		t.Fatalf("unexpected values %d %d", v1, v2)
	}

	if called.Load() != 1 {
		t.Fatalf("init called %d times", called.Load())
	}
}

func TestMap_LoadOrInit_Concurrent(t *testing.T) {
	var m Map[int, int]

	var called atomic.Int32

	init := func() int {
		called.Add(1)
		return 100
	}

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			v, _ := m.LoadOrInit(1, init)
			if v != 100 {
				t.Errorf("unexpected value %d", v)
			}
		}()
	}

	wg.Wait()

	if called.Load() != 1 {
		t.Fatalf("init called %d times", called.Load())
	}
}
