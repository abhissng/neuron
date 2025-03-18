package idempotency

import (
	"sync"
	"time"
)

const (
	DefaultCleanupInterval = 10 * time.Minute
)

// IdempotencyManager provides a mechanism to track and prevent duplicate processing of events.
// It uses a map to store the last processed time for each event with a given tracking ID.
// A background goroutine periodically cleans up the map to remove entries that have not been
// accessed within the specified cleanup interval.
type IdempotencyManager[K comparable] struct {
	trackedEvents   map[K]time.Time // Map to store the last processed time for each event
	mu              sync.Mutex      // Mutex for thread-safe access to the trackedEvents map
	cleanupInterval time.Duration   // Interval for cleaning up expired entries
	cleanupTicker   *time.Ticker    // Ticker to trigger periodic cleanup
	done            chan struct{}   // Channel to signal the manager to stop the cleanup routine
}

// NewIdempotencyManager creates a new instance of IdempotencyManager with the specified cleanup interval.
// It starts a background goroutine to perform periodic cleanup.
func NewIdempotencyManager[K comparable](cleanupInterval time.Duration) *IdempotencyManager[K] {
	manager := &IdempotencyManager[K]{
		trackedEvents:   make(map[K]time.Time),
		cleanupInterval: cleanupInterval,
		done:            make(chan struct{}),
	}
	go manager.startCleanup()
	return manager
}

// startCleanup starts the background goroutine for periodic cleanup.
// It runs until the 'done' channel receives a signal.
func (m *IdempotencyManager[K]) startCleanup() {
	m.cleanupTicker = time.NewTicker(m.cleanupInterval)
	defer m.cleanupTicker.Stop()

	for {
		select {
		case <-m.done:
			return
		case <-m.cleanupTicker.C:
			m.cleanupProcessedMessages()
		}
	}
}

// cleanupProcessedMessages removes entries from the trackedEvents map
// that have not been accessed within the cleanup interval.
func (m *IdempotencyManager[K]) cleanupProcessedMessages() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for trackingID, timestamp := range m.trackedEvents {
		if now.Sub(timestamp) > m.cleanupInterval {
			delete(m.trackedEvents, trackingID)
		}
	}
}

// MarkAsProcessed marks an event with the given trackingID as processed.
// It updates the last processed time for the event in the map.
func (m *IdempotencyManager[K]) MarkAsProcessed(trackingID K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.trackedEvents[trackingID] = time.Now()
}

// IsProcessed checks if an event with the given trackingID has already been processed.
func (m *IdempotencyManager[K]) IsProcessed(trackingID K) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, exists := m.trackedEvents[trackingID]
	return exists
}

// Close signals the cleanup goroutine to stop and releases any acquired resources.
func (m *IdempotencyManager[K]) Close() {
	close(m.done)
}
