package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/dbrun3/nexus-vector/model"
	"github.com/dbrun3/nexus-vector/nexus"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

func Test_OpenAI_Prompt_Integration(t *testing.T) {
	// Check required environment variables
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	ctx := context.Background()

	// Initialize OpenAI client
	oaClient := openai.NewClient(
		option.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
	)

	// Test data
	testUser := model.CreateRandomSnapshot(12345)
	testTrigger := model.CreateRandomTrigger(67890)

	t.Run("SyncPrompt", func(t *testing.T) {
		// Marshal trigger for ChatGPT
		triggerJSON, err := json.Marshal(testTrigger)
		if err != nil {
			t.Fatalf("Failed to marshal trigger: %v", err)
		}

		// Call OpenAI with sync prompt
		chatCompletion, err := oaClient.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage(fmt.Sprintf("%s\n\nTrigger Context: %s", nexus.SyncPrompt, string(triggerJSON))),
			},
			Model: openai.ChatModelGPT4o,
		})
		if err != nil {
			t.Fatalf("OpenAI API error: %v", err)
		}

		// Test that response can be unmarshalled as a Page
		var page model.Page
		responseContent := stripCodeFences(chatCompletion.Choices[0].Message.Content)
		t.Logf("SyncPrompt response: %s", responseContent)

		err = json.Unmarshal([]byte(responseContent), &page)
		if err != nil {
			t.Fatalf("Failed to unmarshal sync prompt response as Page: %v", err)
		}

		// Verify basic page structure (values don't matter, just structure)
		if page.Layout == "" {
			t.Error("SyncPrompt response missing 'layout' field")
		}
		if page.Type == "" {
			t.Error("SyncPrompt response missing 'type' field")
		}
		if page.Category == "" {
			t.Error("SyncPrompt response missing 'category' field")
		}
		if len(page.Title) == 0 {
			t.Error("SyncPrompt response missing 'title' field")
		}
		if len(page.SubTitle) == 0 {
			t.Error("SyncPrompt response missing 'subTitle' field")
		}

		t.Logf("SyncPrompt successfully generated valid Page: Layout=%s, Type=%s, Category=%s",
			page.Layout, page.Type, page.Category)
	})

	t.Run("AsyncPrompt", func(t *testing.T) {
		// Marshal user snapshot for ChatGPT
		userJSON, err := json.Marshal(testUser)
		if err != nil {
			t.Fatalf("Failed to marshal user snapshot: %v", err)
		}

		// Call OpenAI with async prompt
		chatCompletion, err := oaClient.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage(fmt.Sprintf("%s\n\nUser Profile: %s", nexus.AsyncPrompt, string(userJSON))),
			},
			Model: openai.ChatModelGPT4o,
		})
		if err != nil {
			t.Fatalf("OpenAI API error: %v", err)
		}

		// Test that response can be unmarshalled as a Page
		var page model.Page
		responseContent := stripCodeFences(chatCompletion.Choices[0].Message.Content)
		t.Logf("AsyncPrompt response: %s", responseContent)

		err = json.Unmarshal([]byte(responseContent), &page)
		if err != nil {
			t.Fatalf("Failed to unmarshal async prompt response as Page: %v", err)
		}

		// Verify basic page structure (values don't matter, just structure)
		if page.Layout == "" {
			t.Error("AsyncPrompt response missing 'layout' field")
		}
		if page.Type == "" {
			t.Error("AsyncPrompt response missing 'type' field")
		}
		if page.Category == "" {
			t.Error("AsyncPrompt response missing 'category' field")
		}
		if len(page.Title) == 0 {
			t.Error("AsyncPrompt response missing 'title' field")
		}
		if len(page.SubTitle) == 0 {
			t.Error("AsyncPrompt response missing 'subTitle' field")
		}

		t.Logf("AsyncPrompt successfully generated valid Page: Layout=%s, Type=%s, Category=%s",
			page.Layout, page.Type, page.Category)
	})
}

func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
