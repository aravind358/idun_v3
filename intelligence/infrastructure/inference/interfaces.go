// Package inference implements the Shared Inference Service for IDUN Intelligence Infrastructure.
//
// Architecture Version: 1.0.0-FROZEN-SPRINT1
//
// The Inference Service provides cognitive abilities with a shared, thread-safe,
// budget-governed computational engine for analytical, generative, and causal
// inference without coupling callers to specific AI architectures or backends.
package inference

import (
	"context"
	"errors"
	"time"
)

// Sentinel errors returned by InferenceService methods.
var (
	ErrInferenceTimeout   = errors.New("inference: execution exceeded budget SLA timeout")
	ErrInferenceCancelled = errors.New("inference: request cancelled by caller")
	ErrBackendFailure     = errors.New("inference: physical backend execution failed")
	ErrServiceClosed      = errors.New("inference: service closed")
	ErrInvalidRequest     = errors.New("inference: invalid inference request")
)

// Modality classifies input and output payload formats.
type Modality string

const (
	// ModalityText denotes unstructured or natural language text.
	ModalityText Modality = "text"
	// ModalityStructured denotes typed JSON or semantic frame structures.
	ModalityStructured Modality = "structured"
	// ModalityMultimodal denotes multi-channel sensory or perceptual data.
	ModalityMultimodal Modality = "multimodal"
)

// ExecutionHints provides backend-agnostic execution guidance without coupling to LLM hyperparameters.
type ExecutionHints struct {
	// ExploratoryVariance ranges from 0.0 (strictly deterministic) to 1.0 (high exploratory breadth).
	ExploratoryVariance float64
	// ComputeBudgetUnits hints at the upper bound on computational steps or resource units.
	ComputeBudgetUnits int
	// OutputDetailHint suggests response verbosity ("compact", "standard", "comprehensive").
	OutputDetailHint string
}

// InferenceRequest defines the input contract submitted by a cognitive ability.
type InferenceRequest struct {
	// ModelID identifies the logical capability to execute (resolved via registry.Resolver).
	ModelID string
	// InputRef is the content-addressed reference URI of input data stored in idun/core/storage.
	InputRef string
	// Modality declares the payload format.
	Modality Modality
	// Budget specifies priority tier ("REFLEXIVE", "STANDARD", "DELIBERATIVE").
	Budget string
	// Hints provides backend-agnostic execution guidance.
	Hints ExecutionHints
	// CallerID identifies the calling cognitive ability ("Understanding", "Reasoning", etc.).
	CallerID string
}

// InferenceResult defines the output contract returned to a cognitive ability.
type InferenceResult struct {
	// OutputRef is the content-addressed URI of output data written to idun/core/storage.
	OutputRef string
	// ModelID is the logical capability ID that fulfilled the request.
	ModelID string
	// BackendID is the physical backend instance ID resolved by registry.Resolver.
	BackendID string
	// ComputeUnits records normalized computational units consumed.
	ComputeUnits int
	// ExecutionDuration records actual processing wall-clock time.
	ExecutionDuration time.Duration
	// Cached indicates whether the result was served from exact content-addressed cache.
	Cached bool
}

// StreamChunk represents an incremental output chunk for interactive streaming workflows.
type StreamChunk struct {
	// ChunkRef is the storage URI or inline reference for this segment.
	ChunkRef string
	// Done is true when streaming has completed.
	Done bool
	// Error contains any stream execution failure.
	Error error
}

// TelemetrySnapshot captures operational health and worker metrics for Host/Kernel monitoring.
type TelemetrySnapshot struct {
	// ActiveWorkers reports the number of concurrently executing inference tasks.
	ActiveWorkers int
	// QueueDepths reports pending tasks per Budget tier.
	QueueDepths map[string]int
	// TotalExecutions counts total requests processed.
	TotalExecutions int64
	// FailedExecutions counts total requests failed.
	FailedExecutions int64
	// CacheHits counts total requests served from exact storage cache.
	CacheHits int64
}

// TelemetryProvider exposes operational telemetry strictly to Host/Kernel monitors.
// Executive Functions and Cognitive Abilities MUST NEVER import or use TelemetryProvider.
type TelemetryProvider interface {
	GetTelemetry() TelemetrySnapshot
}

// InferenceService defines the capability interface injected into cognitive abilities.
type InferenceService interface {
	// Execute runs inference synchronously or via priority queues and returns immutable storage references.
	Execute(ctx context.Context, req InferenceRequest) (InferenceResult, error)

	// ExecuteStream runs streaming inference, delivering output chunks to the provided channel.
	ExecuteStream(ctx context.Context, req InferenceRequest, stream chan<- StreamChunk) error

	// Name returns canonical component name ("Intelligence.Infrastructure.Inference").
	Name() string
	// Start boots worker pools and queue processing.
	Start() error
	// Close gracefully shuts down workers and rejects new requests.
	Close() error
}
