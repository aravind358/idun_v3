package output

import (
	"context"
	"time"
)

// DefaultOutputEngine transforms a CompositeResponse into an OutputDocument.
// It is strictly a transformation layer — it does not access CAS, perform inference,
// or contain modality-specific logic.
// The ResolvedContent field of CompositeResponse is expected to be fully populated
// by the OutputManager before Realize is called.
type DefaultOutputEngine struct{}

// NewDefaultOutputEngine creates a new DefaultOutputEngine.
func NewDefaultOutputEngine() *DefaultOutputEngine {
	return &DefaultOutputEngine{}
}

// Realize transforms a CompositeResponse into an OutputDocument.
// The document's Content is taken directly from CompositeResponse.ResolvedContent,
// which was assembled by the OutputManager from CAS payloads in semantic goal order.
// OutputDocument is the canonical modality-agnostic output object for all plugins.
func (e *DefaultOutputEngine) Realize(ctx context.Context, response CompositeResponse) (OutputDocument, error) {
	return OutputDocument{
		ID:        "out_" + response.ExecutionID,
		Content:   response.ResolvedContent,
		CreatedAt: time.Now(),
	}, nil
}
