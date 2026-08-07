package v3

import (
	"time"
	"idun/core/foundation"
)

// UnderstandingBatch is the single immutable wrapper for all intents recognized
// within a single perception envelope. It contains one or more independent
// SemanticInterpretation objects.
type UnderstandingBatch struct {
	artifactID       foundation.ArtifactID
	parentArtifactID foundation.ParentArtifactID
	envelopeID       foundation.EnvelopeID
	timestamp        foundation.Timestamp
	
	originalUtterance string
	interpretations   []*SemanticInterpretation
}

// NewUnderstandingBatch creates a new batch.
func NewUnderstandingBatch(envID foundation.EnvelopeID, parentID foundation.ParentArtifactID, utterance string, interps []*SemanticInterpretation) *UnderstandingBatch {
	uuidStr, _ := foundation.NewUUID()
	return &UnderstandingBatch{
		artifactID:       foundation.ArtifactID(uuidStr),
		parentArtifactID: parentID,
		envelopeID:       envID,
		timestamp:        foundation.Timestamp(time.Now()),
		originalUtterance: utterance,
		interpretations:   interps,
	}
}

// IsImmutable satisfies foundation.Immutable.
func (b *UnderstandingBatch) IsImmutable() bool { return true }

func (b *UnderstandingBatch) ArtifactID() foundation.ArtifactID         { return b.artifactID }
func (b *UnderstandingBatch) ParentArtifactID() foundation.ParentArtifactID { return b.parentArtifactID }
func (b *UnderstandingBatch) Timestamp() foundation.Timestamp           { return b.timestamp }
func (b *UnderstandingBatch) EnvelopeID() foundation.EnvelopeID         { return b.envelopeID }

// OriginalUtterance returns the original user text, preserved for auditing and tracing.
func (b *UnderstandingBatch) OriginalUtterance() string { return b.originalUtterance }

// Interpretations returns the ordered slice of semantic interpretations.
// The chronological order of split intents is strictly preserved to allow
// downstream subsystems (e.g. Context Resolver) to resolve sequential dependencies.
func (b *UnderstandingBatch) Interpretations() []*SemanticInterpretation {
	result := make([]*SemanticInterpretation, len(b.interpretations))
	copy(result, b.interpretations)
	return result
}
