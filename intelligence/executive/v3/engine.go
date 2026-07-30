package v3

import (
	"context"
	"fmt"
	"idun/core/foundation"
	"time"
)

// ExecutionEngine serves as the orchestrator for the Executive layer Phase 6.
// It manages the Episode Lifecycle, enforces timeout budgets, and integrates with the Workspace.
// It remains completely content-blind.
type ExecutionEngine struct {
	planProvider PlanProvider
	dagExecutor  *DAGExecutor
}

// NewExecutionEngine initializes the Execution Engine.
func NewExecutionEngine(planProvider PlanProvider, registry CapabilityRegistry, memory MemoryProvider) *ExecutionEngine {
	return &ExecutionEngine{
		planProvider: planProvider,
		dagExecutor:  NewDAGExecutor(registry, memory),
	}
}

// ExecuteDecision takes an approved DecisionRecord and orchestrates its execution.
// It returns an immutable ExecutionResult.
func (e *ExecutionEngine) ExecuteDecision(ctx context.Context, decisionArtifactID foundation.ArtifactID, planArtifactID foundation.ArtifactID, envelopeID string, budget time.Duration) (*ExecutionResult, error) {
	// 1. Fetch ExecutionPlan from Memory (via abstract PlanProvider)
	plan, err := e.planProvider.GetPlan(ctx, planArtifactID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve ExecutionPlan: %w", err)
	}

	// 2. Bound the execution Context based on allocated budget
	execCtx, cancel := context.WithTimeout(ctx, budget)
	defer cancel()

	start := time.Now()

	// 3. Traverse and execute DAG
	nodeResults, status, dagErr := e.dagExecutor.Execute(execCtx, plan)

	totalDuration := time.Since(start)

	// 4. Build Immutable ExecutionResult Artifact
	builder := NewBuilder().
		WithParentArtifactID(decisionArtifactID).
		WithEnvelopeID(envelopeID).
		WithStatus(status).
		WithTotalDuration(totalDuration)

	if dagErr != nil {
		builder = builder.WithOverallError(dagErr.Error())
	}

	for _, nr := range nodeResults {
		builder = builder.AddNodeResult(nr)
	}

	result, buildErr := builder.Build()
	if buildErr != nil {
		return nil, fmt.Errorf("failed to build ExecutionResult: %w", buildErr)
	}

	return result, nil
}
