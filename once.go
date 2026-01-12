package tsync

import "sync"

type OnceValue[T any] struct {
	once sync.Once
	fn   func() T
	v    T
}

func NewOnceValue[T any](fn func() T) *OnceValue[T] {
	if fn == nil {
		panic("tsync.OnceValue: nil init function")
	}
	return &OnceValue[T]{fn: fn}
}

func (o *OnceValue[T]) Get() T {
	o.once.Do(func() {
		o.v = o.fn()
		o.fn = nil
	})
	return o.v
}
