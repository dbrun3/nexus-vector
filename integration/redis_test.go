package integration

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/redis/go-redis/v9"
)

func Test_Redis_Integration(t *testing.T) {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		t.Skip("REDIS_HOST not set, skipping integration test")
	}

	// Add default port if not specified
	if !strings.Contains(redisHost, ":") {
		redisHost += ":6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	ctx := context.Background()

	// Test ping
	pong, err := client.Ping(ctx).Result()
	if err != nil {
		t.Fatalf("Redis ping failed: %v", err)
	}
	if pong != "PONG" {
		t.Fatalf("Expected PONG, got: %s", pong)
	}

	// Test set
	testKey := "integration_test_key"
	testValue := "integration_test_value"
	
	err = client.Set(ctx, testKey, testValue, 0).Err()
	if err != nil {
		t.Fatalf("Redis set failed: %v", err)
	}

	// Test get
	retrievedValue, err := client.Get(ctx, testKey).Result()
	if err != nil {
		t.Fatalf("Redis get failed: %v", err)
	}

	if retrievedValue != testValue {
		t.Fatalf("Expected '%s', got '%s'", testValue, retrievedValue)
	}

	// Clean up
	err = client.Del(ctx, testKey).Err()
	if err != nil {
		t.Fatalf("Redis delete failed: %v", err)
	}

	t.Logf("Redis integration test successful: set/get '%s'='%s'", testKey, testValue)
}