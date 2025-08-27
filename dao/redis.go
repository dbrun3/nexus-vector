package dao

import (
	"encoding/json"
	"fmt"
)

// EmbeddingFromRedis converts a JSON array string to []float32
// Example: "[1.5,2.3,0.8]" -> [1.5, 2.3, 0.8]
func EmbeddingFromRedis(s string) ([]float32, error) {
	if s == "" {
		return []float32{}, nil
	}

	var result []float32
	err := json.Unmarshal([]byte(s), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return result, nil
}

// EmbeddingToRedis converts []float32 to a JSON array string
// Example: [1.5, 2.3, 0.8] -> "[1.5,2.3,0.8]"
func EmbeddingToRedis(arr []float32) (string, error) {
	if len(arr) == 0 {
		return "[]", nil
	}

	data, err := json.Marshal(arr)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(data), nil
}
