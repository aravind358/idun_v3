package v3

import (
	"context"
	reasoning "idun/intelligence/reasoning/v3"
	understanding "idun/intelligence/understanding/v3"
)

// PlanningService defines the entry point for Phase 4 cognitive processing.
type PlanningService interface {
	Plan(ctx context.Context, interp *understanding.SemanticInterpretation, reasonCtx *reasoning.ReasoningContext) (*ExecutionPlan, error)
}
