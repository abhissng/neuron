package cache

import (
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

// lruCacheItem wraps values with expiry time
type lruCacheItem[V any] struct {
	value  V
	expiry *time.Time // nil means no expiry
}

// LRUCache provides a thread-safe wrapper around the `lru.Cache` library.
// It implements a simple Least Recently Used (LRU) cache with a fixed capacity.
//
// The LRUCache ensures thread safety by using a mutex to synchronize access to the underlying `lru.Cache` instance.
// This allows multiple goroutines to safely read from and write to the cache concurrently.
type LRUCache[K comparable, V any] struct {
	cache           *lru.Cache[K, lruCacheItem[V]]
	mu              sync.RWMutex
	cleanupInterval time.Duration
	stopCleanup     chan bool
}

// NewLRUCache creates a new LRU cache with the specified maximum size.
func NewLRUCache[K comparable, V any](maxSize int) Cache[K, V] {
	return NewLRUCacheWithCleanupInterval[K, V](maxSize, 5*time.Minute)
}

// NewLRUCacheWithCleanupInterval creates a new LRU cache with the specified capacity and cleanup interval
func NewLRUCacheWithCleanupInterval[K comparable, V any](maxSize int, cleanupInterval time.Duration) Cache[K, V] {
	if maxSize <= 0 {
		maxSize = 1000 // Default size if an invalid value is provided
	}

	// Create underlying LRU cache from hashicorp
	cache, err := lru.New[K, lruCacheItem[V]](maxSize)
	if err != nil {
		// This should only happen if maxSize is negative, which we already checked
		panic("Failed to create LRU cache: " + err.Error())
	}

	lruCache := &LRUCache[K, V]{
		cache:           cache,
		cleanupInterval: cleanupInterval,
		stopCleanup:     make(chan bool),
	}

	// Start cleanup goroutine
	go lruCache.startCleanup()

	return lruCache
}

// Set adds or updates an item in the cache without expiry
func (c *LRUCache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item := lruCacheItem[V]{
		value:  value,
		expiry: nil,
	}
	c.cache.Add(key, item)
}

// SetWithExpiry adds or updates an item in the cache with an expiry time
func (c *LRUCache[K, V]) SetWithExpiry(key K, value V, expiry time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiryTime := time.Now().Add(expiry)
	item := lruCacheItem[V]{
		value:  value,
		expiry: &expiryTime,
	}
	c.cache.Add(key, item)
}

// Get retrieves an item from the cache
func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.cache.Get(key)
	if !found {
		var zero V
		return zero, false
	}

	// Check if item is expired
	if item.expiry != nil && time.Now().After(*item.expiry) {
		c.mu.RUnlock() // Release read lock before acquiring write lock

		// Need a write lock to remove the item
		c.mu.Lock()
		c.cache.Remove(key)
		c.mu.Unlock()

		// Re-acquire read lock to maintain proper lock ordering
		c.mu.RLock()

		var zero V
		return zero, false
	}

	return item.value, true
}

// Delete removes an item from the cache
func (c *LRUCache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache.Remove(key)
}

// Clear removes all items from the cache
func (c *LRUCache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache.Purge()
}

// Len returns the number of items in the cache
func (c *LRUCache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cache.Len()
}

// startCleanup starts a goroutine that periodically removes expired items
func (c *LRUCache[K, V]) startCleanup() {
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

// cleanupExpired removes all expired items from the cache
func (c *LRUCache[K, V]) cleanupExpired() {
	now := time.Now()

	// Get all keys to check
	c.mu.RLock()
	keys := c.cache.Keys()
	c.mu.RUnlock()

	// Check each key
	for _, key := range keys {
		// Need to reacquire lock each iteration to minimize lock contention
		c.mu.RLock()
		item, exists := c.cache.Get(key)
		expired := exists && item.expiry != nil && now.After(*item.expiry)
		c.mu.RUnlock()

		if expired {
			c.mu.Lock()
			c.cache.Remove(key)
			c.mu.Unlock()
		}
	}
}

// StopCleanup stops the background cleanup goroutine
func (c *LRUCache[K, V]) StopCleanup() {
	c.stopCleanup <- true
}
