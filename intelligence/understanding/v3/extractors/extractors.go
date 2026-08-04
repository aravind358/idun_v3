package extractors

import (
	"idun/intelligence/understanding/v3"
)

// Extractors is the entry point that wires together deterministic extractors.
type Extractors struct {
	Entity    EntityExtractor
	Reference ReferenceExtractor
	Temporal  TemporalExtractor
}

// NewDeterministicExtractors initializes the modular extractors for Phase 4B.3.
func NewDeterministicExtractors() *Extractors {
	return &Extractors{
		Entity:    NewSlotBasedEntityExtractor(),
		Reference: NewReferenceExtractor(),
		Temporal:  NewTemporalExtractor(),
	}
}

// Run executes all extractors against a Hypothesis to populate a SemanticInterpretation Builder.
func (e *Extractors) Run(hyp v3.Hypothesis, b *v3.Builder) {
	b.Entities(e.Entity.Extract(hyp))
	b.References(e.Reference.Extract(hyp))
	b.TemporalAnchors(e.Temporal.Extract(hyp))
}
