package integration

import (
	"context"
	"os"
	"testing"

	"github.com/dbrun3/nexus-vector/model"
	"github.com/dbrun3/nexus-vector/mongo"
	"github.com/google/uuid"
)

func Test_MongoDB_Integration(t *testing.T) {
	mongoHost := os.Getenv("MONGODB_HOST")
	mongoUser := os.Getenv("MONGODB_USER")
	mongoPass := os.Getenv("MONGODB_PASS")

	if mongoHost == "" {
		t.Skip("MONGODB_HOST not set, skipping integration test")
	}
	if mongoUser == "" {
		mongoUser = "nexus" // default from docker-compose
	}
	if mongoPass == "" {
		mongoPass = "password" // default from docker-compose
	}

	client, err := mongo.NewClient(mongoHost, mongoUser, mongoPass)
	if err != nil {
		t.Fatalf("Failed to create MongoDB client: %v", err)
	}

	ctx := context.Background()
	defer client.Close(ctx)

	// Create test user snapshot with deterministic seed for testing
	testSnapshot := model.CreateRandomSnapshot(12345)
	testSnapshot.ID = uuid.New().String() // Ensure unique ID for test

	// Test store
	err = client.StoreUserSnapshot(ctx, testSnapshot)
	if err != nil {
		t.Fatalf("Failed to store user snapshot: %v", err)
	}

	// Test get
	retrievedSnapshot, err := client.GetUserSnapshot(ctx, testSnapshot.ID)
	if err != nil {
		t.Fatalf("Failed to get user snapshot: %v", err)
	}

	if retrievedSnapshot == nil {
		t.Fatal("Retrieved snapshot is nil")
	}

	if retrievedSnapshot.ID != testSnapshot.ID {
		t.Fatalf("Expected ID '%s', got '%s'", testSnapshot.ID, retrievedSnapshot.ID)
	}

	if retrievedSnapshot.Gender != testSnapshot.Gender {
		t.Fatalf("Expected Gender '%s', got '%s'", testSnapshot.Gender, retrievedSnapshot.Gender)
	}

	if retrievedSnapshot.Age != testSnapshot.Age {
		t.Fatalf("Expected Age %d, got %d", testSnapshot.Age, retrievedSnapshot.Age)
	}

	// Clean up - delete the test snapshot
	err = client.DeleteUserSnapshot(ctx, testSnapshot.ID)
	if err != nil {
		t.Fatalf("Failed to delete user snapshot: %v", err)
	}

	// Verify deletion
	deletedSnapshot, err := client.GetUserSnapshot(ctx, testSnapshot.ID)
	if err != nil {
		t.Fatalf("Error checking deleted snapshot: %v", err)
	}
	if deletedSnapshot != nil {
		t.Fatal("Snapshot was not deleted properly")
	}

	t.Logf("MongoDB integration test successful: stored/retrieved/deleted user %s", testSnapshot.ID)
}