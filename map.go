package tsync

import "sync"

type Map[K comparable, V any] struct {
	m sync.Map
}

func (m *Map[K, V]) Load(key K) (value V, ok bool) {
	v, ok := m.m.Load(key)
	if !ok {
		var zero V
		return zero, false
	}
	return v.(V), true
}

func (m *Map[K, V]) Store(key K, value V) {
	m.m.Store(key, value)
}

func (m *Map[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	v, loaded := m.m.LoadOrStore(key, value)
	return v.(V), loaded
}

func (m *Map[K, V]) Delete(key K) {
	m.m.Delete(key)
}

func (m *Map[K, V]) Range(fn func(key K, value V) bool) {
	m.m.Range(func(k, v any) bool {
		return fn(k.(K), v.(V))
	})
}

func (m *Map[K, V]) MustLoad(key K) V {
	v, ok := m.Load(key)
	if !ok {
		panic("tsync.Map: key not found")
	}
	return v
}

func (m *Map[K, V]) LoadOrInit(key K, init func() V) (value V, loaded bool) {
	actual, loaded := m.m.Load(key)
	if loaded {
		return actual.(V), true
	}

	v := init()
	actual, loaded = m.m.LoadOrStore(key, v)
	return actual.(V), loaded
}
