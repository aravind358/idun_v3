package v3

import (
	"context"
)

// MemoryProvider defines the minimal abstract interface Reasoning needs to
// query the episodic or semantic memory systems.
type MemoryProvider interface {
	// RetrieveEntity attempts to map a surface name to a canonical database ID.
	RetrieveEntity(ctx context.Context, surfaceName string) (memoryID string, confidence float64, err error)

	// ResolveReference attempts to find the most likely recent target for a pronoun (e.g. "it", "them").
	ResolveReference(ctx context.Context, pronoun string) (targetSurface string, memoryID string, confidence float64, err error)

	// RetrieveContext retrieves relevant factual/episodic snippets based on current intents and topics.
	RetrieveContext(ctx context.Context, intent string, topics []string) ([]ContextEvidence, error)

	// EvaluateCondition checks if a specific condition clause evaluates to true against current world state/memory.
	EvaluateCondition(ctx context.Context, condition string) (bool, error)

	// EvaluateFact checks if a given premise is factually true according to memory.
	EvaluateFact(ctx context.Context, premise string) (bool, error)
}
