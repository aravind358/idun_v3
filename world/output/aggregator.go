package output

import (
	"context"

	"idun/intelligence/executive/v3"
)

// RealizationOptions defines the modality-agnostic presentation parameters for output.
type RealizationOptions struct {
	Tone     string
	Language string
	Style    string
	Intent   string
}

// Slot represents an extracted semantic argument for realization.
type Slot struct {
	Name  string
	Value string
}

// CompositeResponse represents the aggregated results of a multi-intent ExecutionPlan.
//
// - OrderedNodeResults preserves the canonical semantic goal order established by
//   the Understanding subsystem (sorted by NodeResult.Metadata.GoalIndex).
// - NodeResults provides fast lookup by NodeID for backward compatibility.
// - ResolvedContent is the fully realized human-readable content, assembled by the
//   OutputManager from CAS payloads. The OutputEngine reads this field and never
//   touches CAS directly.
// - RealizationOptions specifies the presentation metadata for formatting.
type CompositeResponse struct {
	ExecutionID        string
	NodeResults        map[string]v3.NodeResult
	OrderedNodeResults []v3.NodeResult
	ResolvedContent    string
	Options            RealizationOptions
}

// Aggregator combines multiple NodeResults from an ExecutionResult into a single CompositeResponse.
type Aggregator interface {
	Aggregate(ctx context.Context, execResult *v3.ExecutionResult) (CompositeResponse, error)
}
