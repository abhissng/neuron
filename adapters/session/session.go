package session

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/abhissng/neuron/adapters/redis"
	"github.com/abhissng/neuron/blame"
	"github.com/abhissng/neuron/result"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SessionData represents the data stored in a session
type SessionData struct {
	OrgID           types.OrgID    `json:"org_id,omitempty"`
	UserID          types.UserID   `json:"user_id,omitempty"`
	Username        string         `json:"username,omitempty"`
	Email           string         `json:"email,omitempty"`
	Role            string         `json:"role,omitempty"`
	LastAccess      time.Time      `json:"last_access,omitempty"`
	CustomData      map[string]any `json:"custom_data,omitempty"`
	IsAuthenticated bool           `json:"is_authenticated"`
}

func NewSessionData() *SessionData {
	return &SessionData{
		CustomData: make(map[string]any),
	}
}

// SessionManager handles session operations
type SessionManager struct {
	store                   *redis.RedisManager
	sessionPrefix           string
	defaultExpiry           time.Duration
	sessionMiddlewareOption *SessionMiddlewareOptions
}

// Option is a function that configures the SessionManager
type Option func(*SessionManager)

// WithRedisManager sets the Redis manager
func WithRedisManager(manager *redis.RedisManager) Option {
	return func(sm *SessionManager) {
		sm.store = manager
	}
}

// WithSessionPrefix sets the session prefix
func WithSessionPrefix(prefix string) Option {
	return func(sm *SessionManager) {
		sm.sessionPrefix = prefix
	}
}

// WithDefaultExpiry sets the default expiry duration
func WithDefaultExpiry(expiry time.Duration) Option {
	return func(sm *SessionManager) {
		sm.defaultExpiry = expiry
	}
}

// WithSessionMiddlewareOption sets the session middleware options
func WithSessionMiddlewareOption(option *SessionMiddlewareOptions) Option {
	return func(sm *SessionManager) {
		sm.sessionMiddlewareOption = option
	}
}

// NewSessionManager creates a new session manager with the provided options
func NewSessionManager(opts ...Option) (*SessionManager, error) {
	sm := &SessionManager{
		store:                   &redis.RedisManager{},
		sessionPrefix:           "session:",     // Default prefix
		defaultExpiry:           24 * time.Hour, // Default expiry
		sessionMiddlewareOption: NewSessionMiddlewareOptions(),
	}

	// Apply all the options
	for _, opt := range opts {
		opt(sm)
	}

	return sm, nil
}

// CreateSession creates a new session
func (sm *SessionManager) CreateSession(ctx context.Context, userData SessionData) (string, error) {
	sessionID := uuid.New().String()
	userData.LastAccess = time.Now()

	data, err := json.Marshal(userData)
	if err != nil {
		return "", err
	}

	key := sm.sessionPrefix + sessionID
	err = sm.store.Set(ctx, key, data, sm.defaultExpiry)
	if err != nil {
		return "", err
	}

	return sessionID, nil
}

// GetSession retrieves session data
func (sm *SessionManager) GetSession(ctx context.Context, sessionID string) (*SessionData, error) {
	key := sm.sessionPrefix + sessionID
	data, err := sm.store.Get(ctx, key)
	if err != nil {
		return nil, errors.New("session not found")
	}

	var sessionData SessionData
	err = json.Unmarshal([]byte(data), &sessionData)
	if err != nil {
		return nil, err
	}

	// Update last access time
	sessionData.LastAccess = time.Now()
	_ = sm.RefreshSession(ctx, sessionID, sessionData)

	return &sessionData, nil
}

// RefreshSession updates the session data and extends its expiry
func (sm *SessionManager) RefreshSession(ctx context.Context, sessionID string, data SessionData) error {
	key := sm.sessionPrefix + sessionID
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return sm.store.Set(ctx, key, jsonData, sm.defaultExpiry)
}

// DestroySession removes a session
func (sm *SessionManager) DestroySession(ctx context.Context, sessionID string) error {
	key := sm.sessionPrefix + sessionID
	_, err := sm.store.Delete(ctx, key)
	return err
}

// UpdateSessionData updates specific fields in the session
func (sm *SessionManager) UpdateSessionData(ctx context.Context, sessionID string, updates map[string]interface{}) error {
	session, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	// Update the custom data
	if session.CustomData == nil {
		session.CustomData = make(map[string]interface{})
	}

	for k, v := range updates {
		session.CustomData[k] = v
	}

	return sm.RefreshSession(ctx, sessionID, *session)
}

// IsAuthenticated checks if the session exists and is authenticated
func (sm *SessionManager) IsAuthenticated(ctx context.Context, sessionID string) bool {
	session, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		return false
	}
	return session.IsAuthenticated
}

// SessionMiddlewareOption returns the session middleware options.
func (sm *SessionManager) SessionMiddlewareOption() *SessionMiddlewareOptions {
	return sm.sessionMiddlewareOption
}

// SessionValidator defines a function that validates session data.
type SessionValidator func(session *SessionData, extra map[string]any) error

// ValidateSession validates a session and runs multiple custom validators.
func (s *SessionManager) ValidateSession(
	c *gin.Context,
	sessionID string,
	extra map[string]any,
	validators ...SessionValidator,
) result.Result[SessionData] {

	// 🔹 Fetch session data from your store (Redis, DB, memory, etc.)
	data, err := s.GetSession(c, sessionID)
	if err != nil {
		return result.NewFailure[SessionData](blame.SessionNotFound())
	}

	if !data.IsAuthenticated {
		return result.NewFailure[SessionData](blame.SessionUnauthenticated())
	}

	// 🔹 Validate essential fields
	if helpers.IsEmpty(data.UserID) {
		return result.NewFailure[SessionData](blame.SessionMalformed(errors.New("missing user related information")))
	}

	// 🔹 Run custom validators (if any)
	for _, validator := range validators {
		if validator == nil {
			continue
		}
		if err := validator(data, extra); err != nil {
			return result.NewFailure[SessionData](blame.SessionValidationFailed(err))
		}
	}

	// ✅ Valid session
	return result.NewSuccess(data)
}
