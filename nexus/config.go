package nexus

type Env string

const Test Env = "test"
const Prod Env = "prod"

// Config holds all configuration parameters for initializing Nexus
type Config struct {
	// OpenAI configuration
	OpenAIKey string

	// Qdrant configuration
	QdrantHost string

	// Redis configuration
	RedisHost string

	// MongoDB configuration
	MongoHost string
	MongoUser string
	MongoPass string

	// TorchServe configuration
	TorchServeHost string
	ModelName      string

	Env Env
}
