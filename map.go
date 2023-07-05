package rpc

import "sync"

type threadSafeMap[K comparable, V any] struct {
	coreMap map[K]V
	mtx     sync.RWMutex
}

func newtsmap[K comparable, V any]() threadSafeMap[K, V] {
	return threadSafeMap[K, V]{coreMap: map[K]V{}}
}

func (m *threadSafeMap[K, V]) Get(key K) (value V, exist bool) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	value, exist = m.coreMap[key]
	return value, exist
}

func (m *threadSafeMap[K, V]) Set(key K, value V) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.coreMap[key] = value
}

func (m *threadSafeMap[K, V]) Pop(key K) (value V, exist bool) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	value, exist = m.coreMap[key]
	if exist {
		delete(m.coreMap, key)
	}
	return value, exist
}

func (m *threadSafeMap[K, V]) Delete(key K) (deleted bool) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	if _, exists := m.coreMap[key]; exists {
		delete(m.coreMap, key)
		return true
	} else {
		return false
	}
}

func (m *threadSafeMap[K, V]) DeleteMultiple(keys ...K) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	for _, key := range keys {
		delete(m.coreMap, key)
	}
}

func (m *threadSafeMap[K, V]) ForEach(f func(value V)) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	for _, value := range m.coreMap {
		f(value)
	}
}

func (m *threadSafeMap[K, V]) Size() int {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	return len(m.coreMap)
}

func (m *threadSafeMap[K, V]) Flush() {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.coreMap = make(map[K]V)
}
