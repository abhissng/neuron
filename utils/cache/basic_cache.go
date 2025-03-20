package cache

import (
	"sync"
	"time"
)

// cacheItem represents an item in the cache with optional expiration
type cacheItem[V any] struct {
	value  V
	expiry *time.Time // Pointer so nil means no expiry
}

// BasicCache implements the Cache interface using a simple in-memory map.
// It provides basic caching functionality with thread-safe operations using a mutex.
type BasicCache[K comparable, V any] struct {
	store           map[K]cacheItem[V] // Underlying map to store key-value pairs with expiry.
	mu              sync.Mutex         // Mutex for thread-safe access to the store.
	cleanupInterval time.Duration      // Interval for cleanup of expired items
	stopCleanup     chan bool          // Channel to signal cleanup goroutine to stop
}

// NewBasicCache creates a new instance of BasicCache.
func NewBasicCache[K comparable, V any]() Cache[K, V] {
	cache := &BasicCache[K, V]{
		store:           make(map[K]cacheItem[V]),
		cleanupInterval: 5 * time.Minute, // Default cleanup interval
		stopCleanup:     make(chan bool),
	}

	// Start the cleanup goroutine
	go cache.startCleanup()

	return cache
}

// NewBasicCacheWithCleanupInterval creates a new instance of BasicCache with a custom cleanup interval.
func NewBasicCacheWithCleanupInterval[K comparable, V any](cleanupInterval time.Duration) Cache[K, V] {
	cache := &BasicCache[K, V]{
		store:           make(map[K]cacheItem[V]),
		cleanupInterval: cleanupInterval,
		stopCleanup:     make(chan bool),
	}

	// Start the cleanup goroutine
	go cache.startCleanup()

	return cache
}

// Set implements the Cache interface method.
// It sets the value associated with the given key in the cache with no expiry.
func (c *BasicCache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = cacheItem[V]{value: value, expiry: nil}
}

// SetWithExpiry implements the Cache interface method.
// It sets the value associated with the given key in the cache with an expiry time.
func (c *BasicCache[K, V]) SetWithExpiry(key K, value V, expiry time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	expiryTime := time.Now().Add(expiry)
	c.store[key] = cacheItem[V]{value: value, expiry: &expiryTime}
}

// Get implements the Cache interface method.
// It retrieves the value associated with the given key from the cache.
// Returns the value and a boolean indicating whether the key exists and is not expired.
func (c *BasicCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.store[key]
	if !exists {
		var zero V
		return zero, false
	}

	// Check if item is expired
	if item.expiry != nil && time.Now().After(*item.expiry) {
		delete(c.store, key)
		var zero V
		return zero, false
	}

	return item.value, true
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
	c.store = make(map[K]cacheItem[V]) // Create a new empty map to clear the existing one
}

// Len implements the Cache interface method.
// It returns the number of entries currently in the cache (including expired ones that haven't been cleaned up yet).
func (c *BasicCache[K, V]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.store)
}

// startCleanup starts a goroutine that periodically removes expired items from the cache.
func (c *BasicCache[K, V]) startCleanup() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanupExpired()
		case <-c.stopCleanup:
			return
		}
	}
}

// cleanupExpired removes all expired items from the cache.
func (c *BasicCache[K, V]) cleanupExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.store {
		if item.expiry != nil && now.After(*item.expiry) {
			delete(c.store, key)
		}
	}
}

// StopCleanup stops the background cleanup goroutine.
// This should be called when the cache is no longer needed to prevent resource leaks.
func (c *BasicCache[K, V]) StopCleanup() {
	c.stopCleanup <- true
}
