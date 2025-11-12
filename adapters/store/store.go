package store

import "github.com/abhissng/neuron/adapters/store/regex"

type StoreManager struct {
	*regex.RegexManager
}

// StoreOption is a function that configures a StoreManager
type StoreOption func(*StoreManager)

// NewStoreManager creates a new StoreManager with the given options
func NewStoreManager(options ...StoreOption) *StoreManager {
	store := &StoreManager{}
	for _, opt := range options {
		opt(store)
	}
	return store
}

// WithRegexManager sets the regex manager for the store
func WithRegexManager(manager *regex.RegexManager) StoreOption {
	return func(store *StoreManager) {
		store.RegexManager = manager
	}
}

func (s *StoreManager) GetRegexManager() *regex.RegexManager {
	return s.RegexManager
}
