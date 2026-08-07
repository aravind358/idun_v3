package v3

import (
	"fmt"
	"idun/core/foundation"
)

// Builder constructs a SemanticInterpretation.
type Builder struct {
	obj *SemanticInterpretation
}

func NewBuilder() *Builder {
	return &Builder{
		obj: &SemanticInterpretation{
			specVersion: SpecVersion,
		},
	}
}

// CloneBuilder creates a Builder initialized with a deep copy of an existing interpretation.
func CloneBuilder(orig *SemanticInterpretation) *Builder {
	// Deep copy slices to ensure mutation safety
	clone := *orig
	
	clone.topics = make([]string, len(orig.topics))
	copy(clone.topics, orig.topics)
	
	clone.entities = make([]Entity, len(orig.entities))
	copy(clone.entities, orig.entities)
	
	clone.references = make([]Reference, len(orig.references))
	copy(clone.references, orig.references)
	
	clone.temporalAnchors = make([]TemporalAnchor, len(orig.temporalAnchors))
	copy(clone.temporalAnchors, orig.temporalAnchors)
	
	clone.composedTimestamps = make([]string, len(orig.composedTimestamps))
	copy(clone.composedTimestamps, orig.composedTimestamps)
	
	clone.openSlots = make([]string, len(orig.openSlots))
	copy(clone.openSlots, orig.openSlots)
	
	clone.secondaryIntents = make([]SecondaryIntent, len(orig.secondaryIntents))
	copy(clone.secondaryIntents, orig.secondaryIntents)
	
	clone.assumptions = make([]Assumption, len(orig.assumptions))
	copy(clone.assumptions, orig.assumptions)
	
	clone.ambiguities = make([]Ambiguity, len(orig.ambiguities))
	copy(clone.ambiguities, orig.ambiguities)

	return &Builder{
		obj: &clone,
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
func (b *Builder) Status(s InterpretationStatus) *Builder {
	b.obj.status = s
	return b
}
func (b *Builder) PrimaryIntent(i string) *Builder {
	b.obj.primaryIntent = i
	return b
}
func (b *Builder) CommunicativeAct(a CommunicativeAct) *Builder {
	b.obj.communicativeAct = a
	return b
}
func (b *Builder) PrimaryHypothesis(h Hypothesis) *Builder {
	b.obj.primaryHypothesis = h
	return b
}
func (b *Builder) AmbiguitySet(a []Hypothesis) *Builder {
	b.obj.ambiguitySet = a
	return b
}
func (b *Builder) Topics(t []string) *Builder {
	b.obj.topics = t
	return b
}
func (b *Builder) Entities(e []Entity) *Builder {
	b.obj.entities = e
	return b
}
func (b *Builder) Slots(s []Slot) *Builder {
	b.obj.primaryHypothesis.slots = s
	return b
}
func (b *Builder) References(r []Reference) *Builder {
	b.obj.references = r
	return b
}
func (b *Builder) TemporalAnchors(t []TemporalAnchor) *Builder {
	b.obj.temporalAnchors = t
	return b
}
func (b *Builder) GetTemporalAnchors() []TemporalAnchor {
	return b.obj.temporalAnchors
}
func (b *Builder) ComposedTimestamps(c []string) *Builder {
	b.obj.composedTimestamps = c
	return b
}
func (b *Builder) GetComposedTimestamps() []string {
	return b.obj.composedTimestamps
}
func (b *Builder) OpenSlots(o []string) *Builder {
	b.obj.openSlots = o
	return b
}

func (b *Builder) IsConditional(c bool) *Builder {
	b.obj.isConditional = c
	return b
}
func (b *Builder) ConditionClause(c string) *Builder {
	b.obj.conditionClause = c
	return b
}
func (b *Builder) ConsequentClause(c string) *Builder {
	b.obj.consequentClause = c
	return b
}
func (b *Builder) CompoundIntentCount(c int) *Builder {
	b.obj.compoundIntentCount = c
	return b
}
func (b *Builder) SecondaryIntents(s []SecondaryIntent) *Builder {
	b.obj.secondaryIntents = s
	return b
}
func (b *Builder) RequiresContext(r bool) *Builder {
	b.obj.requiresContext = r
	return b
}
func (b *Builder) Polarity(p Polarity) *Builder {
	b.obj.polarity = p
	return b
}
func (b *Builder) Sentiment(s Sentiment) *Builder {
	b.obj.sentiment = s
	return b
}
func (b *Builder) InputCharacteristics(i InputCharacteristics) *Builder {
	b.obj.inputCharacteristics = i
	return b
}
func (b *Builder) DialoguePosition(d DialoguePosition) *Builder {
	b.obj.dialoguePosition = d
	return b
}
func (b *Builder) ExecutionHints(e ExecutionHints) *Builder {
	b.obj.executionHints = e
	return b
}
func (b *Builder) Assumptions(a []Assumption) *Builder {
	b.obj.assumptions = a
	return b
}
func (b *Builder) Ambiguities(a []Ambiguity) *Builder {
	b.obj.ambiguities = a
	return b
}


func (b *Builder) Confidence(c float64) *Builder {
	b.obj.confidence = c
	return b
}
func (b *Builder) GoalIndex(g int) *Builder {
	b.obj.goalIndex = g
	return b
}
func (b *Builder) TotalGoals(t int) *Builder {
	b.obj.totalGoals = t
	return b
}

// Build constructs and validates the SemanticInterpretation.
func (b *Builder) Build() (*SemanticInterpretation, error) {
	if err := b.obj.Validate(); err != nil {
		return nil, fmt.Errorf("build failed: %w", err)
	}
	return b.obj, nil
}
