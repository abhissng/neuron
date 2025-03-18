package session

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// SessionData represents the data stored in a session
type SessionData struct {
	UserID          string                 `json:"user_id"`
	Username        string                 `json:"username"`
	Email           string                 `json:"email"`
	Role            string                 `json:"role"`
	LastAccess      time.Time              `json:"last_access"`
	CustomData      map[string]interface{} `json:"custom_data"`
	IsAuthenticated bool                   `json:"is_authenticated"`
}

// SessionManager handles session operations
type SessionManager struct {
	store         *redis.Client
	sessionPrefix string
	defaultExpiry time.Duration
}

// SessionConfig contains configuration for the session manager
type SessionConfig struct {
	RedisAddr     string
	RedisPassword string
	SessionPrefix string
	DefaultExpiry time.Duration
}

// NewSessionManager creates a new session manager
func NewSessionManager(config SessionConfig) (*SessionManager, error) {
	if config.DefaultExpiry == 0 {
		config.DefaultExpiry = 24 * time.Hour // Default to 24 hours
	}

	if config.SessionPrefix == "" {
		config.SessionPrefix = "session:"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       0,
	})

	return &SessionManager{
		store:         rdb,
		sessionPrefix: config.SessionPrefix,
		defaultExpiry: config.DefaultExpiry,
	}, nil
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
	err = sm.store.Set(ctx, key, data, sm.defaultExpiry).Err()
	if err != nil {
		return "", err
	}

	return sessionID, nil
}

// GetSession retrieves session data
func (sm *SessionManager) GetSession(ctx context.Context, sessionID string) (*SessionData, error) {
	key := sm.sessionPrefix + sessionID
	data, err := sm.store.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, errors.New("session not found")
		}
		return nil, err
	}

	var sessionData SessionData
	err = json.Unmarshal(data, &sessionData)
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

	return sm.store.Set(ctx, key, jsonData, sm.defaultExpiry).Err()
}

// DestroySession removes a session
func (sm *SessionManager) DestroySession(ctx context.Context, sessionID string) error {
	key := sm.sessionPrefix + sessionID
	return sm.store.Del(ctx, key).Err()
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
