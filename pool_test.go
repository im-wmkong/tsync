package tsync

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestPool_GetPut(t *testing.T) {
	p := NewPool(func() int {
		return 1
	})

	v := p.Get()
	if v != 1 {
		t.Fatalf("expected 1, got %d", v)
	}

	p.Put(2)

	v2 := p.Get()
	if v2 != 1 && v2 != 2 {
		t.Fatalf("unexpected value %d", v2)
	}
}

func TestPool_NewCalled(t *testing.T) {
	var called atomic.Int32

	p := NewPool(func() int {
		called.Add(1)
		return 42
	})

	_ = p.Get()

	if called.Load() == 0 {
		t.Fatalf("expected New to be called at least once")
	}
}

func TestPool_Concurrent(t *testing.T) {
	p := NewPool(func() int {
		return 0
	})

	const goroutines = 10
	const iterations = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				v := p.Get()
				p.Put(v + 1)
			}
		}()
	}

	wg.Wait()
}

func TestPool_GenericType(t *testing.T) {
	type item struct {
		n int
	}

	p := NewPool(func() item {
		return item{n: 1}
	})

	v := p.Get()
	if v.n != 1 {
		t.Fatalf("unexpected value %+v", v)
	}

	p.Put(item{n: 2})

	v2 := p.Get()
	if v2.n != 1 && v2.n != 2 {
		t.Fatalf("unexpected value %+v", v2)
	}
}
