package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/dbrun3/nexus-vector/model"
)

const (
	DatabaseName           = "nexus"
	UserSnapshotCollection = "user_snapshots"
)

type Client struct {
	client *mongo.Client
	db     *mongo.Database
}

// NewClient creates a new MongoDB client and establishes connection
func NewClient(host, user, pass string) (*Client, error) {

	// Early return err on empty
	if host == "" || user == "" || pass == "" {
		return nil, fmt.Errorf("missing host, user or pass")
	}

	// Construct connection URI
	uri := fmt.Sprintf("mongodb://%s:%s@%s:27017", user, pass, host)

	// Set client options
	clientOptions := options.Client().ApplyURI(uri)

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	db := client.Database(DatabaseName)

	mongoClient := &Client{
		client: client,
		db:     db,
	}

	// Ensure UserSnapshot collection exists
	if err := mongoClient.ensureUserSnapshotCollection(ctx); err != nil {
		client.Disconnect(ctx)
		return nil, fmt.Errorf("failed to ensure UserSnapshot collection: %w", err)
	}

	return mongoClient, nil
}

// Close disconnects the MongoDB client
func (c *Client) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

// ensureUserSnapshotCollection creates the UserSnapshot collection if it doesn't exist
func (c *Client) ensureUserSnapshotCollection(ctx context.Context) error {
	// Check if collection already exists
	collections, err := c.db.ListCollectionNames(ctx, bson.M{"name": UserSnapshotCollection})
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	// Collection already exists
	if len(collections) > 0 {
		return nil
	}

	// Create collection with validation schema
	collectionOptions := options.CreateCollection().SetValidator(bson.M{
		"$jsonSchema": bson.M{
			"bsonType": "object",
			"required": []string{"id", "gender", "age", "location"},
			"properties": bson.M{
				"id": bson.M{
					"bsonType":    "string",
					"description": "Unique user snapshot ID",
				},
				"gender": bson.M{
					"bsonType":    "string",
					"description": "User gender",
				},
				"age": bson.M{
					"bsonType":    "int",
					"minimum":     0,
					"description": "User age",
				},
				"location": bson.M{
					"bsonType":    "string",
					"description": "User location",
				},
			},
		},
	})

	err = c.db.CreateCollection(ctx, UserSnapshotCollection, collectionOptions)
	if err != nil {
		return fmt.Errorf("failed to create UserSnapshot collection: %w", err)
	}

	// Create index on ID field for faster queries
	collection := c.db.Collection(UserSnapshotCollection)
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err = collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("failed to create index on UserSnapshot collection: %w", err)
	}

	return nil
}

// StoreUserSnapshot stores a new UserSnapshot document
func (c *Client) StoreUserSnapshot(ctx context.Context, snapshot model.UserSnapshot) error {
	collection := c.db.Collection(UserSnapshotCollection)

	_, err := collection.InsertOne(ctx, snapshot)
	if err != nil {
		return fmt.Errorf("failed to store UserSnapshot: %w", err)
	}

	return nil
}

// GetUserSnapshot retrieves a UserSnapshot by ID
func (c *Client) GetUserSnapshot(ctx context.Context, id string) (*model.UserSnapshot, error) {
	collection := c.db.Collection(UserSnapshotCollection)

	var snapshot model.UserSnapshot
	err := collection.FindOne(ctx, bson.M{"id": id}).Decode(&snapshot)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get UserSnapshot: %w", err)
	}

	return &snapshot, nil
}

// UpdateUserSnapshot updates an existing UserSnapshot
func (c *Client) UpdateUserSnapshot(ctx context.Context, snapshot model.UserSnapshot) error {
	collection := c.db.Collection(UserSnapshotCollection)

	filter := bson.M{"id": snapshot.ID}
	update := bson.M{"$set": snapshot}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update UserSnapshot: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("UserSnapshot with ID %s not found", snapshot.ID)
	}

	return nil
}

// DeleteUserSnapshot deletes a UserSnapshot by ID
func (c *Client) DeleteUserSnapshot(ctx context.Context, id string) error {
	collection := c.db.Collection(UserSnapshotCollection)

	result, err := collection.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return fmt.Errorf("failed to delete UserSnapshot: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("UserSnapshot with ID %s not found", id)
	}

	return nil
}
