package tsync

import (
	"sync"
	"testing"
)

func TestAtomicPointer_ZeroValue(t *testing.T) {
	var p AtomicPointer[int]

	if v := p.Load(); v != nil {
		t.Fatalf("expected nil, got %v", v)
	}
}

func TestAtomicPointer_New(t *testing.T) {
	v := 10
	p := NewAtomicPointer(&v)

	got := p.Load()
	if got == nil || *got != 10 {
		t.Fatalf("unexpected value %v", got)
	}
}

func TestAtomicPointer_StoreLoad(t *testing.T) {
	var p AtomicPointer[int]

	v1 := 1
	p.Store(&v1)

	got := p.Load()
	if got == nil || *got != 1 {
		t.Fatalf("expected 1, got %v", got)
	}

	v2 := 2
	p.Store(&v2)

	got = p.Load()
	if got == nil || *got != 2 {
		t.Fatalf("expected 2, got %v", got)
	}
}

func TestAtomicPointer_Swap(t *testing.T) {
	var p AtomicPointer[int]

	v1 := 1
	old := p.Swap(&v1)
	if old != nil {
		t.Fatalf("expected old nil, got %v", old)
	}

	v2 := 2
	old = p.Swap(&v2)
	if old == nil || *old != 1 {
		t.Fatalf("expected old 1, got %v", old)
	}

	got := p.Load()
	if got == nil || *got != 2 {
		t.Fatalf("expected 2, got %v", got)
	}
}

func TestAtomicPointer_CompareAndSwap_Success(t *testing.T) {
	var p AtomicPointer[int]

	v1 := 1
	ok := p.CompareAndSwap(nil, &v1)
	if !ok {
		t.Fatalf("expected CAS success")
	}

	got := p.Load()
	if got == nil || *got != 1 {
		t.Fatalf("unexpected value %v", got)
	}
}

func TestAtomicPointer_CompareAndSwap_Failure(t *testing.T) {
	var p AtomicPointer[int]

	v1 := 1
	v2 := 2

	p.Store(&v1)

	ok := p.CompareAndSwap(&v2, &v1)
	if ok {
		t.Fatalf("expected CAS failure")
	}

	got := p.Load()
	if got == nil || *got != 1 {
		t.Fatalf("unexpected value %v", got)
	}
}

func TestAtomicPointer_ConcurrentLoadStore(t *testing.T) {
	var p AtomicPointer[int]

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(v int) {
			defer wg.Done()
			p.Store(&v)
			_ = p.Load()
		}(i)
	}

	wg.Wait()
}

func TestAtomicPointer_ConcurrentCAS(t *testing.T) {
	var p AtomicPointer[int]

	v := 0
	p.Store(&v)

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			old := p.Load()
			if old != nil {
				p.CompareAndSwap(old, old)
			}
		}()
	}

	wg.Wait()
}

func TestAtomicPointer_GenericType(t *testing.T) {
	type data struct {
		n int
	}

	d1 := &data{n: 1}
	d2 := &data{n: 2}

	p := NewAtomicPointer(d1)

	got := p.Load()
	if got == nil || got.n != 1 {
		t.Fatalf("unexpected value %+v", got)
	}

	ok := p.CompareAndSwap(d1, d2)
	if !ok {
		t.Fatalf("expected CAS success")
	}

	got = p.Load()
	if got == nil || got.n != 2 {
		t.Fatalf("unexpected value %+v", got)
	}
}
