// Package v3 implements the Decision layer for IDUN Architecture 2.0.0.
//
// Architectural Note: Planning selects, Decision validates.
// In V1, the Decision layer received multiple candidate actions and performed
// multi-criteria scoring and option ranking. In V3, the Planning layer resolves
// intents into a single concrete ExecutionPlan. Therefore, multi-candidate
// evaluation has been superseded. Decision V3 acts purely as a validation gate
// (evaluating Safety, Auth, Policy, and Budget) on the proposed plan.
package v3

import (
	"context"
	"idun/core/foundation"
	planning "idun/intelligence/planning/v3"
	reasoning "idun/intelligence/reasoning/v3"
	understanding "idun/intelligence/understanding/v3"
	"strings"
	"time"
)

// Orchestrator coordinates the Decision pipeline.
type Orchestrator struct {
	safety SafetyValidator
	auth   AuthValidator
	policy PolicyValidator
	budget BudgetValidator
}

// NewOrchestrator creates a new decision Orchestrator with the given validators.
func NewOrchestrator(safety SafetyValidator, auth AuthValidator, policy PolicyValidator, budget BudgetValidator) *Orchestrator {
	return &Orchestrator{
		safety: safety,
		auth:   auth,
		policy: policy,
		budget: budget,
	}
}

// Decide evaluates an ExecutionPlan against safety, policy, permissions, and budget validators.
func (o *Orchestrator) Decide(
	ctx context.Context,
	interp *understanding.SemanticInterpretation,
	reasonCtx *reasoning.ReasoningContext,
	plan *planning.ExecutionPlan,
) (*DecisionRecord, error) {

	uuidStr, _ := foundation.NewUUID()
	artifactID := foundation.ArtifactID(uuidStr)

	builder := NewBuilder().
		ArtifactID(artifactID).
		ParentArtifactID(foundation.ParentArtifactID(plan.ArtifactID())). // Trace back to ExecutionPlan
		EnvelopeID(plan.EnvelopeID()).
		Timestamp(foundation.Timestamp(time.Now()))

	safetyPass := true
	policyPass := true
	authPass := true
	budgetPass := true
	var allFindings []DecisionFinding

	for _, node := range plan.Nodes() {
		// 1. Safety Check
		if o.safety != nil {
			pass, finding, _ := o.safety.CheckSafety(ctx, node)
			if !pass {
				safetyPass = false
				allFindings = append(allFindings, finding)
			}
		}

		// 2. Policy Check
		if o.policy != nil {
			pass, finding, _ := o.policy.CheckPolicy(ctx, node)
			if !pass {
				policyPass = false
				allFindings = append(allFindings, finding)
			}
		}

		// 3. Permissions Check
		if o.auth != nil {
			pass, finding, _ := o.auth.CheckPermissions(ctx, node)
			if !pass {
				authPass = false
				allFindings = append(allFindings, finding)
			}
		}

		// 4. Budget Check
		if o.budget != nil {
			pass, finding, _ := o.budget.CheckBudget(ctx, node)
			if !pass {
				budgetPass = false
				allFindings = append(allFindings, finding)
			}
		}
	}

	builder.EvaluationFlags(safetyPass, policyPass, authPass, budgetPass)
	builder.Findings(allFindings)

	// Determine final resolution
	if safetyPass && policyPass && authPass && budgetPass {
		builder.Resolution(StatusApproved)
		builder.Reason("All evaluations passed unconditionally.")
	} else {
		builder.Resolution(StatusRejected)
		var errs []string
		for _, f := range allFindings {
			errs = append(errs, f.Message())
		}
		builder.Reason("Violations found: " + strings.Join(errs, "; "))
	}

	return builder.Build()
}
