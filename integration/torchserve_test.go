package integration

import (
	"context"
	"os"
	"testing"

	"github.com/dbrun3/nexus-vector/torchserve"
)

func Test_TorchServe_Integration(t *testing.T) {
	torchserveHost := os.Getenv("TORCHSERVE_HOST")
	modelName := os.Getenv("MODEL")

	if torchserveHost == "" {
		t.Skip("TORCHSERVE_HOST not set, skipping integration test")
	}
	if modelName == "" {
		modelName = "my_model" // default from docker-compose
	}

	client, err := torchserve.NewClient(torchserveHost, modelName)
	if err != nil {
		t.Fatalf("Failed to create TorchServe client: %v", err)
	}

	ctx := context.Background()

	// Test text to embeddings
	testTexts := []string{"hello world", "integration test"}
	embeddings, err := client.TextToEmbeddings(ctx, testTexts...)
	if err != nil {
		t.Fatalf("Failed to generate embeddings: %v", err)
	}

	if len(embeddings) != len(testTexts) {
		t.Fatalf("Expected %d embeddings, got %d", len(testTexts), len(embeddings))
	}

	for i, embedding := range embeddings {
		if len(embedding) == 0 {
			t.Fatalf("Embedding %d is empty", i)
		}

		// Check that embedding values are reasonable
		var sum float32
		for _, val := range embedding {
			sum += val * val
		}
		magnitude := sum

		if magnitude == 0 {
			t.Fatalf("Embedding %d has zero magnitude", i)
		}

		t.Logf("Text: '%s', Embedding length: %d, Magnitude: %f", testTexts[i], len(embedding), magnitude)
	}
}
