package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readconcern"
)

// Collection is a generic wrapper around mongo.Collection that provides type-safe CRUD operations.
type Collection[T any] struct {
	collection *mongo.Collection
	timeout    time.Duration
}

// GetCollection returns a new generic Collection instance for a specific document type.
func GetCollection[T any](m *MongoManager, collectionName string) *Collection[T] {
	return &Collection[T]{
		collection: m.database.Collection(collectionName),
		timeout:    m.timeout,
	}
}

// --- Generic CRUD Operations ---

// contextWithTimeout returns a context with timeout
func (c *Collection[T]) contextWithTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	if c.timeout <= 0 {
		c.timeout = 60 * time.Second
	}
	return context.WithTimeout(parent, c.timeout)
}

// InsertOne inserts a single document into the collection.
func (c *Collection[T]) InsertOne(ctx context.Context, document T) (*mongo.InsertOneResult, error) {
	ctx, cancel := c.contextWithTimeout(ctx)
	defer cancel()
	return c.collection.InsertOne(ctx, document)
}

// InsertMany inserts multiple documents into the collection.
func (c *Collection[T]) InsertMany(ctx context.Context, documents []T) (*mongo.InsertManyResult, error) {
	ctx, cancel := c.contextWithTimeout(ctx)
	defer cancel()
	docs := make([]interface{}, len(documents))
	for i, d := range documents {
		docs[i] = d
	}
	return c.collection.InsertMany(ctx, docs)
}

// FindOne finds a single document matching the filter.
func (c *Collection[T]) FindOne(ctx context.Context, filter bson.M, opts ...options.Lister[options.FindOneOptions]) (*T, error) {
	var result T
	err := c.collection.FindOne(ctx, filter, opts...).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Return nil, nil if no document is found, a common Go pattern.
		}
		return nil, err
	}
	return &result, nil
}

// Find finds multiple documents matching the filter.
func (c *Collection[T]) Find(ctx context.Context, filter bson.M, opts ...options.Lister[options.FindOptions]) ([]*T, error) {
	cursor, err := c.collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*T
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode documents: %w", err)
	}

	return results, nil
}

// UpdateOne updates a single document matching the filter.
func (c *Collection[T]) UpdateOne(ctx context.Context, filter, update bson.M, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error) {
	ctx, cancel := c.contextWithTimeout(ctx)
	defer cancel()
	return c.collection.UpdateOne(ctx, filter, update, opts...)
}

// UpdateMany updates multiple documents matching the filter.
func (c *Collection[T]) UpdateMany(ctx context.Context, filter, update bson.M, opts ...options.Lister[options.UpdateManyOptions]) (*mongo.UpdateResult, error) {
	ctx, cancel := c.contextWithTimeout(ctx)
	defer cancel()
	return c.collection.UpdateMany(ctx, filter, update, opts...)
}

// DeleteOne deletes a single document matching the filter.
func (c *Collection[T]) DeleteOne(ctx context.Context, filter bson.M, opts ...options.Lister[options.DeleteOneOptions]) (*mongo.DeleteResult, error) {
	ctx, cancel := c.contextWithTimeout(ctx)
	defer cancel()
	return c.collection.DeleteOne(ctx, filter, opts...)
}

// DeleteMany deletes multiple documents matching the filter.
func (c *Collection[T]) DeleteMany(ctx context.Context, filter bson.M, opts ...options.Lister[options.DeleteManyOptions]) (*mongo.DeleteResult, error) {
	ctx, cancel := c.contextWithTimeout(ctx)
	defer cancel()
	return c.collection.DeleteMany(ctx, filter, opts...)
}

// Aggregate executes an aggregation pipeline.
// Note: Decoding the results is the responsibility of the caller, as aggregation
// results can have a different structure than the collection's document type T.
func (c *Collection[T]) Aggregate(ctx context.Context, pipeline mongo.Pipeline, opts ...options.Lister[options.AggregateOptions]) (*mongo.Cursor, context.CancelFunc, error) {
	ctx, cancel := c.contextWithTimeout(ctx)
	cursor, err := c.collection.Aggregate(ctx, pipeline, opts...)
	if err != nil {
		cancel()
		return nil, nil, err
	}
	return cursor, cancel, nil
}

// --- Transaction Management ---

// WithTransaction runs the given function within a MongoDB transaction.
// It handles starting the session, committing, and aborting the transaction automatically.
// The function `fn` receives a `mongo.SessionContext` which MUST be used for all
// operations within the transaction.
func (m *MongoManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
	ctx, cancel := m.contextWithTimeout(ctx)
	defer cancel()
	txnOpts := options.Transaction().SetReadConcern(readconcern.Majority())
	sessOpts := options.Session().SetDefaultTransactionOptions(txnOpts)

	session, err := m.client.StartSession(sessOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	// mongo.WithSession starts a transaction, runs the callback, and handles commit/abort.
	result, err := session.WithTransaction(ctx, func(ctx context.Context) (any, error) {
		return fn(ctx)
	})

	if err != nil {
		return nil, fmt.Errorf("transaction failed: %w", err)
	}

	return result, nil
}
