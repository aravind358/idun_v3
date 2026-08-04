package presentation

import (
	"context"

	"idun/capabilities"
)

// RealizationEngine represents a component capable of transforming structured capability results
// into human-readable RealizedOutput.
type RealizationEngine interface {
	Realize(ctx context.Context, res capabilities.CapabilityResult, parentRef string, responseID string) (*RealizedOutput, error)
}
