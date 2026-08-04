package v3

import (
	"context"
	"fmt"
	planning "idun/intelligence/planning/v3"
)

// DefaultAuthValidator provides a baseline implementation of AuthValidator.
// It acts as the extension point for future external permission systems.
type DefaultAuthValidator struct{}

func NewDefaultAuthValidator() *DefaultAuthValidator {
	return &DefaultAuthValidator{}
}

func (v *DefaultAuthValidator) CheckPermissions(ctx context.Context, node planning.PlanNode) (bool, DecisionFinding, error) {
	fmt.Printf(">>> Decision V3 AuthValidator: Checking permissions for '%s'\n", node.Capability())
	// Baseline implementation: unconditionally allows capability invocation.
	finding := NewDecisionFinding(
		"Authorization",
		node.NodeID(),
		true,
		"AUTH_PASSED",
		fmt.Sprintf("Capability '%s' authorized", node.Capability()),
	)
	return true, finding, nil
}

// DefaultPolicyValidator provides a baseline implementation of PolicyValidator.
// It acts as the extension point for future policy enforcement engines (e.g., OPA).
type DefaultPolicyValidator struct{}

func NewDefaultPolicyValidator() *DefaultPolicyValidator {
	return &DefaultPolicyValidator{}
}

func (v *DefaultPolicyValidator) CheckPolicy(ctx context.Context, node planning.PlanNode) (bool, DecisionFinding, error) {
	fmt.Printf(">>> Decision V3 PolicyValidator: Checking policy for '%s'\n", node.Capability())
	// Baseline implementation: unconditionally conforms to policy.
	finding := NewDecisionFinding(
		"Policy",
		node.NodeID(),
		true,
		"POLICY_CONFORMS",
		fmt.Sprintf("Capability '%s' conforms to baseline policy", node.Capability()),
	)
	return true, finding, nil
}

// DefaultBudgetValidator provides a baseline implementation of BudgetValidator.
// It acts as the extension point for future integration with the Executive's BudgetManager.
type DefaultBudgetValidator struct{}

func NewDefaultBudgetValidator() *DefaultBudgetValidator {
	return &DefaultBudgetValidator{}
}

func (v *DefaultBudgetValidator) CheckBudget(ctx context.Context, node planning.PlanNode) (bool, DecisionFinding, error) {
	fmt.Printf(">>> Decision V3 BudgetValidator: Checking budget for '%s'\n", node.Capability())
	// Baseline implementation: unconditionally passes budget checks.
	finding := NewDecisionFinding(
		"Budget",
		node.NodeID(),
		true,
		"BUDGET_AVAILABLE",
		fmt.Sprintf("Sufficient budget available for capability '%s'", node.Capability()),
	)
	return true, finding, nil
}
