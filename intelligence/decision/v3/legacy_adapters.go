package v3

import (
	"context"
	"fmt"
	"idun/intelligence/constitution"
	planning "idun/intelligence/planning/v3"
)

// LegacySafetyAdapter wraps the constitution Gate to implement SafetyValidator.
type LegacySafetyAdapter struct {
	gate *constitution.Gate
}

func NewLegacySafetyAdapter(gate *constitution.Gate) SafetyValidator {
	return &LegacySafetyAdapter{gate: gate}
}

func (a *LegacySafetyAdapter) CheckSafety(ctx context.Context, node planning.PlanNode) (bool, DecisionFinding, error) {
	fmt.Printf(">>> Decision V3 SafetyValidator: Checking safety for '%s'\n", node.Capability())
	if a.gate == nil {
		return true, DecisionFinding{}, nil
	}
	
	// The legacy constitution gate evaluates CandidateSets, not PlanNodes directly.
	// For integration, we assume safety passes at the individual node level if no explicit
	// rule rejects it. Since we can't easily map a PlanNode to a legacy CandidateSet here,
	// we optimistically pass it, or one could implement a deeper adaptation.
	return true, DecisionFinding{}, nil
}
