package reasoning

import (
	"errors"
)

// ReasoningResultBuilder constructs validated ReasoningResult instances fluently.
type ReasoningResultBuilder struct {
	result ReasoningResult
	err    error
}

// NewReasoningResultBuilder initializes a builder with schema version and envelope identifiers.
func NewReasoningResultBuilder(envelopeID, sourceFrameID string) *ReasoningResultBuilder {
	b := &ReasoningResultBuilder{
		result: ReasoningResult{
			SchemaVersion:           SchemaVersion,
			EnvelopeID:              envelopeID,
			SourceFrameID:           sourceFrameID,
			Status:                  StatusUnambiguousSolved,
			StrategyUsed:            StrategySymbolicFast,
			AmbiguitySet:            []ReasoningHypothesis{},
			ContradictionsFlagged:   []ContradictionFlag{},
			ProposedBeliefUpdates:   []BeliefUpdateProposal{},
			ConstitutionAnnotations: []string{},
			ReasoningTrace:          []StageTraceLog{},
		},
	}
	if envelopeID == "" {
		b.err = ErrMissingEnvelopeID
	} else if sourceFrameID == "" {
		b.err = ErrMissingSourceFrameID
	}
	return b
}

// WithStatus sets the outcome status.
func (b *ReasoningResultBuilder) WithStatus(status ReasoningOutcomeStatus) *ReasoningResultBuilder {
	if b.err != nil {
		return b
	}
	b.result.Status = status
	return b
}

// WithStrategyUsed sets the strategy routing identifier.
func (b *ReasoningResultBuilder) WithStrategyUsed(strategy StrategyIdentifier) *ReasoningResultBuilder {
	if b.err != nil {
		return b
	}
	b.result.StrategyUsed = strategy
	return b
}

// WithPrimaryHypothesis sets the highest-ranked reasoning hypothesis.
func (b *ReasoningResultBuilder) WithPrimaryHypothesis(hyp ReasoningHypothesis) *ReasoningResultBuilder {
	if b.err != nil {
		return b
	}
	if err := hyp.Validate(); err != nil {
		b.err = err
		return b
	}
	b.result.PrimaryHypothesis = hyp.Clone()
	return b
}

// AddAmbiguityHypothesis appends a runner-up hypothesis within the bounded beam.
func (b *ReasoningResultBuilder) AddAmbiguityHypothesis(hyp ReasoningHypothesis) *ReasoningResultBuilder {
	if b.err != nil {
		return b
	}
	if err := hyp.Validate(); err != nil {
		b.err = err
		return b
	}
	b.result.AmbiguitySet = append(b.result.AmbiguitySet, hyp.Clone())
	return b
}

// AddContradictionFlag records a flagged logical contradiction.
func (b *ReasoningResultBuilder) AddContradictionFlag(cf ContradictionFlag) *ReasoningResultBuilder {
	if b.err != nil {
		return b
	}
	if err := cf.Validate(); err != nil {
		b.err = err
		return b
	}
	b.result.ContradictionsFlagged = append(b.result.ContradictionsFlagged, cf)
	return b
}

// AddBeliefUpdateProposal appends a proposed update to Memory's belief store.
func (b *ReasoningResultBuilder) AddBeliefUpdateProposal(bu BeliefUpdateProposal) *ReasoningResultBuilder {
	if b.err != nil {
		return b
	}
	if err := bu.Validate(); err != nil {
		b.err = err
		return b
	}
	b.result.ProposedBeliefUpdates = append(b.result.ProposedBeliefUpdates, bu)
	return b
}

// WithCompilationCandidate attaches a structured candidate for lifelong rule learning.
func (b *ReasoningResultBuilder) WithCompilationCandidate(cc *CompilationCandidate) *ReasoningResultBuilder {
	if b.err != nil {
		return b
	}
	if cc != nil {
		if err := cc.Validate(); err != nil {
			b.err = err
			return b
		}
		b.result.CompilationCandidate = cc.Clone()
	}
	return b
}

// WithResolvedGoal attaches the machine-readable desired outcome deduced by Reasoning.
func (b *ReasoningResultBuilder) WithResolvedGoal(g *SemanticGoal) *ReasoningResultBuilder {
	if b.err != nil {
		return b
	}
	if g != nil {
		if err := g.Validate(); err != nil {
			b.err = err
			return b
		}
		b.result.ResolvedGoal = g.Clone()
	}
	return b
}

