package decision

// DecisionBuilder implements a fluent builder pattern for assembling a production-ready DecisionService.
type DecisionBuilder struct {
	gate     Tier1ConstitutionalGate
	scorer   Tier2ObjectiveScorer
	provider StrategyProvider
	config   DecisionConfig
}

// NewDecisionBuilder creates a new fluent DecisionBuilder initialized with DefaultDecisionConfig.
func NewDecisionBuilder() *DecisionBuilder {
	return &DecisionBuilder{
		config: DefaultDecisionConfig(),
	}
}

// WithConstitutionalGate injects a custom Tier 1 Constitutional Gate.
func (b *DecisionBuilder) WithConstitutionalGate(gate Tier1ConstitutionalGate) *DecisionBuilder {
	b.gate = gate
	return b
}

// WithObjectiveScorer injects a custom Tier 2 Objective Scorer.
func (b *DecisionBuilder) WithObjectiveScorer(scorer Tier2ObjectiveScorer) *DecisionBuilder {
	b.scorer = scorer
	return b
}

// WithStrategyProvider injects a custom lock-free StrategyProvider.
func (b *DecisionBuilder) WithStrategyProvider(provider StrategyProvider) *DecisionBuilder {
	b.provider = provider
	return b
}

// WithConfig sets a declarative runtime configuration.
func (b *DecisionBuilder) WithConfig(config DecisionConfig) *DecisionBuilder {
	b.config = config
	return b
}

// Build validates components and constructs the configured DecisionService instance.
func (b *DecisionBuilder) Build() (DecisionService, error) {
	gate := b.gate
	if gate == nil {
		gate = NewDefaultConstitutionalGate()
	}

	scorer := b.scorer
	if scorer == nil {
		scorer = NewDefaultObjectiveScorer()
	}

	provider := b.provider
	if provider == nil {
		provider = NewDefaultStrategyProvider(NewDefaultStrategySnapshot(b.config.StrategyVersion))
	}

	service := NewService(
		WithTier1Gate(gate),
		WithTier2Scorer(scorer),
		WithStrategyProvider(provider),
	)

	return service, nil
}
