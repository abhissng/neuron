package lruCache

import (
	"sync"

	"github.com/hashicorp/golang-lru/v2"
)

// LRUCache provides a thread-safe wrapper around the `lru.Cache` library.
// It implements a simple Least Recently Used (LRU) cache with a fixed capacity.
//
// The LRUCache ensures thread safety by using a mutex to synchronize access to the underlying `lru.Cache` instance.
// This allows multiple goroutines to safely read from and write to the cache concurrently.
type LRUCache[K comparable, V any] struct {
	cache *lru.Cache[K, V] // Underlying lru.Cache instance
	mu    sync.Mutex       // Mutex for thread-safe access
}

// NewLRUCache creates a new LRUCache instance with the specified maximum size.
// It returns an error if the provided maxSize is invalid (e.g., negative).
func NewLRUCache[K comparable, V any](maxSize int) (*LRUCache[K, V], error) {
	cache, err := lru.New[K, V](maxSize)
	if err != nil {
		return nil, err
	}
	return &LRUCache[K, V]{cache: cache}, nil
}

// Add adds a key-value pair to the cache. If the cache is full, the least recently used entry will be evicted.
func (l *LRUCache[K, V]) Add(key K, value V) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cache.Add(key, value)
}

// Get retrieves the value associated with the given key from the cache.
// Returns the value and a boolean indicating whether the key was found.
func (l *LRUCache[K, V]) Get(key K) (V, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	value, ok := l.cache.Get(key)
	return value, ok
}

// Remove removes the entry associated with the given key from the cache.
func (l *LRUCache[K, V]) Remove(key K) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cache.Remove(key)
}

// Purge removes all entries from the cache.
func (l *LRUCache[K, V]) Purge() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cache.Purge()
}

// Len returns the current number of entries in the cache.
func (l *LRUCache[K, V]) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.cache.Len()
}
