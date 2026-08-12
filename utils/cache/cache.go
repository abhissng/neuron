package cache

import (
	"sync"
	"time"
)

// CacheType represents the type of cache implementation to use
type CacheType int

const (
	// Basic is a simple in-memory cache with optional expiry
	Basic CacheType = iota
	// LRU is a Least Recently Used cache implementation
	LRU
)

// Cache defines an interface for generic caching operations.
//
// It provides methods for setting, getting, deleting, clearing, and retrieving the length of the cache.
type Cache[K comparable, V any] interface {
	Set(key K, value V)                                 // Set the value associated with the given key.
	SetWithExpiry(key K, value V, expiry time.Duration) // Set the value with an expiration time.
	Get(key K) (V, bool)                                // Get the value associated with the given key, and a boolean indicating whether the key exists.
	Delete(key K)                                       // Delete the entry associated with the given key.
	Clear()                                             // Clear all entries from the cache.
	Len() int                                           // Return the number of entries currently in the cache.
	StopCleanup()                                       // Stop the background cleanup goroutine.
}

// DefaultCacheName is the conventional name for a default named cache.
const DefaultCacheName = "default"

// CacheConfig holds configuration options for creating caches
type CacheConfig struct {
	// Type determines which cache implementation to use
	Type CacheType
	// CleanupInterval specifies how often to run the cleanup routine
	CleanupInterval time.Duration
	// MaxSize sets the maximum number of items for LRU cache (ignored for Basic cache)
	MaxSize int
}

// defaultConfig returns a default configuration using BasicCache
func defaultConfig() *CacheConfig {
	return &CacheConfig{
		Type:            Basic,
		CleanupInterval: 5 * time.Minute,
	}
}

// DefaultLRUConfig returns a default configuration using LRUCache with 1000 items
func DefaultLRUConfig() *CacheConfig {
	return &CacheConfig{
		Type:            LRU,
		CleanupInterval: 5 * time.Minute,
		MaxSize:         1000,
	}
}

// CacheManager provides a factory for creating and managing different types of caches
type CacheManager struct {
	mu sync.Mutex
	// Store active caches to ensure they can be properly stopped when needed
	caches map[string]Cache[string, any]
	// configuration to use when none is provided
	config *CacheConfig
}

// NewCacheManager creates a new cache manager with the basic cache as default
func NewCacheManager() *CacheManager {
	return NewCacheManagerWithConfig(defaultConfig())
}

// NewCacheManagerWithConfig creates a new cache manager with the specified default configuration
func NewCacheManagerWithConfig(config *CacheConfig) *CacheManager {
	newConfig := defaultConfig()
	if config != nil {
		newConfig = config
	}
	return &CacheManager{
		caches: make(map[string]Cache[string, any]),
		config: newConfig,
	}
}

// CreateCache creates a new cache with the given name using the manager default configuration.
func (m *CacheManager) CreateCache(name string) Cache[string, any] {
	return m.CreateCacheWithConfig(name)
}

// CreateCacheWithConfig creates a new cache with the given name using the manager default configuration.
// Per-call configuration is not supported; name is retained for compatibility.
// The key type is string and value type is any for maximum flexibility.
func (m *CacheManager) CreateCacheWithConfig(name string) Cache[string, any] {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.createCacheLocked(name)
}

// createCacheLocked creates and stores a cache. Caller must hold m.mu.
func (m *CacheManager) createCacheLocked(name string) Cache[string, any] {
	if m.config == nil {
		m.config = defaultConfig()
	}

	var cache Cache[string, any]
	switch m.config.Type {
	case LRU:
		maxSize := m.config.MaxSize
		if maxSize <= 0 {
			maxSize = 1000 // Default size if not specified
			m.config.MaxSize = maxSize
		}
		cache = NewLRUCacheWithCleanupInterval[string, any](maxSize, m.config.CleanupInterval)
	default: // Basic cache is the default
		cache = NewBasicCacheWithCleanupInterval[string, any](m.config.CleanupInterval)
	}

	m.caches[name] = cache
	return cache
}

// GetOrCreateCache returns an existing cache or creates a new one if it doesn't exist.
// Uses the manager default configuration if it needs to create the cache.
func (m *CacheManager) GetOrCreateCache(name string) Cache[string, any] {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cache, exists := m.caches[name]; exists {
		return cache
	}
	return m.createCacheLocked(name)
}

// GetOrCreateCacheWithConfig is a compatibility alias for GetOrCreateCache.
// It uses the manager-level configuration; per-call configuration is not supported.
func (m *CacheManager) GetOrCreateCacheWithConfig(name string) Cache[string, any] {
	return m.GetOrCreateCache(name)
}

// GetCache retrieves a cache by name
func (m *CacheManager) GetCache(name string) (Cache[string, any], bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cache, exists := m.caches[name]
	return cache, exists
}

// RemoveCache stops and removes a cache
func (m *CacheManager) RemoveCache(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cache, exists := m.caches[name]; exists {
		cache.StopCleanup()
		delete(m.caches, name)
	}
}

// StopAll stops all running caches
func (m *CacheManager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, cache := range m.caches {
		cache.StopCleanup()
	}
	m.caches = make(map[string]Cache[string, any])
}

// Getconfig returns the current default configuration
func (m *CacheManager) Getconfig() *CacheConfig {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.config
}

// Setconfig changes the default configuration for new caches.
// A nil config is replaced with defaultConfig() so later CreateCache calls always have a non-nil configuration.
func (m *CacheManager) Setconfig(config *CacheConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if config == nil {
		m.config = defaultConfig()
		return
	}
	m.config = config
}

// CreateTypedCache creates a new cache with specified key and value types
// This is a generic helper for when you need strongly typed caches
func CreateTypedCache[K comparable, V any](config CacheConfig) Cache[K, V] {
	switch config.Type {
	case LRU:
		if config.MaxSize <= 0 {
			config.MaxSize = 1000 // Default size if not specified
		}
		return NewLRUCacheWithCleanupInterval[K, V](config.MaxSize, config.CleanupInterval)
	default: // Basic cache is the default
		return NewBasicCacheWithCleanupInterval[K, V](config.CleanupInterval)
	}
}
