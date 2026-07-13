package reasoning

import (
	"context"

	"idun/intelligence/communication"
)

// DefaultStrategySelector implements StrategySelector to choose a dynamic StrategySpec
// based on incoming perceptual envelope metadata or operational priorities.
type DefaultStrategySelector struct{}

// NewDefaultStrategySelector returns an initialized DefaultStrategySelector.
func NewDefaultStrategySelector() *DefaultStrategySelector {
	return &DefaultStrategySelector{}
}

// SelectStrategy returns a StrategySpec tailored to the incoming perception envelope.
// By default in Phase 2, it selects StrategySymbolicFast for local, deterministic reasoning.
func (s *DefaultStrategySelector) SelectStrategy(ctx context.Context, perceptionEnv communication.Envelope) (StrategySpec, error) {
	if err := ctx.Err(); err != nil {
		return StrategySpec{}, err
	}

	return SelectStrategyForID(StrategySymbolicFast), nil
}

// SelectStrategyForID constructs a canonical StrategySpec for a requested strategy identifier.
func SelectStrategyForID(id StrategyIdentifier) StrategySpec {
	switch id {
	case StrategyAnalogicalBayes:
		return StrategySpec{
			StrategyID: id,
			EnabledStages: []StageIdentifier{
				StageS0ContextAssembly,
				StageS1SymbolicFast,
				StageS4EvidenceFusion,
				StageS5CaseAnalogy,
				StageS6BeamSelection,
				StageS7Calibration,
				StageS9Constitution,
				StageS10ResultAssembly,
			},
			PriorityOrder: []StageIdentifier{
				StageS1SymbolicFast,
				StageS5CaseAnalogy,
				StageS4EvidenceFusion,
			},
			MaxBudgetMs:         25.0,
			MaxGraphNodes:       500,
			MaxGraphEdges:       2000,
			MaxGraphDepth:       3,
			EscalationThreshold: 0.65,
		}
	case StrategyGraphDeliberative:
		return StrategySpec{
			StrategyID: id,
			EnabledStages: []StageIdentifier{
				StageS0ContextAssembly,
				StageS1SymbolicFast,
				StageS2RelationalGraph,
				StageS3CSPCheck,
				StageS6BeamSelection,
				StageS7Calibration,
				StageS8DeliberativeLLM,
				StageS9Constitution,
				StageS10ResultAssembly,
			},
			PriorityOrder: []StageIdentifier{
				StageS1SymbolicFast,
				StageS2RelationalGraph,
				StageS8DeliberativeLLM,
			},
			MaxBudgetMs:         50.0,
			MaxGraphNodes:       500,
			MaxGraphEdges:       2000,
			MaxGraphDepth:       3,
			EscalationThreshold: 0.65,
		}
	case StrategyDeliberativeEscalate:
		return StrategySpec{
			StrategyID: id,
			EnabledStages: []StageIdentifier{
				StageS0ContextAssembly,
				StageS1SymbolicFast,
				StageS5CaseAnalogy,
				StageS6BeamSelection,
				StageS7Calibration,
				StageS8DeliberativeLLM,
				StageS9Constitution,
				StageS10ResultAssembly,
			},
			PriorityOrder: []StageIdentifier{
				StageS1SymbolicFast,
				StageS5CaseAnalogy,
				StageS8DeliberativeLLM,
			},
			MaxBudgetMs:         100.0,
			MaxGraphNodes:       500,
			MaxGraphEdges:       2000,
			MaxGraphDepth:       3,
			EscalationThreshold: 0.65,
		}
	default:
		// StrategySymbolicFast (default deterministic fast path)
		return StrategySpec{
			StrategyID: StrategySymbolicFast,
			EnabledStages: []StageIdentifier{
				StageS0ContextAssembly,
				StageS1SymbolicFast,
				StageS3CSPCheck,
				StageS6BeamSelection,
				StageS7Calibration,
				StageS9Constitution,
				StageS10ResultAssembly,
			},
			PriorityOrder: []StageIdentifier{
				StageS1SymbolicFast,
				StageS3CSPCheck,
			},
			MaxBudgetMs:         15.0,
			MaxGraphNodes:       500,
			MaxGraphEdges:       2000,
			MaxGraphDepth:       3,
			EscalationThreshold: 0.65,
		}
	}
}
