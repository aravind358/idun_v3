package v3

import (
	"context"
	planning "idun/intelligence/planning/v3"
	reasoning "idun/intelligence/reasoning/v3"
	understanding "idun/intelligence/understanding/v3"
)

// DecisionService defines the entry point for Phase 5 cognitive processing.
type DecisionService interface {
	Decide(
		ctx context.Context,
		interp *understanding.SemanticInterpretation,
		reasonCtx *reasoning.ReasoningContext,
		plan *planning.ExecutionPlan,
	) (*DecisionRecord, error)
}
