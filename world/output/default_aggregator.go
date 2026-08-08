package output

import (
	"context"
	"sort"

	"idun/intelligence/executive/v3"
)

// DefaultAggregator implements the Aggregator interface.
type DefaultAggregator struct{}

// NewDefaultAggregator creates a new DefaultAggregator.
func NewDefaultAggregator() *DefaultAggregator {
	return &DefaultAggregator{}
}

// Aggregate combines multiple NodeResults from an ExecutionResult into a single CompositeResponse.
// OrderedNodeResults is sorted by Metadata.GoalIndex to reconstruct the canonical semantic
// presentation order established by the Understanding subsystem, regardless of the order
// in which concurrent goroutines completed.
func (a *DefaultAggregator) Aggregate(ctx context.Context, execResult *v3.ExecutionResult) (CompositeResponse, error) {
	if execResult == nil {
		return CompositeResponse{}, nil
	}

	nodeResults := execResult.NodeResults()

	// Collect into a slice and sort by GoalIndex for deterministic semantic ordering.
	ordered := make([]v3.NodeResult, 0, len(nodeResults))
	for _, nr := range nodeResults {
		ordered = append(ordered, nr)
	}
	sort.Slice(ordered, func(i, j int) bool {
		gi := ordered[i].Metadata.GoalIndex
		gj := ordered[j].Metadata.GoalIndex
		if gi != gj {
			return gi < gj
		}
		// Stable tiebreaker: node IDs are deterministic UUIDs assigned by Planning.
		return ordered[i].NodeID < ordered[j].NodeID
	})

	return CompositeResponse{
		ExecutionID:        execResult.EnvelopeID(),
		NodeResults:        nodeResults,
		OrderedNodeResults: ordered,
	}, nil
}
