package basicCache

import "sync"

// Cache defines an interface for generic caching operations.
//
// It provides methods for setting, getting, deleting, clearing, and retrieving the length of the cache.
type Cache[K comparable, V any] interface {
	Set(key K, value V)  // Set the value associated with the given key.
	Get(key K) (V, bool) // Get the value associated with the given key, and a boolean indicating whether the key exists.
	Delete(key K)        // Delete the entry associated with the given key.
	Clear()              // Clear all entries from the cache.
	Len() int            // Return the number of entries currently in the cache.
}

// BasicCache implements the Cache interface using a simple in-memory map.
// It provides basic caching functionality with thread-safe operations using a mutex.
type BasicCache[K comparable, V any] struct {
	store map[K]V    // Underlying map to store key-value pairs.
	mu    sync.Mutex // Mutex for thread-safe access to the store.
}

// NewBasicCache creates a new instance of BasicCache.
func NewBasicCache[K comparable, V any]() Cache[K, V] {
	return &BasicCache[K, V]{
		store: make(map[K]V),
	}
}

// Set implements the Cache interface method.
// It sets the value associated with the given key in the cache.
func (c *BasicCache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = value
}

// Get implements the Cache interface method.
// It retrieves the value associated with the given key from the cache.
// Returns the value and a boolean indicating whether the key exists.
func (c *BasicCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	value, exists := c.store[key]
	return value, exists
}

// Delete implements the Cache interface method.
// It removes the entry associated with the given key from the cache.
func (c *BasicCache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, key)
}

// Clear implements the Cache interface method.
// It removes all entries from the cache.
func (c *BasicCache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store = make(map[K]V) // Create a new empty map to clear the existing one
}

// Len implements the Cache interface method.
// It returns the number of entries currently in the cache.
func (c *BasicCache[K, V]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.store)
}
