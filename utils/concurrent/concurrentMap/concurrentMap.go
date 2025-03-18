package concurrentMap

import (
	"sync"
)

// ConcurrentMap is a concurrent-safe map.
type ConcurrentMap[K comparable, V any] struct {
	mu    sync.RWMutex // Read/write mutex for synchronization
	items map[K]V      // The underlying map
}

// NewConcurrentMap creates a new ConcurrentMap.
func NewConcurrentMap[K comparable, V any]() *ConcurrentMap[K, V] {
	return &ConcurrentMap[K, V]{
		items: make(map[K]V),
	}
}

// Get retrieves the value associated with the given key.
func (m *ConcurrentMap[K, V]) Get(key K) (V, bool) {
	m.mu.RLock()         // Acquire read lock
	defer m.mu.RUnlock() // Release read lock

	value, ok := m.items[key]
	return value, ok
}

// Set sets the value associated with the given key.
func (m *ConcurrentMap[K, V]) Set(key K, value V) {
	m.mu.Lock()         // Acquire write lock
	defer m.mu.Unlock() // Release write lock

	m.items[key] = value
}

// Delete removes the key-value pair associated with the given key.
func (m *ConcurrentMap[K, V]) Delete(key K) {
	m.mu.Lock()         // Acquire write lock
	defer m.mu.Unlock() // Release write lock

	delete(m.items, key)
}

// Range iterates over the map.  It's important to note that while
// iterating, other writes might happen.  If you need a truly consistent
// snapshot, you should consider copying the map (using a lock).
func (m *ConcurrentMap[K, V]) Range(f func(key K, value V)) {
	m.mu.RLock() // Read lock for consistent iteration
	defer m.mu.RUnlock()

	for key, value := range m.items {
		f(key, value)
	}
}

// Len returns the number of items in the map.
func (m *ConcurrentMap[K, V]) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.items)
}

// Items returns a copy of all items in the map. This is safer for concurrent access, but uses more memory.
func (m *ConcurrentMap[K, V]) Items() map[K]V {
	m.mu.RLock()
	defer m.mu.RUnlock()

	itemsCopy := make(map[K]V, len(m.items))
	for k, v := range m.items {
		itemsCopy[k] = v
	}
	return itemsCopy
}
