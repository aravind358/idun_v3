package planning

import (
	"errors"
	"fmt"
	"time"
)

// ============================================================================
// PlanningRequestBuilder
// ============================================================================

// PlanningRequestBuilder provides a fluent API for constructing valid PlanningRequest instances.
type PlanningRequestBuilder struct {
	req *PlanningRequest
	err error
}

// NewPlanningRequestBuilder initializes a builder with safe baseline defaults.
func NewPlanningRequestBuilder() *PlanningRequestBuilder {
	return &PlanningRequestBuilder{
		req: &PlanningRequest{
			Domain:             "General",
			TargetDepth:        DepthTactical,
			MaxExecutionBudget: 250 * time.Millisecond,
			MinConfidenceFloor: 0.70,
			Metadata:           make(map[string]string),
		},
	}
}

// WithRequestID sets the unique RequestID.
func (b *PlanningRequestBuilder) WithRequestID(id string) *PlanningRequestBuilder {
	if b.err != nil {
		return b
	}
	b.req.RequestID = id
	return b
}

// WithGoal sets the target goal string.
func (b *PlanningRequestBuilder) WithGoal(goal string) *PlanningRequestBuilder {
	if b.err != nil {
		return b
	}
	b.req.Goal = goal
	return b
}

// WithDomain sets the open string domain tag.
func (b *PlanningRequestBuilder) WithDomain(domain string) *PlanningRequestBuilder {
	if b.err != nil {
		return b
	}
	if domain != "" {
		b.req.Domain = domain
	}
	return b
}

// WithContextRef sets the upstream correlation URI (`PayloadRef`).
func (b *PlanningRequestBuilder) WithContextRef(ref string) *PlanningRequestBuilder {
	if b.err != nil {
		return b
	}
	b.req.ContextRef = ref
	return b
}

// WithConstraints sets the hard and soft constraint sets.
func (b *PlanningRequestBuilder) WithConstraints(hard, soft []string) *PlanningRequestBuilder {
	if b.err != nil {
		return b
	}
	if hard != nil {
		b.req.HardConstraints = append(b.req.HardConstraints, hard...)
	}
	if soft != nil {
		b.req.SoftConstraints = append(b.req.SoftConstraints, soft...)
	}
	return b
}

// WithTargetDepth specifies the desired planning depth tier.
func (b *PlanningRequestBuilder) WithTargetDepth(depth PlanningDepth) *PlanningRequestBuilder {
	if b.err != nil {
		return b
	}
	b.req.TargetDepth = depth
	return b
}

// WithBudget sets the maximum execution budget duration and confidence floor.
func (b *PlanningRequestBuilder) WithBudget(budget time.Duration, confidenceFloor float64) *PlanningRequestBuilder {
	if b.err != nil {
		return b
	}
	if budget <= 0 {
		b.err = errors.New("execution budget must be positive")
		return b
	}
	if confidenceFloor < 0.0 || confidenceFloor > 1.0 {
		b.err = fmt.Errorf("confidence floor out of bounds [0.0, 1.0]: %f", confidenceFloor)
		return b
	}
	b.req.MaxExecutionBudget = budget
	b.req.MinConfidenceFloor = confidenceFloor
	return b
}

// Build validates and returns the final PlanningRequest.
func (b *PlanningRequestBuilder) Build() (*PlanningRequest, error) {
	if b.err != nil {
		return nil, b.err
	}
	if err := b.req.Validate(); err != nil {
		return nil, fmt.Errorf("failed to build PlanningRequest: %w", err)
	}
	return b.req, nil
}

// ============================================================================
// PlanBuilder
// ============================================================================

// PlanBuilder provides a fluent API for constructing valid Plan artifacts.
type PlanBuilder struct {
	plan *Plan
	err  error
}

// NewPlanBuilder initializes a builder enforcing canonical SchemaVersion and current timestamps.
func NewPlanBuilder() *PlanBuilder {
	return &PlanBuilder{
		plan: &Plan{
			SchemaVersion: SchemaVersion2_0_0,
			CreatedAt:     time.Now().UTC(),
			Status:        PlanStatusComplete,
			Domain:        "General",
		},
	}
}

