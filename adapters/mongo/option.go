package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Option pattern for MongoManager configuration
type MongoOption func(*MongoManager) error

// MongoManager is the core struct that holds the MongoDB client and database connections.
type MongoManager struct {
	client   *mongo.Client
	database *mongo.Database
	timeout  time.Duration
}

// NewMongoManager creates a MongoManager with options
func NewMongoManager(opts ...MongoOption) (*MongoManager, error) {
	m := &MongoManager{
		timeout: 10 * time.Second, // default timeout
	}

	for _, opt := range opts {
		if err := opt(m); err != nil {
			return nil, err
		}
	}

	if m.client == nil || m.database == nil {
		return nil, mongo.ErrClientDisconnected
	}

	return m, nil
}

// WithURI sets the MongoDB URI and database
func WithURI(uri, dbName string) MongoOption {
	return func(m *MongoManager) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		clientOpts := options.Client().ApplyURI(uri)
		client, err := mongo.Connect(clientOpts)
		if err != nil {
			return err
		}

		if err := client.Ping(ctx, nil); err != nil {
			return err
		}

		m.client = client
		m.database = client.Database(dbName)
		return nil
	}
}

// contextWithTimeout returns a context with timeout
func (m *MongoManager) contextWithTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	if m.timeout <= 0 {
		m.timeout = 60 * time.Second
	}
	return context.WithTimeout(parent, m.timeout)
}

// WithTimeout sets custom timeout
func WithTimeout(d time.Duration) MongoOption {
	return func(m *MongoManager) error {
		m.timeout = d
		return nil
	}
}

// GetClient returns the underlying mongo.Client.
func (m *MongoManager) GetClient() *mongo.Client {
	return m.client
}

// GetDB returns the underlying mongo.Database.
func (m *MongoManager) GetDB() *mongo.Database {
	return m.database
}

// Disconnect closes the MongoDB connection
func (m *MongoManager) Disconnect(ctx context.Context) error {
	if m.client != nil {
		if err := m.client.Disconnect(ctx); err != nil {
			return fmt.Errorf("failed to disconnect from mongo: %w", err)
		}
	}
	return nil
}
