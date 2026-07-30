package v3

import (
	"context"
	planning "idun/intelligence/planning/v3"
)

// SafetyValidator abstracts external safety evaluation models.
type SafetyValidator interface {
	CheckSafety(ctx context.Context, node planning.PlanNode) (passed bool, finding DecisionFinding, err error)
}

// PolicyValidator abstracts external policy enforcement engines (e.g. OPA).
type PolicyValidator interface {
	CheckPolicy(ctx context.Context, node planning.PlanNode) (passed bool, finding DecisionFinding, err error)
}

// AuthValidator abstracts external permission systems.
type AuthValidator interface {
	CheckPermissions(ctx context.Context, node planning.PlanNode) (passed bool, finding DecisionFinding, err error)
}

// BudgetValidator abstracts external resource trackers.
type BudgetValidator interface {
	CheckBudget(ctx context.Context, node planning.PlanNode) (passed bool, finding DecisionFinding, err error)
}