// WithIdentity sets core correlation IDs for the Plan.
func (b *PlanBuilder) WithIdentity(planID, strategySnapshotID, traceID string) *PlanBuilder {
	if b.err != nil {
		return b
	}
	b.plan.PlanID = planID
	b.plan.StrategySnapshotID = strategySnapshotID
	b.plan.TraceID = traceID
	return b
}

// WithGoalAndDomain sets the target goal string and domain tag.
func (b *PlanBuilder) WithGoalAndDomain(goal, domain, sourceTier string) *PlanBuilder {
	if b.err != nil {
		return b
	}
	b.plan.Goal = goal
	if domain != "" {
		b.plan.Domain = domain
	}
	b.plan.SourceTier = sourceTier
	return b
}

// AddSubgoal appends a validated Subgoal to the decomposition.
func (b *PlanBuilder) AddSubgoal(sg Subgoal) *PlanBuilder {
	if b.err != nil {
		return b
	}
	if err := sg.Validate(); err != nil {
		b.err = fmt.Errorf("invalid subgoal: %w", err)
		return b
	}
	b.plan.Subgoals = append(b.plan.Subgoals, sg)
	return b
}

// AddDependency appends a validated DependencyEdge to the plan graph.
func (b *PlanBuilder) AddDependency(dep DependencyEdge) *PlanBuilder {
	if b.err != nil {
		return b
	}
	if err := dep.Validate(); err != nil {
		b.err = fmt.Errorf("invalid dependency edge: %w", err)
		return b
	}
	b.plan.Dependencies = append(b.plan.Dependencies, dep)
	return b
}

// WithEstimates sets resource, cost, and duration estimates.
func (b *PlanBuilder) WithEstimates(cost float64, duration time.Duration, resources []ResourceRequirement) *PlanBuilder {
	if b.err != nil {
		return b
	}
	if cost < 0 {
		b.err = fmt.Errorf("negative cost: %f", cost)
		return b
	}
	b.plan.EstimatedCost = cost
	b.plan.EstimatedDuration = duration
	for i, res := range resources {
		if err := res.Validate(); err != nil {
			b.err = fmt.Errorf("resource[%d] invalid: %w", i, err)
			return b
		}
	}
	b.plan.RequiredResources = append(b.plan.RequiredResources, resources...)
	return b
}

// WithConfidenceProfile sets the 6-dimensional confidence profile.
func (b *PlanBuilder) WithConfidenceProfile(cp ConfidenceProfile) *PlanBuilder {
	if b.err != nil {
		return b
	}
	if err := cp.Validate(); err != nil {
		b.err = fmt.Errorf("invalid confidence profile: %w", err)
		return b
	}
	b.plan.ConfidenceProfile = cp
	return b
}

// WithStatus sets the operational status and any structured information requirements.
func (b *PlanBuilder) WithStatus(status PlanStatus, reqs []InformationRequirement) *PlanBuilder {
	if b.err != nil {
		return b
	}
	b.plan.Status = status
	for i, r := range reqs {
		if err := r.Validate(); err != nil {
			b.err = fmt.Errorf("info requirement[%d] invalid: %w", i, err)
			return b
		}
	}
	b.plan.InformationRequirements = append(b.plan.InformationRequirements, reqs...)
	return b
}

// WithReplayMetadata attaches deterministic replay provenance.
func (b *PlanBuilder) WithReplayMetadata(rm ReplayMetadata) *PlanBuilder {
	if b.err != nil {
		return b
	}
	if err := rm.Validate(); err != nil {
		b.err = fmt.Errorf("invalid replay metadata: %w", err)
		return b
	}
	b.plan.ReplayMetadata = rm
	return b
}

// WithFingerprint explicitly sets the structural content hash.
func (b *PlanBuilder) WithFingerprint(fp string) *PlanBuilder {
	if b.err != nil {
		return b
	}
	b.plan.PlanFingerprint = fp
	return b
}

// Build validates and returns the constructed Plan artifact.
func (b *PlanBuilder) Build() (*Plan, error) {
	if b.err != nil {
		return nil, b.err
	}
	if err := b.plan.Validate(); err != nil {
		return nil, fmt.Errorf("failed to build Plan: %w", err)
	}
	return b.plan, nil
}

// ============================================================================
// PlanningTraceBuilder
// ============================================================================

// PlanningTraceBuilder provides a fluent API for constructing valid PlanningTrace diagnostic artifacts.
type PlanningTraceBuilder struct {
	trace *PlanningTrace
	err   error
}

