package torchserve

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/dbrun3/nexus-vector/torchserve/pb"
)

// protoc --go_out=. --go-grpc_out=. --proto_path=internal/torchserve/proto --go_opt=Minference.proto=internal/torchserve/pb --go-grpc_opt=Minference.proto=internal/torchserve/pb internal/torchserve/proto/inference.proto

type Client struct {
	conn   *grpc.ClientConn
	client pb.InferenceAPIsServiceClient
	model  string
}

func NewClient(address, modelName string) (*Client, error) {
	// Add default port if not specified, assumes using proto
	if !strings.Contains(address, ":") {
		address += ":7070"
	}

	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to TorchServe: %w", err)
	}

	client := pb.NewInferenceAPIsServiceClient(conn)

	c := &Client{
		conn:   conn,
		client: client,
		model:  modelName,
	}

	// Verify the server is healthy before returning with retry logic
	maxRetries := 30
	retryDelay := 2 * time.Second

	for attempt := range maxRetries {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, healthErr := c.Health(ctx)
		cancel()

		if healthErr == nil {
			break
		}

		if attempt == maxRetries-1 {
			c.Close()
			return nil, fmt.Errorf("server health check failed after %d attempts: %w", maxRetries, healthErr)
		}

		fmt.Printf("TorchServe health check failed (attempt %d/%d), retrying in %v: %v\n", attempt+1, maxRetries, retryDelay, healthErr)
		time.Sleep(retryDelay)
	}

	return c, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// TextToEmbeddings takes a list of sentences and returns an array of resulting vectors
// Assumes we are using a torchserver that handles pre-processing like https://github.com/clems4ever/torchserve-all-minilm-l6-v2
func (c *Client) TextToEmbeddings(ctx context.Context, texts ...string) ([][]float32, error) {
	// For gRPC, send the texts as individual byte inputs
	input := make(map[string][]byte)

	// Send exact same JSON format as HTTP curl example
	jsonBytes, err := json.Marshal(map[string][]string{"input": texts})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	input["body"] = jsonBytes

	req := &pb.PredictionsRequest{
		ModelName: c.model,
		Input:     input,
	}

	resp, err := c.client.Predictions(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("prediction failed: %w", err)
	}

	// Parse the response as embeddings
	var embeddings64 [][]float64
	if err := json.Unmarshal(resp.Prediction, &embeddings64); err != nil {
		return nil, fmt.Errorf("failed to parse embeddings: %w", err)
	}

	// Convert from float64 to float32
	embeddings := make([][]float32, len(embeddings64))
	for i, vec := range embeddings64 {
		embeddings[i] = make([]float32, len(vec))
		for j, val := range vec {
			embeddings[i][j] = float32(val)
		}
	}

	return embeddings, nil
}

func (c *Client) Health(ctx context.Context) (string, error) {
	resp, err := c.client.Ping(ctx, &emptypb.Empty{})
	if err != nil {
		return "", fmt.Errorf("health check failed: %w", err)
	}

	return resp.Health, nil
}
