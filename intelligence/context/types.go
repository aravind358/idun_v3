package context

import (
	"idun/intelligence/understanding"
)

// ResolutionStatus represents the outcome of the context resolution process.
type ResolutionStatus string

const (
	// StatusContextUnnecessary indicates the input was fully explicit and did not require context resolution.
	StatusContextUnnecessary ResolutionStatus = "CONTEXT_UNNECESSARY"

	// StatusResolved indicates implicit context (pronouns, ellipsis, etc.) was successfully resolved.
	StatusResolved ResolutionStatus = "RESOLVED"

	// StatusAmbiguous indicates multiple candidates matched equally well, requiring user disambiguation.
	StatusAmbiguous ResolutionStatus = "AMBIGUOUS"

	// StatusFailed indicates the required context could not be found or resolved.
	StatusFailed ResolutionStatus = "FAILED"
)

// IsValid checks whether the ResolutionStatus is recognized.
func (s ResolutionStatus) IsValid() bool {
	switch s {
	case StatusContextUnnecessary, StatusResolved, StatusAmbiguous, StatusFailed:
		return true
	default:
		return false
	}
}

// Deprecated: ResolvedSemanticFrame is the legacy V1 output of the Context Resolver.
// It is deprecated in favor of natively modifying and returning the V3 underv3.UnderstandingBatch.
// Use UnderstandingBatch directly for all future multi-intent pipelines.
type ResolvedSemanticFrame struct {
	// Frame is the underlying syntactically parsed SemanticFrame from the Understanding subsystem.
	Frame understanding.SemanticFrame `json:"Frame"`

	// Status indicates the outcome of the context resolution process.
	Status ResolutionStatus `json:"Status"`

	// ResolvedEntities maps original slot names (e.g., "target_it") to fully grounded entity IDs.
	// This tracks what the Context Resolver modified or expanded.
	ResolvedEntities map[string]string `json:"ResolvedEntities"`
}
