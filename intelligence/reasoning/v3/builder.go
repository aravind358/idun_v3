package v3

import "idun/core/foundation"

// Builder implements the Builder pattern for ReasoningContext to enforce immutability
// and require valid initialization before publishing.
type Builder struct {
	obj *ReasoningContext
}

func NewBuilder() *Builder {
	return &Builder{
		obj: &ReasoningContext{
			specVersion: SpecVersion,
		},
	}
}

func (b *Builder) ArtifactID(id foundation.ArtifactID) *Builder {
	b.obj.artifactID = id
	return b
}
func (b *Builder) ParentArtifactID(id foundation.ParentArtifactID) *Builder {
	b.obj.parentArtifactID = id
	return b
}
func (b *Builder) EnvelopeID(id foundation.EnvelopeID) *Builder {
	b.obj.envelopeID = id
	return b
}
func (b *Builder) Timestamp(t foundation.Timestamp) *Builder {
	b.obj.timestamp = t
	return b
}
func (b *Builder) ResolvedIntent(intent string) *Builder {
	b.obj.resolvedIntent = intent
	return b
}
func (b *Builder) EnrichedSlots(slots []EnrichedSlot) *Builder {
	b.obj.enrichedSlots = make([]EnrichedSlot, len(slots))
	copy(b.obj.enrichedSlots, slots)
	return b
}
func (b *Builder) GroundedEntities(entities []GroundedEntity) *Builder {
	b.obj.groundedEntities = make([]GroundedEntity, len(entities))
	copy(b.obj.groundedEntities, entities)
	return b
}
func (b *Builder) ResolvedReferences(refs []ResolvedReference) *Builder {
	b.obj.resolvedReferences = make([]ResolvedReference, len(refs))
	copy(b.obj.resolvedReferences, refs)
	return b
}
func (b *Builder) RetrievedContexts(ctxs []ContextEvidence) *Builder {
	b.obj.retrievedContexts = make([]ContextEvidence, len(ctxs))
	copy(b.obj.retrievedContexts, ctxs)
	return b
}
func (b *Builder) ConditionEvaluated(evaluated, met bool) *Builder {
	b.obj.conditionEvaluated = evaluated
	b.obj.conditionMet = met
	return b
}
func (b *Builder) TruthEvaluated(evaluated, isTrue bool) *Builder {
	b.obj.truthEvaluated = evaluated
	b.obj.isFactuallyTrue = isTrue
	return b
}

func (b *Builder) Build() (*ReasoningContext, error) {
	if err := b.obj.Validate(); err != nil {
		return nil, err
	}
	return b.obj, nil
}
