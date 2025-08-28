package nexus

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"sort"
	"strings"

	"github.com/dbrun3/nexus-vector/api"
	"github.com/dbrun3/nexus-vector/model"
	"github.com/dbrun3/nexus-vector/mongo"
	"github.com/dbrun3/nexus-vector/qdrant_util"
	"github.com/dbrun3/nexus-vector/util"
	"github.com/qdrant/go-client/qdrant"

	"github.com/dbrun3/nexus-vector/torchserve"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

const PageCollection = "page_collection"
const VectorSize = 384 // default all-minilm-l6-v2 size
const MinScore = 0.9
const NewGenerateChance = 0.1

type Nexus struct {
	oaClient *openai.Client
	qdClient *qdrant.Client
	tsClient *torchserve.Client
	rdClient *redis.Client
	mdClient *mongo.Client
	env      Env
}

func InitializeNexus(ctx context.Context, config *Config) (*Nexus, error) {

	// setup openai
	oaClient := openai.NewClient(
		option.WithAPIKey(config.OpenAIKey),
	)

	// setup redis
	redisHost := config.RedisHost
	if !strings.Contains(redisHost, ":") {
		redisHost += ":6379"
	}
	rdClient := redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// setup mongodb (on prod only)
	mdClient, err := mongo.NewClient(config.MongoHost, config.MongoUser, config.MongoPass)
	if config.Env == Prod && err != nil {
		return nil, fmt.Errorf("failed to create MongoDB client: %w", err)
	}

	// setup qdrant
	qdClient, err := qdrant_util.NewClient(ctx, config.QdrantHost, PageCollection, VectorSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create qdrant client: %w", err)
	}

	// set up torchserve
	tsClient, err := torchserve.NewClient(config.TorchServeHost, config.ModelName)
	if err != nil {
		return nil, fmt.Errorf("failed to create TorchServe client: %w", err)
	}

	return &Nexus{
		oaClient: &oaClient,
		qdClient: qdClient,
		tsClient: tsClient,
		rdClient: rdClient,
		mdClient: mdClient,
		env:      config.Env,
	}, nil
}

// GetNexus returns relevant pages and/or asynchronously creates new one based on the request and its calling user
func (n *Nexus) GetNexus(ctx context.Context, request api.NexusRequest) ([]model.Page, error) {

	g, gctx := errgroup.WithContext(ctx)
	var syncEmbedding []float32
	var asyncEmbedding []float32
	var userResults []*qdrant.ScoredPoint
	var triggerResults []*qdrant.ScoredPoint

	// Fetch pages with "Async Embedding" derivation precomputed from identity combined with long term habits
	g.Go(func() error {
		var err error
		userResults, asyncEmbedding, err = n.getAsyncResults(gctx, request.UserId)
		return err
	})

	// Fetch pages with "Synchronous Embedding" derived from immediate app usage
	g.Go(func() error {
		var err error
		triggerResults, syncEmbedding, err = n.getSyncResults(gctx, request.Trigger)
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Combine results
	allResults := make([]*qdrant.ScoredPoint, len(userResults)+len(triggerResults))
	copy(allResults, userResults)
	copy(allResults[len(userResults):], triggerResults)

	// Sort by relevancy score (descending - highest score first)
	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].Score > allResults[j].Score
	})

	pages := convertResultsToRelevantPages(allResults, MinScore)

	// Chance to generate new pages in the background
	if n.env != Test && (len(pages) == 0 || rand.Float32() < NewGenerateChance) {
		go n.generateNewPages(request, syncEmbedding, asyncEmbedding)
	}

	return pages, nil
}

// InjestUser takes a user snapshot, stores it, and caches its embedding (returning it for debug purposes)
func (n *Nexus) InjestUser(ctx context.Context, request model.UserSnapshot) ([]float32, error) {

	// Mimics slower storage used to query long-term data (not used during benchmarking which only assumes cache hits)
	if n.env == Prod {
		err := n.mdClient.StoreUserSnapshot(ctx, request)
		if err != nil {
			return nil, fmt.Errorf("failed to store user snapshot in MongoDB: %w", err)
		}
	}

	// Embedding is anonymous and simply reflects user trends
	userId := request.ID
	request.ID = ""

	// Clean user snapshot for better embedding generation
	cleanText, err := util.CleanUserSnapshotForEmbedding(request)
	if err != nil {
		return nil, fmt.Errorf("failed to clean user snapshot: %w", err)
	}

	userEmbeddings, err := n.tsClient.TextToEmbeddings(ctx, cleanText)
	if err != nil {
		return nil, fmt.Errorf("failed to create user embedding: %w", err)
	}
	if len(userEmbeddings) != 1 {
		return nil, fmt.Errorf("invalid number of embeddings returned")
	}

	userEmbedding := userEmbeddings[0]

	embeddingBytes, err := json.Marshal(userEmbedding)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding: %w", err)
	}

	err = n.rdClient.Set(ctx, userId, string(embeddingBytes), 0).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to store embedding in Redis: %w", err)
	}

	return userEmbedding, nil
}

