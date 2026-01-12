package tsync

import "sync/atomic"

type AtomicPointer[T any] struct {
	p atomic.Pointer[T]
}

func NewAtomicPointer[T any](v *T) *AtomicPointer[T] {
	var a AtomicPointer[T]
	if v != nil {
		a.p.Store(v)
	}
	return &a
}

func (a *AtomicPointer[T]) Load() *T {
	return a.p.Load()
}

func (a *AtomicPointer[T]) Store(v *T) {
	a.p.Store(v)
}

func (a *AtomicPointer[T]) Swap(v *T) (old *T) {
	return a.p.Swap(v)
}

func (a *AtomicPointer[T]) CompareAndSwap(old, new *T) bool {
	return a.p.CompareAndSwap(old, new)
}