// NewPlanningTraceBuilder initializes a builder enforcing canonical SchemaVersion.
func NewPlanningTraceBuilder() *PlanningTraceBuilder {
	return &PlanningTraceBuilder{
		trace: &PlanningTrace{
			SchemaVersion: SchemaVersion2_0_0,
		},
	}
}

// WithIdentity sets core trace correlation IDs.
func (b *PlanningTraceBuilder) WithIdentity(traceID, planID, snapshotID string) *PlanningTraceBuilder {
	if b.err != nil {
		return b
	}
	b.trace.TraceID = traceID
	b.trace.PlanID = planID
	b.trace.StrategySnapshotID = snapshotID
	b.trace.ReplayMetadata.StrategySnapshotID = snapshotID
	return b
}

// AddStepLog appends a specialist step execution log to the trace.
func (b *PlanningTraceBuilder) AddStepLog(step PlanningStepLog) *PlanningTraceBuilder {
	if b.err != nil {
		return b
	}
	b.trace.PlanningSteps = append(b.trace.PlanningSteps, step)
	return b
}

// WithProvenance sets immutable fingerprint and replay metadata on the diagnostic trace.
func (b *PlanningTraceBuilder) WithProvenance(policyFP, capabilityFP, searchStrategyID string, rm ReplayMetadata) *PlanningTraceBuilder {
	if b.err != nil {
		return b
	}
	if err := rm.Validate(); err != nil {
		b.err = fmt.Errorf("invalid replay metadata: %w", err)
		return b
	}
	b.trace.PolicyFingerprint = policyFP
	b.trace.CapabilityFingerprint = capabilityFP
	b.trace.SearchStrategyID = searchStrategyID
	b.trace.ReplayMetadata = rm
	return b
}

// AddSpecialistUsage appends factual usage telemetry for a specialist to the trace.
func (b *PlanningTraceBuilder) AddSpecialistUsage(usage PlanningSpecialistUsage) *PlanningTraceBuilder {
	if b.err != nil {
		return b
	}
	if err := usage.Validate(); err != nil {
		b.err = fmt.Errorf("invalid specialist usage: %w", err)
		return b
	}
	b.trace.SpecialistUsage = append(b.trace.SpecialistUsage, usage)
	return b
}

// AddRejectedBranch appends a discarded branch with exact structural discard rationale.
func (b *PlanningTraceBuilder) AddRejectedBranch(rb RejectedBranch) *PlanningTraceBuilder {
	if b.err != nil {
		return b
	}
	if err := rb.Validate(); err != nil {
		b.err = fmt.Errorf("invalid rejected branch: %w", err)
		return b
	}
	b.trace.RejectedBranches = append(b.trace.RejectedBranches, rb)
	return b
}

// WithDiagnostics sets search statistics, termination reason, complexity, and quality metrics.
func (b *PlanningTraceBuilder) WithDiagnostics(
	reason PlanningTerminationReason,
	stats SearchStatistics,
	complexity float64,
	cp ConfidenceProfile,
	qm QualityMetrics,
) *PlanningTraceBuilder {
	if b.err != nil {
		return b
	}
	if reason == "" {
		b.err = errors.New("missing termination reason")
		return b
	}
	if complexity < 0 {
		b.err = fmt.Errorf("negative complexity: %f", complexity)
		return b
	}
	if err := cp.Validate(); err != nil {
		b.err = fmt.Errorf("invalid confidence profile: %w", err)
		return b
	}
	if err := qm.Validate(); err != nil {
		b.err = fmt.Errorf("invalid quality metrics: %w", err)
		return b
	}
	b.trace.TerminationReason = reason
	b.trace.SearchStatistics = stats
	b.trace.EstimatedComplexity = complexity
	b.trace.ConfidenceProfile = cp
	b.trace.QualityMetrics = qm
	return b
}

// Build validates and returns the constructed PlanningTrace artifact.
func (b *PlanningTraceBuilder) Build() (*PlanningTrace, error) {
	if b.err != nil {
		return nil, b.err
	}
	if err := b.trace.Validate(); err != nil {
		return nil, fmt.Errorf("failed to build PlanningTrace: %w", err)
	}
	return b.trace, nil
}
