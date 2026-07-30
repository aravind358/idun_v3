package v3

import (
	"context"
	understanding "idun/intelligence/understanding/v3"
)

// ReasoningService defines the entry point for Phase 3 cognitive processing.
type ReasoningService interface {
	Reason(ctx context.Context, interpretation *understanding.SemanticInterpretation) (*ReasoningContext, error)
}
