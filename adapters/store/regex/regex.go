package regex

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
)

// RegexEntry defines the structure for a single regex pattern.
type RegexEntry struct {
	Pattern     string `json:"pattern"`
	Description string `json:"description"`
}

// RegexStore holds the definitions loaded from JSON.
type RegexStore map[string]RegexEntry

// RegexManager provides a convenient and high-performance way
// to load and use compiled regular expressions.
type RegexManager struct {
	store    RegexStore                // Stores the raw string definitions
	compiled map[string]*regexp.Regexp // Caches compiled regex objects
	mu       sync.RWMutex              // Protects the 'compiled' map
}

// NewRegexManager loads definitions from a JSON file and initializes the manager.
func NewRegexManager(input string) (*RegexManager, error) {
	// Read the JSON file
	var file string
	switch helpers.DetectInputType(input) {
	case "file":
		fileBytes, err := os.ReadFile(filepath.Clean(input))
		if err != nil {
			return nil, fmt.Errorf("failed to read regex file: %w", err)
		}
		file = string(fileBytes)
	case "json":
		file = input
	default:
		return nil, errors.New("invalid input type")

	}

	// Unmarshal the data into our store
	var store RegexStore
	if err := json.Unmarshal([]byte(file), &store); err != nil {
		return nil, fmt.Errorf("failed to parse regex JSON: %w", err)
	}

	// Initialize the manager
	manager := &RegexManager{
		store:    store,
		compiled: make(map[string]*regexp.Regexp),
		mu:       sync.RWMutex{},
	}

	return manager, nil
}

// Get compiled regex by its key (e.g., "email").
// It compiles the regex on the first call and caches it for future use.
func (m *RegexManager) Get(key string) (*regexp.Regexp, error) {
	// Check if it's already compiled (read lock)
	m.mu.RLock()
	re, exists := m.compiled[key]
	m.mu.RUnlock()

	if exists {
		return re, nil
	}

	// Not compiled, so we need a full lock to write
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check in case another goroutine compiled it while we waited for the lock
	re, exists = m.compiled[key]
	if exists {
		return re, nil
	}

	// Find the definition
	entry, exists := m.store[key]
	if !exists {
		return nil, fmt.Errorf("regex definition not found for key: %s", key)
	}

	// Compile and store it
	compiledRe, err := regexp.Compile(entry.Pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to compile regex for key '%s': %w", key, err)
	}

	m.compiled[key] = compiledRe
	return compiledRe, nil
}

// MustGet is a helper that returns nil if the regex key doesn't exist or is invalid.
// Useful for setup code where you know the key must be valid.
func (m *RegexManager) MustGet(key string) *regexp.Regexp {
	re, err := m.Get(key)
	if err != nil {
		helpers.Println(constant.ERROR, "unable to get regex", err.Error())
		return nil
	}
	return re
}

// IsMatch checks if a string matches a pre-defined regex.
func (m *RegexManager) IsMatch(key, s string) bool {
	re, err := m.Get(key)
	if err != nil {
		return false
	}
	return re.MatchString(s)
}
