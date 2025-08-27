package integration

import (
	"context"
	"os"
	"testing"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

func Test_OpenAI_Integration(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	ctx := context.Background()

	// Test a simple chat completion
	chatCompletion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Say hello"),
		},
		Model: openai.ChatModelGPT4o,
	})

	if err != nil {
		t.Fatalf("OpenAI chat completion failed: %v", err)
	}

	if len(chatCompletion.Choices) == 0 {
		t.Fatal("OpenAI returned no choices")
	}

	if chatCompletion.Choices[0].Message.Content == "" {
		t.Fatal("OpenAI returned empty content")
	}

	t.Logf("OpenAI response: %s", chatCompletion.Choices[0].Message.Content)
}