package context

import (
	"context"

	underv3 "idun/intelligence/understanding/v3"
)

// ResolutionStrategy defines the contract for an individual context resolution algorithm.
// Implementing structs handle specific cases like pronouns, ellipsis, temporal markers, etc.
type ResolutionStrategy interface {
	// Execute evaluates the current interpretation and attempts to resolve it against the dialogue state.
	// It returns a boolean indicating whether it handled/modified the interpretation, and a ResolutionStatus.
	// The boolean `handled` tells the orchestrator if this strategy was applicable.
	Execute(ctx context.Context, orig *underv3.SemanticInterpretation, builder *underv3.Builder, state DialogueStateReader, resolvedEntities map[string]string) (bool, ResolutionStatus)
}
