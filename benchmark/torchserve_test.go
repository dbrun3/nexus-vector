package benchmark

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/dbrun3/nexus-vector/model"
	"github.com/dbrun3/nexus-vector/torchserve"
	"github.com/dbrun3/nexus-vector/util"
)

const BatchSize = 5

func setupTorchServeBench() (*torchserve.Client, []string, error) {
	torchserveHost := os.Getenv("TORCHSERVE_HOST")
	modelName := os.Getenv("MODEL")

	// Initialize TorchServe client
	client, err := torchserve.NewClient(torchserveHost, modelName)
	if err != nil {
		return nil, nil, err
	}

	// Create sample text inputs for embedding generation
	texts := make([]string, SampleSize)

	for i := range SampleSize {
		trigger := model.CreateRandomTrigger(uint64(Seed + i))
		cleanText, _ := util.CleanTriggerForEmbedding(trigger)
		texts[i] = cleanText
	}

	return client, texts, nil
}

func BenchmarkTorchServeEmbedding(b *testing.B) {
	start := time.Now()
	client, texts, err := setupTorchServeBench()
	setupDuration := time.Since(start)

	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}

	b.Logf("TorchServe setup completed in %v", setupDuration)
	b.Logf("TorchServe setup rate: %.2f items/second", float64(SampleSize)/setupDuration.Seconds())

	for i := 0; b.Loop(); i++ {
		// Use modulo to cycle through texts
		text := texts[i%len(texts)]

		// Generate embedding
		embeddings, err := client.TextToEmbeddings(context.Background(), text)
		if err != nil {
			b.Fatalf("TorchServe embedding failed: %v", err)
		}

		if len(embeddings) != 1 {
			b.Fatalf("Expected 1 embedding, got %d", len(embeddings))
		}

		// Prevent compiler optimization
		_ = embeddings[0]
	}
}

func BenchmarkTorchServeBatch(b *testing.B) {
	client, texts, err := setupTorchServeBench()
	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}

	batches := make([][]string, 0)

	for i := 0; i < len(texts); i += BatchSize {
		end := min(i+BatchSize, len(texts))
		batches = append(batches, texts[i:end])
	}

	for i := 0; b.Loop(); i++ {
		// Use modulo to cycle through batches
		batch := batches[i%len(batches)]

		// Generate embeddings for batch
		embeddings, err := client.TextToEmbeddings(context.Background(), batch...)
		if err != nil {
			b.Fatalf("TorchServe batch embedding failed: %v", err)
		}

		if len(embeddings) != len(batch) {
			b.Fatalf("Expected %d embeddings, got %d", len(batch), len(embeddings))
		}

		// Prevent compiler optimization
		_ = embeddings
	}
}
