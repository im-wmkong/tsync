package tsync

import "sync"

type MutexValue[T any] struct {
	mu sync.Mutex
	v  T
}

func NewMutexValue[T any](v T) *MutexValue[T] {
	return &MutexValue[T]{v: v}
}

func (m *MutexValue[T]) Lock(fn func(v *T)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	fn(&m.v)
}

func (m *MutexValue[T]) Load() T {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.v
}
