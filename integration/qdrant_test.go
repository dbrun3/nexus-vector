package integration

import (
	"context"
	"os"
	"testing"

	"github.com/dbrun3/nexus-vector/qdrant_util"
	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"
)

func Test_Qdrant_Integration(t *testing.T) {
	qdrantHost := os.Getenv("QDRANT_HOST")
	if qdrantHost == "" {
		t.Skip("QDRANT_HOST not set, skipping integration test")
	}

	ctx := context.Background()

	testCollection := "test_integration_collection"
	client, err := qdrant_util.NewClient(ctx, qdrantHost, testCollection, 3)
	if err != nil {
		t.Fatalf("Failed to create Qdrant client: %v", err)
	}

	// Clean up collection after test
	defer func() {
		client.DeleteCollection(ctx, testCollection)
	}()

	// Test upsert - first point
	pointID1 := uuid.New().String()
	testVector := []float32{1.0, 2.0, 3.0}
	testPayload := map[string]any{
		"test_field": "test_value",
		"number":     42,
	}

	point1 := &qdrant.PointStruct{
		Id:      qdrant.NewID(pointID1),
		Vectors: qdrant.NewVectors(testVector...),
		Payload: qdrant.NewValueMap(testPayload),
	}

	_, err = client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: testCollection,
		Points:         []*qdrant.PointStruct{point1},
	})
	if err != nil {
		t.Fatalf("Failed to upsert first point: %v", err)
	}

	// Test upsert - second point with different vector and payload
	pointID2 := uuid.New().String()
	differentVector := []float32{4.0, 5.0, 6.0}
	differentPayload := map[string]any{
		"test_field": "different_value",
		"number":     99,
	}

	point2 := &qdrant.PointStruct{
		Id:      qdrant.NewID(pointID2),
		Vectors: qdrant.NewVectors(differentVector...),
		Payload: qdrant.NewValueMap(differentPayload),
	}

	_, err = client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: testCollection,
		Points:         []*qdrant.PointStruct{point2},
	})
	if err != nil {
		t.Fatalf("Failed to upsert second point: %v", err)
	}

	// Test search
	var one uint64 = 1
	searchResult, err := client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: testCollection,
		Query:          qdrant.NewQuery(testVector...),
		Limit:          &one,
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		t.Fatalf("Failed to search points: %v", err)
	}

	if len(searchResult) == 0 {
		t.Fatal("Search returned no results")
	}

	if searchResult[0].Score < 0.99 {
		t.Fatalf("Expected high similarity score, got: %f", searchResult[0].Score)
	}

	// Verify the returned payload matches the first point's testPayload, not the second
	payload := searchResult[0].Payload
	if payload == nil {
		t.Fatal("Search result payload is nil")
	}

	// Check test_field value
	testFieldValue, ok := payload["test_field"]
	if !ok {
		t.Fatal("test_field not found in payload")
	}
	if testFieldValue.GetStringValue() != "test_value" {
		t.Fatalf("Expected test_field='test_value', got: %v", testFieldValue)
	}

	// Check number value
	numberValue, ok := payload["number"]
	if !ok {
		t.Fatal("number not found in payload")
	}
	if numberValue.GetIntegerValue() != 42 {
		t.Fatalf("Expected number=42, got: %v", numberValue.GetIntegerValue())
	}

	t.Logf("Qdrant search successful, score: %f, verified correct payload returned", searchResult[0].Score)
}