// DebugBootstrap generates several random user snapshots and triggers initial GetNexus calls to populate pages
func (n *Nexus) DebugBootstrap(ctx context.Context, count int, seedOffset uint64) ([]string, error) {
	userIds := make([]string, 0, count)

	for i := range count {
		userSnapshot := model.CreateRandomSnapshot(seedOffset + uint64(i))

		_, err := n.InjestUser(ctx, userSnapshot)
		if err != nil {
			return nil, fmt.Errorf("failed to inject user %d: %w", i, err)
		}

		userIds = append(userIds, userSnapshot.ID)

		// Generate random triggers and call GetNexus to populate database with pages via cache misses
		// Create 2-3 different trigger scenarios per user to ensure variety
		triggersPerUser := 2 + (i % 2) // 2 or 3 triggers per user
		for j := 0; j < triggersPerUser; j++ {
			trigger := model.CreateRandomTrigger(seedOffset + uint64(i*10) + uint64(j))

			request := api.NexusRequest{
				UserId:  userSnapshot.ID,
				Trigger: trigger,
			}

			// Call GetNexus to trigger page generation and storage
			_, err := n.GetNexus(ctx, request)
			if err != nil {
				// Log error but don't fail bootstrap - some cache misses are expected
				fmt.Printf("Warning: GetNexus call failed for user %s, trigger %d: %v\n", userSnapshot.ID, j, err)
			}
		}
	}

	return userIds, nil
}

// GetUserSnapshot retrieves a user snapshot from MongoDB by ID
func (n *Nexus) GetUserSnapshot(ctx context.Context, userId string) (*model.UserSnapshot, error) {
	if n.mdClient == nil {
		return nil, fmt.Errorf("MongoDB client not available (test environment)")
	}

	userSnapshot, err := n.mdClient.GetUserSnapshot(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get user snapshot: %w", err)
	}

	return userSnapshot, nil
}

func (n *Nexus) DebugQd() *qdrant.Client {
	return n.qdClient
}

// DebugTrigger creates an embedding for a trigger (for debug purposes)
func (n *Nexus) DebugTrigger(ctx context.Context, trigger model.Trigger) ([]float32, error) {

	// Clean trigger for better embedding generation
	cleanText, err := util.CleanTriggerForEmbedding(trigger)
	if err != nil {
		return nil, fmt.Errorf("failed to clean trigger: %w", err)
	}

	triggerEmbeddings, err := n.tsClient.TextToEmbeddings(ctx, cleanText)
	if err != nil {
		return nil, fmt.Errorf("failed to create trigger embedding: %w", err)
	}
	if len(triggerEmbeddings) != 1 {
		return nil, fmt.Errorf("invalid number of embeddings returned")
	}

	triggerEmbedding := triggerEmbeddings[0]

	return triggerEmbedding, nil
}

// DebugUsersnap creates an embedding for a UserSnap (for debug purposes)
func (n *Nexus) DebugUsersnap(ctx context.Context, userSnap model.UserSnapshot) ([]float32, error) {

	// Clean user snapshot for better embedding generation
	cleanText, err := util.CleanUserSnapshotForEmbedding(userSnap)
	if err != nil {
		return nil, fmt.Errorf("failed to clean user snapshot: %w", err)
	}

	userEmbeddings, err := n.tsClient.TextToEmbeddings(ctx, cleanText)
	if err != nil {
		return nil, fmt.Errorf("failed to create user embedding: %w", err)
	}
	if len(userEmbeddings) != 1 {
		return nil, fmt.Errorf("invalid number of embeddings returned")
	}

	userEmbedding := userEmbeddings[0]

	return userEmbedding, nil
}
