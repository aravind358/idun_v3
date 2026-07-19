// Package planning implements IDUN's Intelligence Pillar Planning cognitive ability.
//
// As defined in IDUN V3 Planning Architecture Specification v2.0.0-FROZEN,
// Planning (`idun/intelligence/planning`) is responsible for multi-step goal
// decomposition, Hierarchical Task Networks (HTNs), task sequencing, dependency
// analysis, resource estimation, and bounded contingency generation.
//
// Planning maintains strict single-responsibility boundaries: it constructs
// immutable, versioned Plan objects paired with comprehensive PlanningTrace
// diagnostic artifacts for Decision and Reflection, but never interprets raw
// sensory inputs (`Understanding`), derives formal logical proofs (`Reasoning`),
// commits to action execution (`Decision`), performs post-hoc evaluations (`Reflection`),
// or mutates strategic policy weights (`Learning`).
package planning

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"idun/intelligence/reasoning"
)

// SchemaVersion2_0_0 is the frozen canonical schema version string for Planning artifacts.
const SchemaVersion2_0_0 = "2.0.0-FROZEN"

// SchemaVersion2_0 is the minor-compatible schema version alias for Planning artifacts.
const SchemaVersion2_0 = "2.0"

// ============================================================================
// Core Status and Termination Enums
// ============================================================================

// PlanStatus represents the discrete operational status of a constructed Plan.
type PlanStatus string

const (
	// PlanStatusComplete indicates the plan successfully decomposes the goal within budget.
	PlanStatusComplete PlanStatus = "COMPLETE"

	// PlanStatusPartialBudgetExhausted indicates planning stopped due to budget limits before completing all branches.
	PlanStatusPartialBudgetExhausted PlanStatus = "PARTIAL_BUDGET_EXHAUSTED"

	// PlanStatusInfeasible indicates the goal cannot be achieved given current constraints and resources.
	PlanStatusInfeasible PlanStatus = "INFEASIBLE"

	// PlanStatusConstraintConflict indicates mutually exclusive or irreconcilable constraints were detected.
	PlanStatusConstraintConflict PlanStatus = "CONSTRAINT_CONFLICT"

	// PlanStatusInsufficientInfo indicates critical data gaps prevented feasible plan generation.
	PlanStatusInsufficientInfo PlanStatus = "INSUFFICIENT_INFORMATION"
)

// PlanningResultStatus describes what Planning produced at the service evaluation level,
// completely independent of PlanningTerminationReason (which describes why search stopped).
type PlanningResultStatus string

const (
	ResultSuccess               PlanningResultStatus = "RESULT_SUCCESS"
	ResultPartialPlans          PlanningResultStatus = "RESULT_PARTIAL_PLANS"
	ResultNoPlans               PlanningResultStatus = "RESULT_NO_PLANS"
	ResultEscalationRecommended PlanningResultStatus = "RESULT_ESCALATION_RECOMMENDED"
	ResultAbstained             PlanningResultStatus = "RESULT_ABSTAINED"
	ResultValidationFailed      PlanningResultStatus = "RESULT_VALIDATION_FAILED"
)

// PlanningTerminationReason records the factual reason why planning execution terminated.
// This enum lives exclusively within PlanningTrace and never inside Plan.
type PlanningTerminationReason string

const (
	TerminationGoalFound            PlanningTerminationReason = "GOAL_FOUND"
	TerminationSearchBudgetExceeded PlanningTerminationReason = "SEARCH_BUDGET_EXHAUSTED"
	TerminationTimeLimit            PlanningTerminationReason = "TIME_LIMIT"
	TerminationNodeLimit            PlanningTerminationReason = "NODE_LIMIT"
	TerminationNoValidPlan          PlanningTerminationReason = "NO_VALID_PLAN"
	TerminationInsufficientInfo     PlanningTerminationReason = "INSUFFICIENT_INFORMATION"
	TerminationExecutiveCancelled   PlanningTerminationReason = "EXECUTIVE_CANCELLED"
	TerminationHigherPriority       PlanningTerminationReason = "HIGHER_PRIORITY_INTERRUPT"
	TerminationConstitutionBlocked  PlanningTerminationReason = "CONSTITUTIONAL_BLOCK"
)

// PlanningDepth identifies the computational horizon and specialist composition tier.
type PlanningDepth string

const (
	DepthReflexive PlanningDepth = "REFLEXIVE"
	DepthTactical  PlanningDepth = "TACTICAL"
	DepthStrategic PlanningDepth = "STRATEGIC"
)

// EscalationAction represents an explicit recommendation emitted when current depth is insufficient.
type EscalationAction string

const (
	ActionNone                      EscalationAction = "NONE"
	ActionRecommendMorePlanning     EscalationAction = "RECOMMEND_MORE_PLANNING"
	ActionRecommendHigherDepth      EscalationAction = "RECOMMEND_HIGHER_PLANNING_DEPTH"
)

// ============================================================================
// Supporting Plan Structures
// ============================================================================

