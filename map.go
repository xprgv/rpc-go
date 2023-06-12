package rpc

import "sync"

type tsmap[K comparable, V any] struct {
	coreMap map[K]V
	mtx     sync.RWMutex
}

func newtsmap[K comparable, V any]() tsmap[K, V] {
	return tsmap[K, V]{coreMap: map[K]V{}}
}

func (m *tsmap[K, V]) Get(key K) (value V, exist bool) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	value, exist = m.coreMap[key]
	return value, exist
}

func (m *tsmap[K, V]) Set(key K, value V) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.coreMap[key] = value
}

func (m *tsmap[K, V]) Pop(key K) (value V, exist bool) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	value, exist = m.coreMap[key]
	if exist {
		delete(m.coreMap, key)
	}
	return value, exist
}

func (m *tsmap[K, V]) Delete(key K) (deleted bool) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	if _, exists := m.coreMap[key]; exists {
		delete(m.coreMap, key)
		return true
	} else {
		return false
	}
}

func (m *tsmap[K, V]) DeleteMultiple(keys ...K) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	for _, key := range keys {
		delete(m.coreMap, key)
	}
}

func (m *tsmap[K, V]) ForEach(f func(value V)) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	for _, value := range m.coreMap {
		f(value)
	}
}

func (m *tsmap[K, V]) Size() int {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	return len(m.coreMap)
}

func (m *tsmap[K, V]) Flush() {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.coreMap = make(map[K]V)
}
