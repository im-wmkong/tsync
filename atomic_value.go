package tsync

import "sync/atomic"

type AtomicValue[T any] struct {
	v atomic.Value
}

func NewAtomicValue[T any](v T) *AtomicValue[T] {
	a := &AtomicValue[T]{}
	a.v.Store(v)
	return a
}

func (a *AtomicValue[T]) Load() T {
	return a.v.Load().(T)
}

func (a *AtomicValue[T]) Store(v T) {
	a.v.Store(v)
}

func (a *AtomicValue[T]) Swap(v T) (old T) {
	old = a.Load()
	a.Store(v)
	return old
}