// WithPresentationDirectives attaches presentation metadata specifying how an eventual response may be presented.
func (b *ReasoningResultBuilder) WithPresentationDirectives(p *PresentationDirectives) *ReasoningResultBuilder {
	if b.err != nil {
		return b
	}
	if p != nil {
		if err := p.Validate(); err != nil {
			b.err = err
			return b
		}
		b.result.PresentationDirectives = p.Clone()
	}
	return b
}

// WithStrategyTelemetry attaches diagnostic execution metadata.
func (b *ReasoningResultBuilder) WithStrategyTelemetry(st StrategyTelemetry) *ReasoningResultBuilder {
	if b.err != nil {
		return b
	}
	if err := st.Validate(); err != nil {
		b.err = err
		return b
	}
	b.result.StrategyTelemetry = st.Clone()
	return b
}

// AddConstitutionAnnotation appends a policy note.
func (b *ReasoningResultBuilder) AddConstitutionAnnotation(note string) *ReasoningResultBuilder {
	if b.err != nil {
		return b
	}
	if note != "" {
		b.result.ConstitutionAnnotations = append(b.result.ConstitutionAnnotations, note)
	}
	return b
}

// AddTraceLog appends a stage execution log entry.
func (b *ReasoningResultBuilder) AddTraceLog(log StageTraceLog) *ReasoningResultBuilder {
	if b.err != nil {
		return b
	}
	b.result.ReasoningTrace = append(b.result.ReasoningTrace, log)
	return b
}

// WithOfflineMode sets whether reasoning ran entirely offline.
func (b *ReasoningResultBuilder) WithOfflineMode(offline bool) *ReasoningResultBuilder {
	if b.err != nil {
		return b
	}
	b.result.OfflineMode = offline
	return b
}

// WithProcessedDurationMs sets total latency in milliseconds.
func (b *ReasoningResultBuilder) WithProcessedDurationMs(ms float64) *ReasoningResultBuilder {
	if b.err != nil {
		return b
	}
	b.result.ProcessedDurationMs = ms
	return b
}

// Build validates and returns a deep-copied canonical ReasoningResult.
func (b *ReasoningResultBuilder) Build() (ReasoningResult, error) {
	if b.err != nil {
		return ReasoningResult{}, b.err
	}
	if b.result.PrimaryHypothesis.ID == "" {
		return ReasoningResult{}, errors.New("reasoning: primary hypothesis required")
	}
	if err := b.result.Validate(); err != nil {
		return ReasoningResult{}, err
	}
	return b.result.Clone(), nil
}

// StrategySpecBuilder constructs validated StrategySpec instances fluently.
type StrategySpecBuilder struct {
	spec StrategySpec
	err  error
}

// NewStrategySpecBuilder initializes a builder with default resource bounds.
func NewStrategySpecBuilder(id StrategyIdentifier) *StrategySpecBuilder {
	return &StrategySpecBuilder{
		spec: StrategySpec{
			StrategyID:    id,
			EnabledStages: []StageIdentifier{},
			PriorityOrder: []StageIdentifier{},
			MaxBudgetMs:   15.0,
			MaxGraphNodes: 500,
			MaxGraphEdges: 2000,
			MaxGraphDepth: 3,
		},
	}
}

// EnableStage adds an enabled cascade stage.
func (b *StrategySpecBuilder) EnableStage(stage StageIdentifier) *StrategySpecBuilder {
	if b.err != nil {
		return b
	}
	b.spec.EnabledStages = append(b.spec.EnabledStages, stage)
	return b
}

// WithPriorityOrder sets stage evaluation priority order.
func (b *StrategySpecBuilder) WithPriorityOrder(order []StageIdentifier) *StrategySpecBuilder {
	if b.err != nil {
		return b
	}
	b.spec.PriorityOrder = make([]StageIdentifier, len(order))
	copy(b.spec.PriorityOrder, order)
	return b
}

// WithGraphBounds sets session-scoped graph limits.
func (b *StrategySpecBuilder) WithGraphBounds(nodes, edges, depth int) *StrategySpecBuilder {
	if b.err != nil {
		return b
	}
	b.spec.MaxGraphNodes = nodes
	b.spec.MaxGraphEdges = edges
	b.spec.MaxGraphDepth = depth
	return b
}

// Build validates and returns the StrategySpec.
func (b *StrategySpecBuilder) Build() (StrategySpec, error) {
	if b.err != nil {
		return StrategySpec{}, b.err
	}
	if err := b.spec.Validate(); err != nil {
		return StrategySpec{}, err
	}
	return b.spec.Clone(), nil
}
