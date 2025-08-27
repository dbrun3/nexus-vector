package benchmark

import (
	"context"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/dbrun3/nexus-vector/api"
	"github.com/dbrun3/nexus-vector/model"
	"github.com/dbrun3/nexus-vector/nexus"
)

const Seed = 2
const SampleSize = 50

func setupNexusBench() (*nexus.Nexus, []api.NexusRequest, error) {

	config := &nexus.Config{
		QdrantHost:     os.Getenv("QDRANT_HOST"),
		RedisHost:      os.Getenv("REDIS_HOST"),
		TorchServeHost: os.Getenv("TORCHSERVE_HOST"),
		ModelName:      os.Getenv("MODEL"),
		Env:            nexus.Test,
	}

	nexus, err := nexus.InitializeNexus(context.Background(), config)
	if err != nil {
		log.Fatalf("Failed to initialize Nexus: %v", err)
	}

	requests := make([]api.NexusRequest, SampleSize)

	for i := range SampleSize {
		userSnap := model.CreateRandomSnapshot(uint64(Seed + i))
		userSimilar := model.CreateSimilarSnapshot(uint64(Seed+1), userSnap)
		userPage := model.CreateRandomPage(uint64(Seed + i*2))
		userPage.Id = strconv.Itoa(i)

		trigger := model.CreateRandomTrigger(uint64(Seed + i))
		triggerSimilar := model.CreateSimilarTrigger(uint64(Seed+1), trigger)
		triggerPage := model.CreateRandomPage(uint64(Seed + (i * 2) + 1))
		triggerPage.Id = strconv.Itoa(i)

		// Craft trigger requests and cache usersnap
		requests[i] = api.NexusRequest{
			UserId:  userSnap.ID,
			Trigger: trigger,
		}
		_, err := nexus.InjestUser(context.Background(), userSnap)
		if err != nil {
			return nil, nil, err
		}

		// Create embeddings similar to the original input
		userEmbedding, err := nexus.DebugUsersnap(context.Background(), userSimilar)
		if err != nil {
			return nil, nil, err
		}
		err = nexus.StorePageInQdrant(context.Background(), userPage, userEmbedding)
		if err != nil {
			return nil, nil, err
		}
		triggerEmbedding, err := nexus.DebugTrigger(context.Background(), triggerSimilar)
		if err != nil {
			return nil, nil, err
		}
		err = nexus.StorePageInQdrant(context.Background(), triggerPage, triggerEmbedding)
		if err != nil {
			return nil, nil, err
		}
	}

	return nexus, requests, nil
}

func BenchmarkGetNexus(b *testing.B) {
	start := time.Now()
	nexus, requests, err := setupNexusBench()
	setupDuration := time.Since(start)

	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}

	b.Logf("Setup completed in %v", setupDuration)
	b.Logf("Setup rate: %.2f items/second", float64(SampleSize)/setupDuration.Seconds())

	var totalRequests int
	var matchedRequests int

	for i := 0; b.Loop(); i++ {
		// Use modulo to cycle through requests if b.N > len(requests)
		n := i % len(requests)
		request := requests[n]

		pages, err := nexus.GetNexus(context.Background(), request)
		if err != nil {
			b.Fatalf("GetNexus failed: %v", err)
		}

		totalRequests++

		if len(pages) == 0 {
			b.Error("GetNexus failed: no pages returned")
			continue
		}

		// Check if any returned page has the expected ID (string representation of n)
		expectedID := strconv.Itoa(n)
		found := false
		for _, page := range pages {
			if page.Id == expectedID {
				found = true
				break
			}
		}

		if found {
			matchedRequests++
		}
	}

	// Calculate and log the match percentage
	matchPercentage := float64(matchedRequests) / float64(totalRequests) * 100
	b.Logf("Page matching results: %d/%d requests found expected page (%.1f%%)",
		matchedRequests, totalRequests, matchPercentage)
}