// Subgoal represents a hierarchical node within the decomposition tree.
type Subgoal struct {
	SubgoalID    string            `json:"subgoal_id"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	AssignedType string            `json:"assigned_type"`
	Dependencies []string          `json:"dependencies"`
	Parameters   map[string]string `json:"parameters"`
}

// Validate verifies structural soundness of the Subgoal.
func (s *Subgoal) Validate() error {
	if s.SubgoalID == "" {
		return errors.New("subgoal missing SubgoalID")
	}
	if s.Title == "" {
		return fmt.Errorf("subgoal %s missing Title", s.SubgoalID)
	}
	return nil
}

// DependencyEdge defines a directed temporal or causal ordering constraint between nodes.
type DependencyEdge struct {
	EdgeID          string `json:"edge_id"`
	SourceNodeID    string `json:"source_node_id"`
	TargetNodeID    string `json:"target_node_id"`
	DependencyType  string `json:"dependency_type"` // e.g., "HARD_PREREQUISITE", "SOFT_PREFERENCE", "DATA_FLOW"
	IsBlocking      bool   `json:"is_blocking"`
}

// Validate verifies structural soundness of the DependencyEdge.
func (d *DependencyEdge) Validate() error {
	if d.EdgeID == "" || d.SourceNodeID == "" || d.TargetNodeID == "" {
		return errors.New("dependency edge missing required IDs (EdgeID, SourceNodeID, TargetNodeID)")
	}
	if d.SourceNodeID == d.TargetNodeID {
		return fmt.Errorf("dependency edge %s self-loop detected: %s", d.EdgeID, d.SourceNodeID)
	}
	return nil
}

// ResourceRequirement specifies a physical, computational, or financial resource needed by the plan.
type ResourceRequirement struct {
	ResourceID   string  `json:"resource_id"`
	ResourceType string  `json:"resource_type"` // e.g., "GPU_UNITS", "MEMORY_MB", "NETWORK_BANDWIDTH", "BUDGET_USD"
	Quantity     float64 `json:"quantity"`
	IsOptional   bool    `json:"is_optional"`
}

// Validate verifies structural soundness of the ResourceRequirement.
func (r *ResourceRequirement) Validate() error {
	if r.ResourceID == "" || r.ResourceType == "" {
		return errors.New("resource requirement missing ResourceID or ResourceType")
	}
	if r.Quantity < 0 {
		return fmt.Errorf("resource requirement %s has negative quantity: %f", r.ResourceID, r.Quantity)
	}
	return nil
}

// RollbackStrategy specifies a pre-planned recovery sequence if an execution step fails.
type RollbackStrategy struct {
	StrategyID    string   `json:"strategy_id"`
	TriggerNodeID string   `json:"trigger_node_id"`
	ActionSteps   []string `json:"action_steps"`
	EstimatedCost float64  `json:"estimated_cost"`
}

// Validate verifies structural soundness of the RollbackStrategy.
func (r *RollbackStrategy) Validate() error {
	if r.StrategyID == "" || r.TriggerNodeID == "" {
		return errors.New("rollback strategy missing StrategyID or TriggerNodeID")
	}
	return nil
}

// AlternativeBranch represents a viable, ranked alternative path surfaced inside the plan.
type AlternativeBranch struct {
	BranchID           string            `json:"branch_id"`
	Title              string            `json:"title"`
	Description        string            `json:"description"`
	CostDelta          float64           `json:"cost_delta"`
	DurationDelta      time.Duration     `json:"duration_delta"`
	ConfidenceDelta    float64           `json:"confidence_delta"`
	TradeoffAttributes map[string]float64 `json:"tradeoff_attributes"`
}

// Validate verifies structural soundness of the AlternativeBranch.
func (a *AlternativeBranch) Validate() error {
	if a.BranchID == "" || a.Title == "" {
		return errors.New("alternative branch missing BranchID or Title")
	}
	return nil
}

// InformationRequirement expresses a structured diagnostic gap when data is missing.
type InformationRequirement struct {
	MissingItem          string `json:"missing_item"`
	Blocking             bool   `json:"blocking"`
	RequestingSpecialist string `json:"requesting_specialist"`
	SuggestedSource      string `json:"suggested_source"`
}

// Validate verifies structural soundness of the InformationRequirement.
func (i *InformationRequirement) Validate() error {
	if i.MissingItem == "" {
		return errors.New("information requirement missing MissingItem")
	}
	return nil
}

// ============================================================================
// Multi-Dimensional Confidence & Replay Metadata
// ============================================================================

// ConfidenceProfile captures 6-dimensional epistemic certainty scores for the plan.
// OverallConfidence is constitutionally bounded to be <= min(all 6 dimensions).
type ConfidenceProfile struct {
	GoalConfidence         float64 `json:"goal_confidence"`
	PreconditionConfidence float64 `json:"precondition_confidence"`
	DependencyConfidence   float64 `json:"dependency_confidence"`
	ResourceConfidence     float64 `json:"resource_confidence"`
	TimingConfidence       float64 `json:"timing_confidence"`
	ConstraintConfidence   float64 `json:"constraint_confidence"`
	OverallConfidence      float64 `json:"overall_confidence"`
}

// Validate verifies that all dimensional scores are within [0.0, 1.0] and checks minimum aggregation bounding.
func (c *ConfidenceProfile) Validate() error {
	dims := []float64{
		c.GoalConfidence, c.PreconditionConfidence, c.DependencyConfidence,
		c.ResourceConfidence, c.TimingConfidence, c.ConstraintConfidence,
	}
	minVal := 1.0
	for _, v := range dims {
		if v < 0.0 || v > 1.0 {
			return fmt.Errorf("confidence dimension out of bounds [0.0, 1.0]: %f", v)
		}
		if v < minVal {
			minVal = v
		}
	}
	if c.OverallConfidence < 0.0 || c.OverallConfidence > 1.0 {
		return fmt.Errorf("OverallConfidence out of bounds [0.0, 1.0]: %f", c.OverallConfidence)
	}
	// Enforce strict minimum bounding (allowing tiny float error tolerance 1e-9)
	if c.OverallConfidence > minVal+1e-9 {
		return fmt.Errorf("OverallConfidence (%f) exceeds minimum dimensional confidence (%f)", c.OverallConfidence, minVal)
	}
	return nil
}

// ReplayMetadata records deterministic replay provenance for audit and scientific regression.
type ReplayMetadata struct {
	StrategySnapshotID    string   `json:"strategy_snapshot_id"`
	SpecialistVersions    []string `json:"specialist_versions"`
	InputHashes           []string `json:"input_hashes"`
	SeedOrProvenanceToken string   `json:"seed_or_provenance_token"`
	ReplayFidelity        string   `json:"replay_fidelity"` // e.g., "EXACT", "BEST_EFFORT", "NOT_SUPPORTED"
	ReplaySeed            uint64   `json:"replay_seed"`
	WorkingMemoryHash     string   `json:"working_memory_hash"`
}

// Validate verifies required fields in ReplayMetadata.
func (r *ReplayMetadata) Validate() error {
	if r.StrategySnapshotID == "" {
		return errors.New("replay metadata missing StrategySnapshotID")
	}
	if r.ReplayFidelity != "EXACT" && r.ReplayFidelity != "BEST_EFFORT" && r.ReplayFidelity != "NOT_SUPPORTED" && r.ReplayFidelity != "" {
		return fmt.Errorf("invalid ReplayFidelity: %s", r.ReplayFidelity)
	}
	return nil
}

// ============================================================================
// Trace Diagnostic Structures (Exclusively inside PlanningTrace)
// ============================================================================

// SearchStatistics records lightweight numerical metrics from the planning search process.
type SearchStatistics struct {
	NodesExpanded             uint64 `json:"nodes_expanded"`
	NodesPruned               uint64 `json:"nodes_pruned"`
	BeamWidthUsed             uint32 `json:"beam_width_used"`
	AlternativePlansGenerated uint32 `json:"alternative_plans_generated"`
	ConstraintViolations      uint32 `json:"constraint_violations"`
	DeadEndsReached           uint32 `json:"dead_ends_reached"`
	CacheHits                 uint64 `json:"cache_hits"`
	CacheMisses               uint64 `json:"cache_misses"`
}

// QualityMetrics quantifies a priori structural properties of the plan's dependency graph and estimates.
type QualityMetrics struct {
	Completeness           float64       `json:"completeness"`
	Efficiency             float64       `json:"efficiency"`
	Robustness             float64       `json:"robustness"`
	Flexibility            float64       `json:"flexibility"`
	ResourceEfficiency     float64       `json:"resource_efficiency"`
	ExpectedExecutionCost  float64       `json:"expected_execution_cost"`
	EstimatedExecutionTime time.Duration `json:"estimated_execution_time"`
	RiskExposure           float64       `json:"risk_exposure"`
	DependencyComplexity   float64       `json:"dependency_complexity"`
	Maintainability        float64       `json:"maintainability"`
	Adaptability           float64       `json:"adaptability"`
}

// Validate checks bounds on QualityMetrics scores.
func (q *QualityMetrics) Validate() error {
	scores := []float64{
		q.Completeness, q.Efficiency, q.Robustness, q.Flexibility,
		q.ResourceEfficiency, q.RiskExposure, q.DependencyComplexity,
		q.Maintainability, q.Adaptability,
	}
	for _, s := range scores {
		if s < 0.0 || s > 1.0 {
			return fmt.Errorf("quality metric out of bounds [0.0, 1.0]: %f", s)
		}
	}
	if q.ExpectedExecutionCost < 0 {
		return fmt.Errorf("negative expected execution cost: %f", q.ExpectedExecutionCost)
	}
	return nil
}

// RejectedBranch records an alternative branch that was considered and discarded during search,
// capturing exact structural rationale for Learning and Reflection.
type RejectedBranch struct {
	BranchID      string  `json:"branch_id"`
	Description   string  `json:"description"`
	DiscardReason string  `json:"discard_reason"` // e.g., "ResourceConflict: GPU quota exceeded"
	DiscardStage  string  `json:"discard_stage"`
	ScoreDelta    float64 `json:"score_delta"`
}

// Validate verifies structural soundness of RejectedBranch.
func (r *RejectedBranch) Validate() error {
	if r.BranchID == "" || r.DiscardReason == "" {
		return errors.New("rejected branch missing BranchID or DiscardReason")
	}
	return nil
}

// SpecialistSkipReason defines factual, observational reasons why a specialist was not invoked.
type SpecialistSkipReason string

const (
	SkipNone                     SpecialistSkipReason = "NONE"
	SkipCapabilityDisabled       SpecialistSkipReason = "CAPABILITY_DISABLED"
	SkipDomainMismatch           SpecialistSkipReason = "DOMAIN_MISMATCH"
	SkipBudgetExceeded           SpecialistSkipReason = "BUDGET_EXCEEDED"
	SkipNoApplicableGoal         SpecialistSkipReason = "NO_APPLICABLE_GOAL"
	SkipHigherPrioritySpecialist SpecialistSkipReason = "HIGHER_PRIORITY_SPECIALIST"
	SkipCancelled                SpecialistSkipReason = "CANCELLED"
)

// PlanningSpecialistUsage records bounded, factual diagnostic usage telemetry
// for a specialist evaluated during a planning episode.
type PlanningSpecialistUsage struct {
	SpecialistID      string               `json:"specialist_id"`
	Invoked           bool                 `json:"invoked"`
	SkipReason        SpecialistSkipReason `json:"skip_reason"`
	ContributionScore float32              `json:"contribution_score"`
	NodesExpanded     uint64               `json:"nodes_expanded"`
	NodesPruned       uint64               `json:"nodes_pruned"`
	PlansGenerated    uint32               `json:"plans_generated"`
	ExecutionTimeUs   uint64               `json:"execution_time_us"`
	Success           bool                 `json:"success"`
}

// Validate ensures structural soundness of PlanningSpecialistUsage.
func (u *PlanningSpecialistUsage) Validate() error {
	if u.SpecialistID == "" {
		return errors.New("PlanningSpecialistUsage missing SpecialistID")
	}
	if u.ContributionScore < 0.0 || u.ContributionScore > 1.0 {
		return fmt.Errorf("PlanningSpecialistUsage %s ContributionScore out of bounds [0.0, 1.0]: %f", u.SpecialistID, u.ContributionScore)
	}
	if !u.Invoked && u.SkipReason == "" {
		return fmt.Errorf("PlanningSpecialistUsage %s not invoked but missing SkipReason", u.SpecialistID)
	}
	return nil
}

// PlanningStepLog records a single step executed by an invoked planning specialist.
type PlanningStepLog struct {
	StepIndex        int               `json:"step_index"`
	SpecialistName   string            `json:"specialist_name"`
	ActionPerformed  string            `json:"action_performed"`
	Duration         time.Duration     `json:"duration"`
	NodesAdded       int               `json:"nodes_added"`
	EdgesAdded       int               `json:"edges_added"`
	Status           string            `json:"status"`
	Metadata         map[string]string `json:"metadata"`
}

// DecompositionNode represents a node within the full hierarchical tree snapshot in PlanningTrace.
type DecompositionNode struct {
	NodeID      string              `json:"node_id"`
	ParentID    string              `json:"parent_id"`
	Title       string              `json:"title"`
	NodeType    string              `json:"node_type"` // "GOAL", "SUBGOAL", "PRIMITIVE_TASK"
	Children    []DecompositionNode `json:"children"`
	Attributes  map[string]string   `json:"attributes"`
}

// DependencyGraphSnapshot represents the complete directed graph inside PlanningTrace.
type DependencyGraphSnapshot struct {
	Nodes map[string]string `json:"nodes"` // NodeID -> Title
	Edges []DependencyEdge  `json:"edges"`
}

// SpecialistContribution holds the isolated output contributed by a single PlanningSpecialist.
type SpecialistContribution struct {
	SpecialistName string           `json:"specialist_name"`
	StepLog        *PlanningStepLog `json:"step_log,omitempty"`
	Subgoals       []Subgoal        `json:"subgoals"`
	Edges          []DependencyEdge `json:"edges"`
	Error          error            `json:"error,omitempty"`
}

// ============================================================================
// Core Output Schemas: Plan & PlanningTrace
// ============================================================================

// Plan is the lean operational payload consumed by Decision and Executive.
// It excludes heavy diagnostic trees and process statistics.
type Plan struct {
	PlanID                  string                   `json:"plan_id"`
	SchemaVersion           string                   `json:"schema_version"`
	CreatedAt               time.Time                `json:"created_at"`
	StrategySnapshotID      string                   `json:"strategy_snapshot_id"`
	PlanFingerprint         string                   `json:"plan_fingerprint"` // Hash over structural content ONLY
	SourceTier              string                   `json:"source_tier"`
	Domain                  string                   `json:"domain"`
	PlannerID               string                   `json:"planner_id,omitempty"`
	PlannerType             string                   `json:"planner_type,omitempty"`
	Goal                    string                   `json:"goal"`
	Subgoals                []Subgoal                `json:"subgoals"`
	Dependencies            []DependencyEdge         `json:"dependencies"`
	Preconditions           []string                 `json:"preconditions"`
	Postconditions          []string                 `json:"postconditions"`
	EstimatedCost           float64                  `json:"estimated_cost"`
	EstimatedDuration       time.Duration            `json:"estimated_duration"`
	RequiredResources       []ResourceRequirement    `json:"required_resources"`
	RollbackStrategies      []RollbackStrategy       `json:"rollback_strategies"`
	AlternativeBranches     []AlternativeBranch      `json:"alternative_branches"`
	ConfidenceProfile       ConfidenceProfile        `json:"confidence_profile"`
	Status                  PlanStatus               `json:"status"`
	InformationRequirements []InformationRequirement `json:"information_requirements"`
	ReplayMetadata          ReplayMetadata           `json:"replay_metadata"`
	TraceID                 string                   `json:"trace_id"`
}

// Validate executes strict structural checks and schema firewall enforcement on Plan.
func (p *Plan) Validate() error {
	if p.PlanID == "" {
		return errors.New("Plan missing PlanID")
	}
	if p.SchemaVersion != SchemaVersion2_0_0 && p.SchemaVersion != SchemaVersion2_0 {
		return fmt.Errorf("invalid Plan SchemaVersion: %s (must be %s or %s)", p.SchemaVersion, SchemaVersion2_0_0, SchemaVersion2_0)
	}
	if p.StrategySnapshotID == "" {
		return errors.New("Plan missing StrategySnapshotID")
	}
	if p.Goal == "" {
		return errors.New("Plan missing Goal")
	}
	if p.TraceID == "" {
		return errors.New("Plan missing linked TraceID")
	}
	if p.EstimatedCost < 0 {
		return fmt.Errorf("Plan has negative EstimatedCost: %f", p.EstimatedCost)
	}
	for i, sg := range p.Subgoals {
		if err := sg.Validate(); err != nil {
			return fmt.Errorf("subgoal[%d] invalid: %w", i, err)
		}
	}
	for i, dep := range p.Dependencies {
		if err := dep.Validate(); err != nil {
			return fmt.Errorf("dependency[%d] invalid: %w", i, err)
		}
	}
	for i, res := range p.RequiredResources {
		if err := res.Validate(); err != nil {
			return fmt.Errorf("resource[%d] invalid: %w", i, err)
		}
	}
	for i, rb := range p.RollbackStrategies {
		if err := rb.Validate(); err != nil {
			return fmt.Errorf("rollback strategy[%d] invalid: %w", i, err)
		}
	}
	for i, alt := range p.AlternativeBranches {
		if err := alt.Validate(); err != nil {
			return fmt.Errorf("alternative branch[%d] invalid: %w", i, err)
		}
	}
	for i, ir := range p.InformationRequirements {
		if err := ir.Validate(); err != nil {
			return fmt.Errorf("information requirement[%d] invalid: %w", i, err)
		}
	}
	if err := p.ConfidenceProfile.Validate(); err != nil {
		return fmt.Errorf("Plan ConfidenceProfile invalid: %w", err)
	}
	if err := p.ReplayMetadata.Validate(); err != nil {
		return fmt.Errorf("Plan ReplayMetadata invalid: %w", err)
	}
	return nil
}

// PlanningTrace is the deep diagnostic record consumed exclusively by Reflection and Learning.
// It contains complete decomposition graphs, process statistics, and factual discard reasons.
type PlanningTrace struct {
	TraceID               string                    `json:"trace_id"`
	PlanID                string                    `json:"plan_id"`
	SchemaVersion         string                    `json:"schema_version"`
	StrategySnapshotID    string                    `json:"strategy_snapshot_id"`
	PolicyFingerprint     string                    `json:"policy_fingerprint"`
	CapabilityFingerprint string                    `json:"capability_fingerprint"`
	SearchStrategyID      string                    `json:"search_strategy_id"`
	ReplayMetadata        ReplayMetadata            `json:"replay_metadata"`
	PlanningSteps         []PlanningStepLog         `json:"planning_steps"`
	SpecialistUsage       []PlanningSpecialistUsage `json:"specialist_usage"`
	DecompositionTree     DecompositionNode         `json:"decomposition_tree"`
	DependencyGraph       DependencyGraphSnapshot   `json:"dependency_graph"`
	Assumptions           []string                  `json:"assumptions"`
	RejectedBranches      []RejectedBranch          `json:"rejected_branches"`
	EstimatedComplexity   float64                   `json:"estimated_complexity"`
	ConfidenceProfile     ConfidenceProfile         `json:"confidence_profile"`
	QualityMetrics        QualityMetrics            `json:"quality_metrics"`
	ResourceAssumptions   []string                  `json:"resource_assumptions"`
	TerminationReason     PlanningTerminationReason `json:"termination_reason"`
	SearchStatistics      SearchStatistics          `json:"search_statistics"`
}

// Validate executes strict structural checks on PlanningTrace.
func (t *PlanningTrace) Validate() error {
	if t.TraceID == "" || t.PlanID == "" {
		return errors.New("PlanningTrace missing TraceID or PlanID")
	}
	if t.SchemaVersion != SchemaVersion2_0_0 && t.SchemaVersion != SchemaVersion2_0 {
		return fmt.Errorf("invalid PlanningTrace SchemaVersion: %s (must be %s or %s)", t.SchemaVersion, SchemaVersion2_0_0, SchemaVersion2_0)
	}
	if t.TerminationReason == "" {
		return errors.New("PlanningTrace missing TerminationReason")
	}
	if t.EstimatedComplexity < 0 {
		return fmt.Errorf("negative EstimatedComplexity: %f", t.EstimatedComplexity)
	}
	for i, rb := range t.RejectedBranches {
		if err := rb.Validate(); err != nil {
			return fmt.Errorf("rejected branch[%d] invalid: %w", i, err)
		}
	}
	for i, su := range t.SpecialistUsage {
		if err := su.Validate(); err != nil {
			return fmt.Errorf("specialist usage[%d] invalid: %w", i, err)
		}
	}
	if t.ReplayMetadata.StrategySnapshotID != "" || t.ReplayMetadata.ReplayFidelity != "" || t.ReplayMetadata.ReplaySeed != 0 {
		if err := t.ReplayMetadata.Validate(); err != nil {
			return fmt.Errorf("PlanningTrace ReplayMetadata invalid: %w", err)
		}
	}
	if err := t.ConfidenceProfile.Validate(); err != nil {
		return fmt.Errorf("PlanningTrace ConfidenceProfile invalid: %w", err)
	}
	if err := t.QualityMetrics.Validate(); err != nil {
		return fmt.Errorf("PlanningTrace QualityMetrics invalid: %w", err)
	}
	return nil
}

// ============================================================================
// Planning Request & Result Contracts
// ============================================================================

// PlanningRequest encapsulates the input arguments passed to PlanningService.
type PlanningRequest struct {
	RequestID          string                  `json:"request_id"`
	Goal               string                  `json:"goal"`
	ResolvedGoal       *reasoning.SemanticGoal `json:"resolved_goal,omitempty"`
	Domain             string                  `json:"domain"` // Open string tag (default: "General")
	ContextRef         string                  `json:"context_ref"` // URI to upstream SemanticFrame / ReasoningResult
	HardConstraints    []string                `json:"hard_constraints"`
	SoftConstraints    []string                `json:"soft_constraints"`
	TargetDepth        PlanningDepth           `json:"target_depth"`
	MaxExecutionBudget time.Duration           `json:"max_execution_budget"`
	MinConfidenceFloor float64                 `json:"min_confidence_floor"`
	PriorPlanRef       string                  `json:"prior_plan_ref"` // Optional URI for plan revision
	Metadata           map[string]string       `json:"metadata"`
}

// Validate verifies inputs on PlanningRequest.
func (r *PlanningRequest) Validate() error {
	if r.RequestID == "" {
		return errors.New("PlanningRequest missing RequestID")
	}
	if r.Goal == "" {
		return errors.New("PlanningRequest missing Goal")
	}
	if r.MinConfidenceFloor < 0.0 || r.MinConfidenceFloor > 1.0 {
		return fmt.Errorf("MinConfidenceFloor out of bounds [0.0, 1.0]: %f", r.MinConfidenceFloor)
	}
	return nil
}

// PlanningResult encapsulates the comprehensive output returned to Decision or Executive.
type PlanningResult struct {
	ResultID          string            `json:"result_id"`
	RequestID         string            `json:"request_id"`
	Plans             []*Plan           `json:"plans"`
	Traces            []*PlanningTrace  `json:"traces"`
	PrimaryPlanID            string               `json:"primary_plan_id"`
	ResultStatus             PlanningResultStatus `json:"result_status"`
	Status                   PlanStatus           `json:"status"`
	EscalationRecommendation EscalationAction `json:"escalation_recommendation"`
	ExecutedDepth     PlanningDepth     `json:"executed_depth"`
	TotalDuration     time.Duration     `json:"total_duration"`
}

// Validate ensures all returned plans and traces are structurally sound.
func (r *PlanningResult) Validate() error {
	if r.ResultID == "" || r.RequestID == "" {
		return errors.New("PlanningResult missing ResultID or RequestID")
	}
	if len(r.Plans) == 0 {
		return errors.New("PlanningResult must contain at least one Plan")
	}
	for i, p := range r.Plans {
		if p == nil {
			return fmt.Errorf("PlanningResult contains nil Plan at index %d", i)
		}
		if err := p.Validate(); err != nil {
			return fmt.Errorf("plan[%d] validation failed: %w", i, err)
		}
	}
	for i, t := range r.Traces {
		if t != nil {
			if err := t.Validate(); err != nil {
				return fmt.Errorf("trace[%d] validation failed: %w", i, err)
			}
		}
	}
	return nil
}

// ============================================================================
// Strategy Snapshot & Policy Profile Models
// ============================================================================

// PlanningCapabilities defines immutable structural boundaries and feature flags advertised by
// the current planning deployment (e.g., Mobile IDUN, Embedded IDUN, Cloud IDUN).
// It separates what mechanisms the engine supports from what policies decide to use.
type PlanningCapabilities struct {
	CapabilityFingerprint    string `json:"capability_fingerprint"`
	SupportsHTN              bool   `json:"supports_htn"`
	SupportsGOAP             bool   `json:"supports_goap"`
	SupportsTreeSearch       bool   `json:"supports_tree_search"`
	SupportsParallelSearch   bool   `json:"supports_parallel_search"`
	MaxParallelWorkers       uint16 `json:"max_parallel_workers"`
	MaxPlanningDepth         uint16 `json:"max_planning_depth"`
	MaxSupportedAlternatives uint16 `json:"max_supported_alternatives"`
	SupportsConstraintSolve  bool   `json:"supports_constraint_solve"`
	SupportsTemporalPlanning bool   `json:"supports_temporal_planning"`
	SupportsContingencies    bool   `json:"supports_contingencies"`
}

// ComputeCapabilityFingerprint returns the deterministic SHA-256 hex digest for the structural features in PlanningCapabilities.
func ComputeCapabilityFingerprint(c *PlanningCapabilities) string {
	if c == nil {
		return ""
	}
	data := fmt.Sprintf("%t|%t|%t|%t|%d|%d|%d|%t|%t|%t",
		c.SupportsHTN, c.SupportsGOAP, c.SupportsTreeSearch, c.SupportsParallelSearch,
		c.MaxParallelWorkers, c.MaxPlanningDepth, c.MaxSupportedAlternatives,
		c.SupportsConstraintSolve, c.SupportsTemporalPlanning, c.SupportsContingencies)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// Validate verifies structural soundness of the PlanningCapabilities boundaries.
func (c *PlanningCapabilities) Validate() error {
	if c.MaxPlanningDepth == 0 {
		return errors.New("PlanningCapabilities MaxPlanningDepth must be positive")
	}
	if c.MaxSupportedAlternatives == 0 {
		return errors.New("PlanningCapabilities MaxSupportedAlternatives must be positive")
	}
	if c.SupportsParallelSearch && c.MaxParallelWorkers == 0 {
		return errors.New("PlanningCapabilities MaxParallelWorkers must be positive when SupportsParallelSearch is true")
	}
	expectedFP := ComputeCapabilityFingerprint(c)
	if c.CapabilityFingerprint != "" && c.CapabilityFingerprint != expectedFP {
		return fmt.Errorf("PlanningCapabilities fingerprint mismatch: got %s, expected %s", c.CapabilityFingerprint, expectedFP)
	}
	return nil
}

// PlanningSearchStrategy defines immutable algorithmic limits and behavioral rules for planning execution.
// It decouples search parameters from specialist implementations, allowing Learning and Executive
// to evolve algorithms across decades without modifying the Planning public API or specialist interfaces.
type PlanningSearchStrategy struct {
	SearchID                string            `json:"search_id"`
	SearchVersion           string            `json:"search_version"`
	SearchFingerprint       string            `json:"search_fingerprint"`
	SearchType              string            `json:"search_type"` // e.g., "HTN", "GOAP", "BEAM_SEARCH", "MCTS"
	Description             string            `json:"description"`
	MaxDepth                uint32            `json:"max_depth"`
	MaxNodes                uint32            `json:"max_nodes"`
	BeamWidth               uint32            `json:"beam_width"`
	AllowParallelExpansion  bool              `json:"allow_parallel_expansion"`
	AllowBacktracking       bool              `json:"allow_backtracking"`
	AllowReplanning         bool              `json:"allow_replanning"`
	HeuristicID             string            `json:"heuristic_id"`
	HeuristicVersion        string            `json:"heuristic_version"`
	ExpansionPolicy         string            `json:"expansion_policy"` // e.g., "BEST_FIRST", "BREADTH_FIRST"
	PruningPolicy           string            `json:"pruning_policy"`   // e.g., "ALPHA_BETA", "PARETO_DOMINANCE"
	MaxConcurrentWorkers    uint32            `json:"max_concurrent_workers"`
	ConstraintSolverPolicy  string            `json:"constraint_solver_policy"`
	ConfigurationParameters map[string]string `json:"configuration_parameters"`
	SearchBudgetMs          uint32            `json:"search_budget_ms"`
}

// Validate verifies structural soundness of the PlanningSearchStrategy.
func (s *PlanningSearchStrategy) Validate() error {
	if s.SearchID == "" || s.SearchType == "" {
		return errors.New("PlanningSearchStrategy missing SearchID or SearchType")
	}
	if s.MaxDepth == 0 || s.MaxNodes == 0 {
		return fmt.Errorf("PlanningSearchStrategy %s must specify positive MaxDepth and MaxNodes", s.SearchID)
	}
	if s.BeamWidth == 0 {
		return fmt.Errorf("PlanningSearchStrategy %s must specify positive BeamWidth (minimum 1)", s.SearchID)
	}
	return nil
}

// PlanningPolicyProfile defines the versioned, immutable policy configuration
// published out-of-band by Learning and passively consumed by Planning.
type PlanningPolicyProfile struct {
	ProfileID            string                             `json:"profile_id"`
	ProfileVersion       string                             `json:"profile_version"`
	PolicyFingerprint    string                             `json:"policy_fingerprint"`
	PolicySource         string                             `json:"policy_source"`
	PlanningDepthLimits  map[string]int                     `json:"planning_depth_limits"`
	SpecialistWeights    map[string]float64                 `json:"specialist_weights"`
	DomainWeights        map[string]float64                 `json:"domain_weights"`
	EscalationThresholds map[string]float64                 `json:"escalation_thresholds"`
	SearchBudgets        map[string]int                     `json:"search_budgets"`
	MaxPlanningTime      time.Duration                      `json:"max_planning_time"`
	MaxPlanningNodes     int                                `json:"max_planning_nodes"`
	MaxAlternatives      int                                `json:"max_alternatives"`
	RiskPreferences      map[string]float64                 `json:"risk_preferences"`
	CalibrationWeight    float64                            `json:"calibration_weight"`
	MaxBeamWidth         int                                `json:"max_beam_width"`
	MaxBranchDepth       int                                `json:"max_branch_depth"`
	MaxInfoRequirements  int                                `json:"max_info_requirements"`
	SearchStrategies     map[string]*PlanningSearchStrategy `json:"search_strategies"`
	Capabilities         *PlanningCapabilities              `json:"capabilities"`
}

// Validate ensures weights and limits in PlanningPolicyProfile are valid.
func (profile *PlanningPolicyProfile) Validate() error {
	if profile.ProfileID == "" || profile.ProfileVersion == "" {
		return errors.New("PlanningPolicyProfile missing ProfileID or ProfileVersion")
	}
	if profile.MaxAlternatives <= 0 {
		return errors.New("PlanningPolicyProfile must have positive MaxAlternatives")
	}
	if profile.MaxPlanningNodes <= 0 {
		return errors.New("PlanningPolicyProfile must have positive MaxPlanningNodes")
	}
	if profile.Capabilities != nil {
		if err := profile.Capabilities.Validate(); err != nil {
			return fmt.Errorf("capabilities validation failed: %w", err)
		}
	}
	for key, strat := range profile.SearchStrategies {
		if strat != nil {
			if err := strat.Validate(); err != nil {
				return fmt.Errorf("search strategy [%s] validation failed: %w", key, err)
			}
		}
	}
	return nil
}

// PlanningStrategySnapshot encapsulates the lock-free, versioned strategy container
// providing read-only access to the active PlanningPolicyProfile.
type PlanningStrategySnapshot struct {
	SnapshotID      string
	Version         string
	Timestamp       time.Time
	activeProfile   atomic.Pointer[PlanningPolicyProfile]
}

// NewPlanningStrategySnapshot initializes a strategy snapshot with a profile pointer.
func NewPlanningStrategySnapshot(id, version string, profile *PlanningPolicyProfile) (*PlanningStrategySnapshot, error) {
	if id == "" || version == "" {
		return nil, errors.New("snapshot missing SnapshotID or Version")
	}
	if profile == nil {
		return nil, errors.New("cannot initialize snapshot with nil profile")
	}
	if err := profile.Validate(); err != nil {
		return nil, fmt.Errorf("invalid profile for snapshot: %w", err)
	}
	snap := &PlanningStrategySnapshot{
		SnapshotID: id,
		Version:    version,
		Timestamp:  time.Now().UTC(),
	}
	snap.activeProfile.Store(profile)
	return snap, nil
}

// ActiveProfile returns the currently active immutable policy profile via lock-free atomic load.
func (s *PlanningStrategySnapshot) ActiveProfile() *PlanningPolicyProfile {
	return s.activeProfile.Load()
}

// SwapProfile atomically replaces the active policy profile with a newly published profile.
func (s *PlanningStrategySnapshot) SwapProfile(newProfile *PlanningPolicyProfile) error {
	if newProfile == nil {
		return errors.New("cannot swap to nil profile")
	}
	if err := newProfile.Validate(); err != nil {
		return fmt.Errorf("new profile validation failed: %w", err)
	}
	s.activeProfile.Store(newProfile)
	return nil
}
