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
// - ResolvedData contains the structured capability payloads extracted from CAS.
// - Options specifies the presentation metadata for formatting.
type CompositeResponse struct {
	ExecutionID        string
	NodeResults        map[string]v3.NodeResult
	OrderedNodeResults []v3.NodeResult
	ResolvedData       []OutputPayload
	Options            RealizationOptions
}

// PrimaryResponseType safely extracts the dominant response type for the payload, eliminating slice ordering assumptions.
func (c *CompositeResponse) PrimaryResponseType() ResponseType {
	if len(c.ResolvedData) > 0 {
		return c.ResolvedData[0].ResponseType
	}
	return ""
}

// Aggregator combines multiple NodeResults from an ExecutionResult into a single CompositeResponse.
type Aggregator interface {
	Aggregate(ctx context.Context, execResult *v3.ExecutionResult) (CompositeResponse, error)
}
