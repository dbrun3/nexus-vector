package benchmark

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/dbrun3/nexus-vector/dao"
	"github.com/dbrun3/nexus-vector/model"
	"github.com/dbrun3/nexus-vector/qdrant_util"
	"github.com/dbrun3/nexus-vector/torchserve"
	"github.com/dbrun3/nexus-vector/util"
	"github.com/qdrant/go-client/qdrant"
)

const QdrantCollection = "qdrant_benchmark_collection"
const VectorSize = 384

func setupQdrantBench() (*qdrant.Client, [][]float32, error) {
	qdrantHost := os.Getenv("QDRANT_HOST")
	torchserveHost := os.Getenv("TORCHSERVE_HOST")
	modelName := os.Getenv("MODEL")

	// Initialize clients
	qdrantClient, err := qdrant_util.NewClient(context.Background(), qdrantHost, QdrantCollection, VectorSize)
	if err != nil {
		return nil, nil, err
	}

	torchserveClient, err := torchserve.NewClient(torchserveHost, modelName)
	if err != nil {
		return nil, nil, err
	}

	// Create real embeddings and store test data
	embeddings := make([][]float32, SampleSize)

	for i := range SampleSize {
		// Generate test data for embedding
		userSnap := model.CreateRandomSnapshot(uint64(Seed + i))
		cleanText, err := util.CleanUserSnapshotForEmbedding(userSnap)
		if err != nil {
			return nil, nil, err
		}

		// Generate REAL embedding using TorchServe
		userEmbeddings, err := torchserveClient.TextToEmbeddings(context.Background(), cleanText)
		if err != nil {
			return nil, nil, err
		}
		if len(userEmbeddings) != 1 {
			return nil, nil, err
		}

		embedding := userEmbeddings[0]
		embeddings[i] = embedding

		// Store test page in Qdrant
		page := model.CreateRandomPage(uint64(Seed + i*2))
		payload := dao.NewQdrantPagePayload(page, 0, 0)

		point := &qdrant.PointStruct{
			Id:      qdrant.NewIDNum(uint64(i)),
			Vectors: qdrant.NewVectors(embedding...),
			Payload: qdrant.NewValueMap(payload.ToMap()),
		}

		_, err = qdrantClient.Upsert(context.Background(), &qdrant.UpsertPoints{
			CollectionName: QdrantCollection,
			Points:         []*qdrant.PointStruct{point},
		})
		if err != nil {
			return nil, nil, err
		}
	}

	return qdrantClient, embeddings, nil
}

func BenchmarkQdrantQuery(b *testing.B) {
	start := time.Now()
	client, embeddings, err := setupQdrantBench()
	setupDuration := time.Since(start)

	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}

	b.Logf("Qdrant setup completed in %v", setupDuration)
	b.Logf("Qdrant setup rate: %.2f items/second", float64(SampleSize)/setupDuration.Seconds())

	for i := 0; b.Loop(); i++ {
		// Use modulo to cycle through embeddings
		queryEmbedding := embeddings[i%len(embeddings)]

		// Perform vector similarity search
		result, err := client.Query(context.Background(), &qdrant.QueryPoints{
			CollectionName: QdrantCollection,
			Query:          qdrant.NewQuery(queryEmbedding...),
			Limit:          &[]uint64{10}[0],
			WithPayload:    qdrant.NewWithPayload(true),
		})
		if err != nil {
			b.Fatalf("Qdrant query failed: %v", err)
		}

		// Prevent compiler optimization
		_ = result
	}
}

func BenchmarkQdrantUpsert(b *testing.B) {
	qdrantHost := os.Getenv("QDRANT_HOST")
	torchserveHost := os.Getenv("TORCHSERVE_HOST")
	modelName := os.Getenv("MODEL")

	// Initialize clients
	qdrantClient, err := qdrant_util.NewClient(context.Background(), qdrantHost, QdrantCollection+"_upsert", VectorSize)
	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}

	torchserveClient, err := torchserve.NewClient(torchserveHost, modelName)
	if err != nil {
		b.Fatalf("TorchServe setup failed: %v", err)
	}

	// Pre-generate some real embeddings to cycle through during benchmark
	precomputedEmbeddings := make([][]float32, 10) // Small cache of real embeddings
	for i := 0; i < 10; i++ {
		userSnap := model.CreateRandomSnapshot(uint64(Seed + i))
		cleanText, _ := util.CleanUserSnapshotForEmbedding(userSnap)

		embeddings, err := torchserveClient.TextToEmbeddings(context.Background(), cleanText)
		if err != nil {
			b.Fatalf("Failed to generate embedding: %v", err)
		}
		precomputedEmbeddings[i] = embeddings[0]
	}

	// Reset timer to exclude setup time
	b.ResetTimer()

	for i := 0; b.Loop(); i++ {
		// Generate test data
		page := model.CreateRandomPage(uint64(Seed + i))
		payload := dao.NewQdrantPagePayload(page, 0, 0)

		// Use a real embedding from our precomputed cache
		embedding := precomputedEmbeddings[i%len(precomputedEmbeddings)]

		point := &qdrant.PointStruct{
			Id:      qdrant.NewIDNum(uint64(i)),
			Vectors: qdrant.NewVectors(embedding...),
			Payload: qdrant.NewValueMap(payload.ToMap()),
		}

		// Perform upsert operation
		_, err = qdrantClient.Upsert(context.Background(), &qdrant.UpsertPoints{
			CollectionName: QdrantCollection + "_upsert",
			Points:         []*qdrant.PointStruct{point},
		})
		if err != nil {
			b.Fatalf("Qdrant upsert failed: %v", err)
		}
	}
}
