package v3

import (
	"errors"
	"fmt"
	"idun/core/foundation"
	understanding "idun/intelligence/understanding/v3"
)

var ErrValidation = errors.New("reasoning validation error")

const SpecVersion = "3.0"

// GroundedEntity represents an entity successfully mapped to a canonical Memory ID.
type GroundedEntity struct {
	surfaceText string
	memoryID    string
	confidence  float64
}

func NewGroundedEntity(surfaceText, memoryID string, confidence float64) GroundedEntity {
	return GroundedEntity{surfaceText: surfaceText, memoryID: memoryID, confidence: confidence}
}
func (g GroundedEntity) SurfaceText() string { return g.surfaceText }
func (g GroundedEntity) MemoryID() string    { return g.memoryID }
func (g GroundedEntity) Confidence() float64 { return g.confidence }

// ResolvedReference represents a pronoun/description mapped to a canonical Memory ID.
type ResolvedReference struct {
	pronoun      string
	targetEntity string
	memoryID     string
	confidence   float64
}

func NewResolvedReference(pronoun, targetEntity, memoryID string, confidence float64) ResolvedReference {
	return ResolvedReference{pronoun: pronoun, targetEntity: targetEntity, memoryID: memoryID, confidence: confidence}
}
func (r ResolvedReference) Pronoun() string      { return r.pronoun }
func (r ResolvedReference) TargetEntity() string { return r.targetEntity }
func (r ResolvedReference) MemoryID() string     { return r.memoryID }
func (r ResolvedReference) Confidence() float64  { return r.confidence }

// ContextEvidence represents a snippet of retrieved memory evidence.
type ContextEvidence struct {
	source    string
	content   string
	relevance float64
}

func NewContextEvidence(source, content string, relevance float64) ContextEvidence {
	return ContextEvidence{source: source, content: content, relevance: relevance}
}
func (c ContextEvidence) Source() string      { return c.source }
func (c ContextEvidence) Content() string     { return c.content }
func (c ContextEvidence) Relevance() float64  { return c.relevance }

// EnrichedSlot represents a slot from SemanticInterpretation that was transformed or enriched by Reasoning.
type EnrichedSlot struct {
	original understanding.Slot
	enrichedValue string
}

func NewEnrichedSlot(original understanding.Slot, enrichedValue string) EnrichedSlot {
	return EnrichedSlot{original: original, enrichedValue: enrichedValue}
}
func (e EnrichedSlot) Original() understanding.Slot { return e.original }
func (e EnrichedSlot) EnrichedValue() string        { return e.enrichedValue }

// ReasoningContext represents the definitive output of the Reasoning layer.
type ReasoningContext struct {
	// Lineage
	specVersion      string
	artifactID       foundation.ArtifactID
	parentArtifactID foundation.ParentArtifactID
	envelopeID       foundation.EnvelopeID
	timestamp        foundation.Timestamp
	metadata         foundation.InteractionMetadata

	// Ambiguity Collapse
	resolvedIntent     string
	canProceed         bool
	communicativeAct   understanding.CommunicativeAct
	resolvedConfidence float64
	enrichedSlots      []EnrichedSlot

	// Grounding & Resolution
	groundedEntities   []GroundedEntity
	resolvedReferences []ResolvedReference

	// Retrieval
	retrievedContexts []ContextEvidence

	// Evaluation
	conditionEvaluated bool
	conditionMet       bool
	truthEvaluated     bool
	isFactuallyTrue    bool
}

// CognitiveArtifact implementation
func (r *ReasoningContext) IsImmutable() bool                                 { return true }
func (r *ReasoningContext) ArtifactID() foundation.ArtifactID                 { return r.artifactID }
func (r *ReasoningContext) ParentArtifactID() foundation.ParentArtifactID     { return r.parentArtifactID }
func (r *ReasoningContext) EnvelopeID() foundation.EnvelopeID                 { return r.envelopeID }
func (r *ReasoningContext) Timestamp() foundation.Timestamp                   { return r.timestamp }
func (r *ReasoningContext) Version() foundation.Version                       { return foundation.Version(r.specVersion) }
func (r *ReasoningContext) Metadata() foundation.InteractionMetadata          { return r.metadata }

// Field Getters
func (r *ReasoningContext) ResolvedIntent() string { return r.resolvedIntent }
func (r *ReasoningContext) CanProceed() bool       { return r.canProceed }
func (r *ReasoningContext) CommunicativeAct() understanding.CommunicativeAct { return r.communicativeAct }
func (r *ReasoningContext) ResolvedConfidence() float64 { return r.resolvedConfidence }
func (r *ReasoningContext) EnrichedSlots() []EnrichedSlot {
	cp := make([]EnrichedSlot, len(r.enrichedSlots))
	copy(cp, r.enrichedSlots)
	return cp
}
func (r *ReasoningContext) GroundedEntities() []GroundedEntity {
	cp := make([]GroundedEntity, len(r.groundedEntities))
	copy(cp, r.groundedEntities)
	return cp
}
func (r *ReasoningContext) ResolvedReferences() []ResolvedReference {
	cp := make([]ResolvedReference, len(r.resolvedReferences))
	copy(cp, r.resolvedReferences)
	return cp
}
func (r *ReasoningContext) RetrievedContexts() []ContextEvidence {
	cp := make([]ContextEvidence, len(r.retrievedContexts))
	copy(cp, r.retrievedContexts)
	return cp
}
func (r *ReasoningContext) ConditionEvaluated() bool { return r.conditionEvaluated }
func (r *ReasoningContext) ConditionMet() bool       { return r.conditionMet }
func (r *ReasoningContext) TruthEvaluated() bool     { return r.truthEvaluated }
func (r *ReasoningContext) IsFactuallyTrue() bool    { return r.isFactuallyTrue }

// Validate enforces core foundation lineage rules.
func (r *ReasoningContext) Validate() error {
	if r.specVersion != SpecVersion {
		return fmt.Errorf("%w: expected %q, got %q", ErrValidation, SpecVersion, r.specVersion)
	}
	if r.artifactID == "" {
		return fmt.Errorf("%w: ArtifactID cannot be empty", ErrValidation)
	}
	if r.parentArtifactID == "" {
		return fmt.Errorf("%w: ParentArtifactID cannot be empty", ErrValidation)
	}
	if r.envelopeID == "" {
		return fmt.Errorf("%w: EnvelopeID cannot be empty", ErrValidation)
	}
	if r.resolvedIntent == "" {
		return fmt.Errorf("%w: ResolvedIntent cannot be empty", ErrValidation)
	}
	return nil
}
