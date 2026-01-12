package tsync

import "sync"

type RWMutexValue[T any] struct {
	mu sync.RWMutex
	v  T
}

func NewRWMutexValue[T any](v T) *RWMutexValue[T] {
	return &RWMutexValue[T]{v: v}
}

func (m *RWMutexValue[T]) RLock(fn func(v T)) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	fn(m.v)
}

func (m *RWMutexValue[T]) Lock(fn func(v *T)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	fn(&m.v)
}
