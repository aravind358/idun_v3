// Package embedding implements the Shared Embedding Service for IDUN Intelligence Infrastructure.
//
// Architecture Version: 1.0.0-FROZEN-SPRINT1
//
// The Embedding Service projects multi-modal content into a Canonical Single
// Semantic Vector Space while strictly hiding vector dimensionalities and raw
// float arrays behind opaque content-addressed handles (VectorRef).
package embedding

import (
	"context"
	"errors"
	"time"
)

// Sentinel errors returned by EmbeddingService methods.
var (
	ErrModalityUnsupported = errors.New("embedding: modality unsupported by canonical embedder")
	ErrEmbeddingFailed     = errors.New("embedding: backend vector computation failed")
	ErrServiceClosed       = errors.New("embedding: service closed")
	ErrInvalidRequest      = errors.New("embedding: invalid embedding request")
	ErrVectorNotFound      = errors.New("embedding: opaque vector reference not found in storage")
)

// EmbeddingRequest specifies the content payload to project into canonical vector space.
type EmbeddingRequest struct {
	// ContentRef is the content-addressed storage URI of the source content.
	ContentRef string
	// Modality classifies the source medium ("text", "audio", "structured").
	Modality string
	// ModelID optionally specifies a capability ID override; defaults to canonical system embedder.
	ModelID string
	// CallerID identifies the calling cognitive ability.
	CallerID string
}

// EmbeddingResult returns the immutable, opaque vector reference.
type EmbeddingResult struct {
	// VectorRef is the opaque content-addressed handle in idun/core/storage representing the embedding.
	// Cognitive abilities MUST NEVER inspect raw floats or vector dimensionality.
	VectorRef string
	// ModelID is the logical capability ID used.
	ModelID string
	// BackendID is the physical backend instance ID.
	BackendID string
	// Duration records wall-clock execution latency.
	Duration time.Duration
	// Cached indicates whether the vector was served from exact content cache.
	Cached bool
}

// TelemetrySnapshot captures operational health metrics for Host/Kernel monitoring.
type TelemetrySnapshot struct {
	// ActiveBatches reports concurrent batch embedding operations in progress.
	ActiveBatches int
	// TotalEmbeddings records total single/batch embedding computations completed.
	TotalEmbeddings int64
	// FailedEmbeddings records failed embedding requests.
	FailedEmbeddings int64
	// CacheHits records exact cache hit count.
	CacheHits int64
	// TotalSimilarityQueries records cosine/metric similarity calculations performed.
	TotalSimilarityQueries int64
}

// TelemetryProvider exposes read-only operational telemetry to Host/Kernel monitors.
// Executive Functions and Cognitive Abilities MUST NEVER import or use TelemetryProvider.
type TelemetryProvider interface {
	GetTelemetry() TelemetrySnapshot
}

// EmbeddingService defines the capability interface injected into cognitive abilities.
type EmbeddingService interface {
	// Embed computes or retrieves the canonical vector representation of ContentRef.
	Embed(ctx context.Context, req EmbeddingRequest) (EmbeddingResult, error)

	// EmbedBatch computes vectors for multiple content references concurrently.
	EmbedBatch(ctx context.Context, reqs []EmbeddingRequest) ([]EmbeddingResult, error)

	// Similarity computes semantic metric similarity between two opaque embedding references via infrastructure.
	Similarity(ctx context.Context, vectorRefA, vectorRefB string) (float64, error)

	// Name returns canonical component name ("Intelligence.Infrastructure.Embedding").
	Name() string
	// Start boots the embedding service.
	Start() error
	// Close gracefully shuts down the service.
	Close() error
}
