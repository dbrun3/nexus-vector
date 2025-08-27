# Nexus Vector

An LLM-based recommendation system that delivers personalized nexus pages to users through intelligent content retrieval using vector embeddings and semantic similarity search.

## Overview

Nexus Vector combines synchronous and asynchronous data processing to provide users with highly relevant, personalized content pages. The system leverages machine learning embeddings to match user profiles and real-time triggers with pre-computed content, enabling fast and contextually appropriate recommendations. It is intended to demonstrate the potential of an LLM-based successor to Nexus Service @fetch-rewards.

## Architecture

### Data Processing Pipeline
- **Synchronous Processing**: Short-form real-time trigger events (purchases, redemptions, app interactions) are processed immediately to generate contextual embeddings
- **Asynchronous Processing**: Long-form user profiles and long-term behavioral patterns are processed in the background to create persistent preference embeddings
- **Content Generation**: Uses OpenAI to simulate a rules-based dynamic page creation which would occur on a cache miss.

### Technology Stack
- **Vector Database**: Qdrant for storing and querying pages via embeddings with cosine similarity search
- **ML Embeddings**: TorchServe with all-MiniLM-L6-v2 model for text-to-embedding conversion
- **Caching Layer**: Redis for fast user-snapshot embedding retrieval
- **Data Storage**: MongoDB to simulate longterm user data storage

## Key Features

- **Dual Embedding Strategy**: Combines immediate trigger-based embeddings (post-snap, e-receipt, redemption) with long-term aggregated user preference embeddings
- **Semantic Search**: Uses vector similarity to find the most relevant pre-computed pages
- **Real-time Processing**: Handles embedding user triggers in realtime and delivers resulting pages with minimal latency
- **Intelligent Content Generation**: Leverages OpenAI to simulate asynchronous, rules-based page creation on cache misses
- **Scalable Architecture**: Microservices design with independent scaling capabilities

## Performance

Comprehensive benchmarking results on ARM64 architecture (Apple Silicon with torchserve emulated for AMD64):

### End-to-End Performance
- **Average Response Time**: 16.3ms per GetNexus operation

### Component Performance Breakdown
- **TorchServe Embedding Generation**: 15.1ms per trigger embedding
- **Qdrant Vector Similarity Search**: 0.35ms per query
- **Page Matching Accuracy**: 94.4% of requests return expected similar pages (can vary greatly by seed in a test setting)

### Text Preprocessing Pipeline
The system employs optimized text preprocessing to significantly improve embedding generation performance:

- **JSON Cleaning**: Single-pass character filtering removes structural JSON noise (`{}[]"`) and converts delimiters (`:,`) to spaces
- **Purchase Item Simplification**: Combines brand and product name into unified text, eliminates pricing/quantity details to focus on semantic content
- **Trigger Simplification**: Streamlines complex trigger objects by preserving core semantic information while removing transactional metadata
- **Memory Optimization**: Uses pre-allocated `strings.Builder` with capacity planning to minimize memory allocations during text processing

These preprocessing optimizations resulted in **57.6% faster TorchServe inference** and **43.3% faster end-to-end response times** as opposed to using the raw marshalled json object.

### Performance Analysis
- **Runtime Bottleneck**: TorchServe embedding generation dominates response time (15.1ms vs 0.35ms Qdrant)
- **Vector Operations**: Qdrant performs exceptionally well with sub-millisecond operations
- **Preprocessing Impact**: Optimized text cleaning reduced TorchServe inference time by over half
- **Architecture Impact**: TorchServe performance may be affected by Linux emulation on ARM64

**Test Environment**: Linux ARM64 (Docker on Apple Silicon), Go 1.24.3, 14 CPU cores
**Note**: TorchServe runs in emulated Linux container which may impact ML inference performance on Apple Silicon hardware. Additionally, the number of samples as well as the seed used to generate them largely affects the accuracy of page fetches during benchmarks and minimally impacts embedding time.

The system efficiently handles complex ML workloads with predictable performance characteristics, making it suitable for real-time applications where 20ms response times are acceptable.

## Getting Started

### Prerequisites
- Docker & Docker Compose
- Go 1.24.3+
- OpenAI API key (optional for benchmarking)

### Quick Start
```bash
# Clone the repository
git clone <repository-url>
cd nexus-vector

# Set environment variables
export OPENAI_API_KEY=your_api_key_here

# Run integration tests
make integration

# Run benchmarks
make bench

# Run the full application
make run
```

### API Usage
The system exposes a REST API endpoint for querying pages with a given userId and event trigger:
```
POST /get-nexus
```
There is also an endpoint to simulate the asynchronous injection of user information for caching snapshot embeddings
```
PUT /inject-nexus
```

## Project Structure
- `/nexus` - Core recommendation engine and business logic
- `/model` - Data models for users, triggers, and pages
- `/handler` - HTTP request handlers and routing
- `/benchmark` - Performance testing and metrics
- `/integration` - End-to-end integration tests
- `/torchserve` - ML embedding service client
- `/qdrant_util` - Vector database utilities