package application

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/dbrun3/nexus-vector/handler"
	"github.com/dbrun3/nexus-vector/nexus"
)

func Run() {
	config := &nexus.Config{
		OpenAIKey:      os.Getenv("OPENAI_API_KEY"),
		QdrantHost:     os.Getenv("QDRANT_HOST"),
		RedisHost:      os.Getenv("REDIS_HOST"),
		MongoHost:      os.Getenv("MONGODB_HOST"),
		MongoUser:      os.Getenv("MONGODB_USER"),
		MongoPass:      os.Getenv("MONGODB_PASS"),
		TorchServeHost: os.Getenv("TORCHSERVE_HOST"),
		ModelName:      os.Getenv("MODEL"),
		Env:            nexus.Prod,
	}

	n, err := nexus.InitializeNexus(context.Background(), config)
	if err != nil {
		log.Fatalf("Failed to initialize Nexus: %v", err)
	}

	h := handler.NewHandler(*n)
	mux := h.SetupRoutes()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
