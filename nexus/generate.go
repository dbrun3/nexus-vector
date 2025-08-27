package nexus

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dbrun3/nexus-vector/api"
	"github.com/dbrun3/nexus-vector/dao"
	"github.com/dbrun3/nexus-vector/model"
	"github.com/google/uuid"
	"github.com/openai/openai-go/v2"
	"github.com/qdrant/go-client/qdrant"
	"golang.org/x/sync/errgroup"
)

func (n *Nexus) generateNewPages(request api.NexusRequest, syncEmbedding, asyncEmbedding []float32) {
	g, gctx := errgroup.WithContext(context.Background())

	// Generate and store user page with async embedding
	g.Go(func() error {
		userPage, err := n.generateNewUserPage(gctx, request.UserId)
		if err != nil {
			return err
		}
		return n.StorePageInQdrant(gctx, userPage, asyncEmbedding)
	})

	// Generate and store trigger page with sync embedding
	g.Go(func() error {
		triggerPage, err := n.generateNewTriggerPage(gctx, request.Trigger)
		if err != nil {
			return err
		}
		return n.StorePageInQdrant(gctx, triggerPage, syncEmbedding)
	})

	// Wait for both generations to complete
	if err := g.Wait(); err != nil {
		// Log error but don't block - this is background processing
		fmt.Printf("Error generating new pages: %v\n", err)
		return
	}
}

func (n *Nexus) generateNewUserPage(ctx context.Context, userId string) (model.Page, error) {
	// Get user snapshot from MongoDB
	userSnapshot, err := n.mdClient.GetUserSnapshot(ctx, userId)
	if err != nil {
		return model.Page{}, fmt.Errorf("failed to get user snapshot: %w", err)
	}
	if userSnapshot == nil {
		return model.Page{}, fmt.Errorf("user snapshot not found for ID: %s", userId)
	}

	// Marshal user snapshot for ChatGPT
	userJSON, err := json.Marshal(userSnapshot)
	if err != nil {
		return model.Page{}, fmt.Errorf("failed to marshal user snapshot: %w", err)
	}

	chatCompletion, err := n.oaClient.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(fmt.Sprintf("%s\n\nUser Profile: %s", AsyncPrompt, string(userJSON))),
		},
		Model: openai.ChatModelGPT4o,
	})
	if err != nil {
		return model.Page{}, fmt.Errorf("OpenAI API error: %w", err)
	}

	var page model.Page
	err = json.Unmarshal([]byte(stripCodeFences(chatCompletion.Choices[0].Message.Content)), &page)
	if err != nil {
		return model.Page{}, fmt.Errorf("failed to unmarshal page: %w", err)
	}

	return page, nil
}

func (n *Nexus) generateNewTriggerPage(ctx context.Context, trigger model.Trigger) (model.Page, error) {
	// Marshal trigger for ChatGPT
	triggerJSON, err := json.Marshal(trigger)
	if err != nil {
		return model.Page{}, fmt.Errorf("failed to marshal trigger: %w", err)
	}

	chatCompletion, err := n.oaClient.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(fmt.Sprintf("%s\n\nTrigger Context: %s", SyncPrompt, string(triggerJSON))),
		},
		Model: openai.ChatModelGPT4o,
	})
	if err != nil {
		return model.Page{}, fmt.Errorf("OpenAI API error: %w", err)
	}

	var page model.Page
	err = json.Unmarshal([]byte(stripCodeFences(chatCompletion.Choices[0].Message.Content)), &page)
	if err != nil {
		return model.Page{}, fmt.Errorf("failed to unmarshal page: %w", err)
	}

	return page, nil
}

// StorePageInQdrant stores a single page with its embedding in Qdrant
func (n *Nexus) StorePageInQdrant(ctx context.Context, page model.Page, embedding []float32) error {
	// Create payload with page, timestamp, and time range (in this case using a dummy range)
	from := time.Now().Unix()
	until := time.Now().Add(24 * time.Hour).Unix() // Valid for 24 hours
	payload := dao.NewQdrantPagePayload(page, from, until)

	// Generate unique ID for this page
	pointID := uuid.New().String()

	// Create Qdrant point
	point := &qdrant.PointStruct{
		Id:      qdrant.NewID(pointID),
		Vectors: qdrant.NewVectors(embedding...),
		Payload: qdrant.NewValueMap(payload.ToMap()),
	}

	// Store in Qdrant
	_, err := n.qdClient.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: PageCollection,
		Points:         []*qdrant.PointStruct{point},
	})

	if err != nil {
		return fmt.Errorf("failed to store page in Qdrant: %w", err)
	}

	return nil
}

func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
