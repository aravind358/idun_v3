package context

import (
	"context"
	"time"

	underv3 "idun/intelligence/understanding/v3"
)

// ContextResolver is the primary interface for resolving ambiguous or implicit context
// within a semantic frame against the current dialogue state.
type ContextResolver interface {
	// Resolve takes a V3 UnderstandingBatch and grounds any implicit references
	// (pronouns, ellipsis, temporal markers) against the provided DialogueStateReader.
	// It returns a modified UnderstandingBatch with resolved entities and statuses.
	Resolve(ctx context.Context, batch *underv3.UnderstandingBatch, state DialogueStateReader) (*underv3.UnderstandingBatch, error)
}

// DialogueStateReader defines the read-only contract the ContextResolver requires
// to interrogate working memory, recent conversation history, and active goals.
// By depending on this interface rather than a concrete state object, the resolver
// enforces the Interface Segregation Principle.
type DialogueStateReader interface {
	// GetRecentCandidates returns recently mentioned entities matching the requested semantic role.
	GetRecentCandidates(role string, limit int) []string

	// GetActiveGoals returns the current list of unsatisfied goals/intents the user is pursuing.
	GetActiveGoals() []string
	
	// GetPreviousBatch returns the semantic batch from the preceding conversational turn.
	// This is required for conversational ellipsis reconstruction (e.g. "and tomorrow").
	GetPreviousBatch() *underv3.UnderstandingBatch

	// GetTemporalAnchor returns the canonical "now" for relative time resolution.
	// E.g., used to resolve "tomorrow".
	GetTemporalAnchor() time.Time
}
