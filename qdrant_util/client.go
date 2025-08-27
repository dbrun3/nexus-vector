package qdrant_util

import (
	"context"
	"fmt"

	"github.com/qdrant/go-client/qdrant"
)

func NewClient(ctx context.Context, host string, collection string, vectorSize uint64) (*qdrant.Client, error) {
	// setup qdrant
	qdClient, err := qdrant.NewClient(&qdrant.Config{
		Host: host,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Qdrant client: %w", err)
	}

	// check/init qdrant collection
	exists, err := qdClient.CollectionExists(ctx, collection)
	if err != nil {
		return nil, fmt.Errorf("failed to get Qdrant collection: %w", err)
	}
	if !exists {
		if qdClient.CreateCollection(context.Background(), &qdrant.CreateCollection{
			CollectionName: collection,
			VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
				Size:     vectorSize,
				Distance: qdrant.Distance_Cosine,
			}),
		}) != nil {
			return nil, fmt.Errorf("failed to create Qdrant collection: %w", err)

		}
	}

	_, err = qdClient.HealthCheck(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Qdrant healthcheck: %w", err)
	}

	return qdClient, nil
}
