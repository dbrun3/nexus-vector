package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/dbrun3/nexus-vector/api"
	"github.com/dbrun3/nexus-vector/model"
	"github.com/dbrun3/nexus-vector/nexus"
	"github.com/qdrant/go-client/qdrant"
)

func Test_Nexus_Integration(t *testing.T) {
	// Check required environment variables
	requiredEnvs := []string{"QDRANT_HOST", "REDIS_HOST", "TORCHSERVE_HOST", "MODEL"}
	for _, env := range requiredEnvs {
		if os.Getenv(env) == "" {
			t.Skipf("%s not set, skipping integration test", env)
		}
	}

	ctx := context.Background()

	// Initialize Nexus with test configuration
	config := &nexus.Config{
		QdrantHost:     os.Getenv("QDRANT_HOST"),
		RedisHost:      os.Getenv("REDIS_HOST"),
		TorchServeHost: os.Getenv("TORCHSERVE_HOST"),
		ModelName:      os.Getenv("MODEL"),
		Env:            nexus.Test,
	}

	n, err := nexus.InitializeNexus(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize Nexus: %v", err)
	}

	// Test 1: Inject a user and verify embedding is cached
	testUser := model.CreateRandomSnapshot(12345)
	testUser.ID = "test-user-integration"

	embedding, err := n.InjestUser(ctx, testUser)
	if err != nil {
		t.Fatalf("Failed to inject user: %v", err)
	}

	if len(embedding) == 0 {
		t.Fatal("Expected non-empty embedding from InjestUser")
	}

	t.Logf("Successfully injected user with %d-dimensional embedding", len(embedding))

	// Test 2: Store a test page in Qdrant using the same embedding
	testPage := model.CreateRandomPage(67890)
	testPage.Id = "test-page-integration"

	err = n.StorePageInQdrant(ctx, testPage, embedding)
	if err != nil {
		t.Fatalf("Failed to store page in Qdrant: %v", err)
	}

	t.Logf("Successfully stored test page: %s", testPage.Id)

	// Test 3: Test GetNexus with a trigger request
	testTrigger := model.CreateRandomTrigger(54321)

	request := api.NexusRequest{
		UserId:  testUser.ID,
		Trigger: testTrigger,
	}

	pages, err := n.GetNexus(ctx, request)
	if err != nil {
		t.Fatalf("GetNexus failed: %v", err)
	}

	t.Logf("GetNexus returned %d pages", len(pages))

	// Test 4: Verify we can get some results (may be 0 if similarity is too low)
	if len(pages) == 0 {
		t.Log("No pages returned - this may indicate similarity threshold too high or filtering issues")
	} else {
		t.Logf("Found pages: %v", pages[0].Id)

		// Check if our test page was returned
		found := false
		for _, page := range pages {
			if page.Id == testPage.Id {
				found = true
				break
			}
		}

		if found {
			t.Log("Successfully retrieved the stored test page")
		} else {
			t.Log("Test page not returned - may be due to similarity threshold or different embeddings")
		}
	}

	// Test 5: Direct debug queries to verify embeddings work
	debugUserEmbedding, err := n.DebugUsersnap(ctx, testUser)
	if err != nil {
		t.Fatalf("DebugUsersnap failed: %v", err)
	}

	if len(debugUserEmbedding) == 0 {
		t.Fatal("Expected non-empty embedding from DebugUsersnap")
	}

	debugTriggerEmbedding, err := n.DebugTrigger(ctx, testTrigger)
	if err != nil {
		t.Fatalf("DebugTrigger failed: %v", err)
	}

	if len(debugTriggerEmbedding) == 0 {
		t.Fatal("Expected non-empty embedding from DebugTrigger")
	}

	t.Logf("Debug embeddings - User: %d dims, Trigger: %d dims",
		len(debugUserEmbedding), len(debugTriggerEmbedding))

	// Test 6: Direct Qdrant query without any filters to see if page exists
	directResult, err := n.DebugQd().Query(ctx, &qdrant.QueryPoints{
		CollectionName: "page_collection",
		Query:          qdrant.NewQuery(embedding...),
		Limit:          qdrant.PtrOf(uint64(10)),
		WithPayload:    qdrant.NewWithPayload(true),
		// NO FILTERS - just see if the page exists at all
	})
	if err != nil {
		t.Fatalf("Direct query failed: %v", err)
	}

	t.Logf("Direct query (no filters) returned %d results", len(directResult))
	if len(directResult) > 0 {
		t.Logf("First result score: %.4f", directResult[0].Score)
	}

	// Test 7: Query with the same time filters that GetNexus uses
	now := float64(time.Now().Unix())
	filteredResult, err := n.DebugQd().Query(ctx, &qdrant.QueryPoints{
		CollectionName: "page_collection",
		Query:          qdrant.NewQuery(embedding...),
		WithPayload:    qdrant.NewWithPayload(true),
		Filter: &qdrant.Filter{
			Must: []*qdrant.Condition{
				qdrant.NewRange("from", &qdrant.Range{
					Lte: &now,
				}),
				qdrant.NewRange("until", &qdrant.Range{
					Gte: &now,
				}),
			},
		},
		Limit: qdrant.PtrOf(uint64(4)),
	})
	if err != nil {
		t.Fatalf("Filtered query failed: %v", err)
	}

	t.Logf("Filtered query (with time filters) returned %d results", len(filteredResult))
	if len(filteredResult) > 0 {
		t.Logf("First filtered result score: %.4f", filteredResult[0].Score)
	}
}
