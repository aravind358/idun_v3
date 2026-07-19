// Package planning provides foundational data structures and domain logic for IDUN V3's
// planning capabilities, including tactical (HTN, GOAP) and strategic (TreeSearch) specialists.
package planning

import (
	"errors"
	"fmt"
	"time"
)

// ============================================================================
// Enumerations & Constants
// ============================================================================

// NodeStatus represents the current lifecycle and expansion state of a SearchNode
// within the strategic exploration graph.
type NodeStatus string

const (
	// NodeStatusUnexpanded indicates a node has been generated but not yet evaluated or expanded.
	NodeStatusUnexpanded NodeStatus = "UNEXPANDED"
	// NodeStatusOpen indicates a node is currently in the search frontier awaiting expansion.
	NodeStatusOpen NodeStatus = "OPEN"
	// NodeStatusClosed indicates a node and all of its feasible transitions have been expanded.
	NodeStatusClosed NodeStatus = "CLOSED"
	// NodeStatusTerminalGoal indicates a node successfully satisfies all desired states and constraints.
	NodeStatusTerminalGoal NodeStatus = "TERMINAL_GOAL"
	// NodeStatusPrunedBudget indicates a node was discarded due to exceeding depth, time, or resource ceilings.
	NodeStatusPrunedBudget NodeStatus = "PRUNED_BUDGET"
	// NodeStatusPrunedConstitutional indicates a node was discarded due to violating safety or constitutional rules.
	NodeStatusPrunedConstitutional NodeStatus = "PRUNED_CONSTITUTIONAL"
)

// IsValid checks whether the NodeStatus is a recognized enumeration value.
func (s NodeStatus) IsValid() bool {
	switch s {
	case NodeStatusUnexpanded, NodeStatusOpen, NodeStatusClosed,
		NodeStatusTerminalGoal, NodeStatusPrunedBudget, NodeStatusPrunedConstitutional:
		return true
	default:
		return false
	}
}

// String returns the canonical string representation of the NodeStatus.
func (s NodeStatus) String() string {
	return string(s)
}

// EdgeType classifies the strategic nature of a transition between two search states.
type EdgeType string

const (
	// EdgeTypeStrategicOperator represents a macro-level domain transition or action.
	EdgeTypeStrategicOperator EdgeType = "STRATEGIC_OPERATOR"
	// EdgeTypeAdversarialContingency represents a simulated environmental failure or exogenous event.
	EdgeTypeAdversarialContingency EdgeType = "ADVERSARIAL_CONTINGENCY"
)

// IsValid checks whether the EdgeType is a recognized enumeration value.
func (e EdgeType) IsValid() bool {
	switch e {
	case EdgeTypeStrategicOperator, EdgeTypeAdversarialContingency:
		return true
	default:
		return false
	}
}

// String returns the canonical string representation of the EdgeType.
func (e EdgeType) String() string {
	return string(e)
}

// Reversibility classifies the operational and risk impact of backtracking across a transition.
type Reversibility string

const (
	// ReversibilityTrivial indicates the transition can be undone with negligible cost or side effects.
	ReversibilityTrivial Reversibility = "REVERSIBLE_TRIVIAL"
	// ReversibilityHighCost indicates the transition can be reversed, but incurs substantial cost or latency.
	ReversibilityHighCost Reversibility = "REVERSIBLE_HIGH_COST"
	// ReversibilityCritical indicates the transition causes permanent, unrecoverable state changes.
	ReversibilityCritical Reversibility = "IRREVERSIBLE_CRITICAL"
)

// IsValid checks whether the Reversibility is a recognized enumeration value.
func (r Reversibility) IsValid() bool {
	switch r {
	case ReversibilityTrivial, ReversibilityHighCost, ReversibilityCritical:
		return true
	default:
		return false
	}
}

// String returns the canonical string representation of the Reversibility.
func (r Reversibility) String() string {
	return string(r)
}

// ============================================================================
// Cost Model & Evaluation Profiles
// ============================================================================

// CostVector models the multi-dimensional cost incurred by executing a strategic transition
// or accumulated along a trajectory from the search root.
type CostVector struct {
	Time                   time.Duration `json:"time"`                    // Simulated execution duration
	Resources              float64       `json:"resources"`               // Normalized physical/computational capacity consumed
	MonetaryCost           float64       `json:"monetary_cost"`           // Direct financial cost (e.g., USD / credits)
	Risk                   float64       `json:"risk"`                    // Compound probability [0.0, 1.0] of failure or adverse side effects
	IrreversibilityPenalty float64       `json:"irreversibility_penalty"` // Additive penalty units applied for irreversible actions
}

// Add returns a new CostVector equal to the component-wise sum of this vector and another.
func (c CostVector) Add(other CostVector) CostVector {
	return CostVector{
		Time:                   c.Time + other.Time,
		Resources:              c.Resources + other.Resources,
		MonetaryCost:           c.MonetaryCost + other.MonetaryCost,
		Risk:                   c.Risk + other.Risk,
		IrreversibilityPenalty: c.IrreversibilityPenalty + other.IrreversibilityPenalty,
	}
}

// IsZero checks whether all dimensions of the cost vector are zero.
func (c CostVector) IsZero() bool {
	return c.Time == 0 && c.Resources == 0 && c.MonetaryCost == 0 && c.Risk == 0 && c.IrreversibilityPenalty == 0
}

// Validate verifies structural boundaries of the CostVector.
func (c CostVector) Validate() error {
	if c.Time < 0 {
		return fmt.Errorf("cost vector has negative time duration: %v", c.Time)
	}
	if c.Resources < 0 {
		return fmt.Errorf("cost vector has negative resource quantity: %f", c.Resources)
	}
	if c.MonetaryCost < 0 {
		return fmt.Errorf("cost vector has negative monetary cost: %f", c.MonetaryCost)
	}
	if c.Risk < 0 || c.Risk > 1.0 {
		return fmt.Errorf("cost vector risk out of bounds [0.0, 1.0]: %f", c.Risk)
	}
	if c.IrreversibilityPenalty < 0 {
		return fmt.Errorf("cost vector has negative irreversibility penalty: %f", c.IrreversibilityPenalty)
	}
	return nil
}

// NodeCostProfile tracks the exact accumulated cost, heuristic remaining estimate, and scalar
// evaluation score associated with a SearchNode.
type NodeCostProfile struct {
	AccumulatedCost        CostVector `json:"accumulated_cost"`
	EstimatedRemainingCost CostVector `json:"estimated_remaining_cost"`
	EvaluationScore        float64    `json:"evaluation_score"` // Composite score f(n) used for priority ranking
}

// Validate checks structural integrity of the NodeCostProfile.
func (n *NodeCostProfile) Validate() error {
	if n == nil {
		return errors.New("node cost profile is nil")
	}
	if err := n.AccumulatedCost.Validate(); err != nil {
		return fmt.Errorf("invalid accumulated cost: %w", err)
	}
	if err := n.EstimatedRemainingCost.Validate(); err != nil {
		return fmt.Errorf("invalid estimated remaining cost: %w", err)
	}
	return nil
}

// ============================================================================
// Search Trajectory Steps & Transitions
// ============================================================================

// SearchStep records a single committed strategic transition inside a search state's trajectory.
type SearchStep struct {
	StepID          string            `json:"step_id"`
	StepIndex       int               `json:"step_index"`
	AppliedEdgeID   string            `json:"applied_edge_id"`
	OperatorName    string            `json:"operator_name"`
	TransitionCost  CostVector        `json:"transition_cost"`
	RiskIncurred    float64           `json:"risk_incurred"`
	SimulatedOffset time.Duration     `json:"simulated_offset"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

// Validate checks structural validity of the SearchStep.
func (s *SearchStep) Validate() error {
	if s == nil {
		return errors.New("search step is nil")
	}
	if s.StepID == "" {
		return errors.New("search step missing StepID")
	}
	if s.AppliedEdgeID == "" {
		return errors.New("search step missing AppliedEdgeID")
	}
	if s.OperatorName == "" {
		return errors.New("search step missing OperatorName")
	}
	if s.StepIndex < 0 {
		return fmt.Errorf("search step %s has negative index: %d", s.StepID, s.StepIndex)
	}
	if s.RiskIncurred < 0 || s.RiskIncurred > 1.0 {
		return fmt.Errorf("search step %s has out of bounds risk: %f", s.StepID, s.RiskIncurred)
	}
	if err := s.TransitionCost.Validate(); err != nil {
		return fmt.Errorf("search step %s transition cost invalid: %w", s.StepID, err)
	}
	return nil
}

// Clone returns a deep copy of the SearchStep.
func (s *SearchStep) Clone() SearchStep {
	if s == nil {
		return SearchStep{}
	}
	out := *s
	if s.Metadata != nil {
		metaCopy := make(map[string]string, len(s.Metadata))
		for k, v := range s.Metadata {
			metaCopy[k] = v
		}
		out.Metadata = metaCopy
	}
	return out
}

// ============================================================================
// Search State & Edge Definitions
// ============================================================================

// SearchEdge models a polymorphic strategic transition connecting two search states.
type SearchEdge struct {
	EdgeID              string            `json:"edge_id"`
	EdgeType            EdgeType          `json:"edge_type"`
	OperatorName        string            `json:"operator_name"`
	Preconditions       map[string]string `json:"preconditions"`
	RequiredAssumptions map[string]string `json:"required_assumptions"`
	Postconditions      map[string]string `json:"postconditions"`
	EdgeCost            CostVector        `json:"edge_cost"`
	RiskDelta           float64           `json:"risk_delta"`
	Reversibility       Reversibility     `json:"reversibility"`
}

// NewSearchEdge constructs and initializes a new SearchEdge with empty maps ready for assignment.
func NewSearchEdge(edgeID string, edgeType EdgeType, operatorName string) *SearchEdge {
	return &SearchEdge{
		EdgeID:              edgeID,
		EdgeType:            edgeType,
		OperatorName:        operatorName,
		Preconditions:       make(map[string]string),
		RequiredAssumptions: make(map[string]string),
		Postconditions:      make(map[string]string),
		Reversibility:       ReversibilityTrivial,
	}
}

// Validate checks required fields, enum correctness, and boundary constraints on SearchEdge.
func (e *SearchEdge) Validate() error {
	if e == nil {
		return errors.New("search edge is nil")
	}
	if e.EdgeID == "" {
		return errors.New("search edge missing EdgeID")
	}
	if !e.EdgeType.IsValid() {
		return fmt.Errorf("search edge %s has invalid EdgeType: %q", e.EdgeID, e.EdgeType)
	}
	if e.OperatorName == "" {
		return fmt.Errorf("search edge %s missing OperatorName", e.EdgeID)
	}
	if !e.Reversibility.IsValid() {
		return fmt.Errorf("search edge %s has invalid Reversibility: %q", e.EdgeID, e.Reversibility)
	}
	if e.RiskDelta < 0 || e.RiskDelta > 1.0 {
		return fmt.Errorf("search edge %s has out of bounds RiskDelta: %f", e.EdgeID, e.RiskDelta)
	}
	if err := e.EdgeCost.Validate(); err != nil {
		return fmt.Errorf("search edge %s has invalid EdgeCost: %w", e.EdgeID, err)
	}
	if e.Preconditions == nil || e.RequiredAssumptions == nil || e.Postconditions == nil {
		return fmt.Errorf("search edge %s must not have nil map fields (use empty map instead)", e.EdgeID)
	}
	return nil
}

// Clone creates an exact deep copy of the SearchEdge and its underlying maps.
func (e *SearchEdge) Clone() *SearchEdge {
	if e == nil {
		return nil
	}
	out := *e
	out.Preconditions = cloneStringMap(e.Preconditions)
	out.RequiredAssumptions = cloneStringMap(e.RequiredAssumptions)
	out.Postconditions = cloneStringMap(e.Postconditions)
	return &out
}

// SearchState captures the complete snapshot of simulated reality, active assumptions, constraints,
// and accumulated metrics at a specific point in the strategic search graph.
type SearchState struct {
	StateID               string                `json:"state_id"`
	SimulatedWorldState   map[string]string     `json:"simulated_world_state"`
	RemainingDesiredState map[string]string     `json:"remaining_desired_state"`
	ActiveConstraints     map[string]string     `json:"active_constraints"`
	Assumptions           map[string]string     `json:"assumptions"`
	ExecutedTrajectory    []SearchStep          `json:"executed_trajectory"`
	AccumulatedCost       CostVector            `json:"accumulated_cost"`
	AccumulatedRisk       float64               `json:"accumulated_risk"`
	EpistemicConfidence   float64               `json:"epistemic_confidence"`
	SimulatedClock        time.Duration         `json:"simulated_clock"`
	ResourceReservations  []ResourceRequirement `json:"resource_reservations"`
}

// NewSearchState constructs a clean SearchState with non-nil maps and slices initialized.
func NewSearchState(stateID string) *SearchState {
	return &SearchState{
		StateID:               stateID,
		SimulatedWorldState:   make(map[string]string),
		RemainingDesiredState: make(map[string]string),
		ActiveConstraints:     make(map[string]string),
		Assumptions:           make(map[string]string),
		ExecutedTrajectory:    make([]SearchStep, 0),
		ResourceReservations:  make([]ResourceRequirement, 0),
		EpistemicConfidence:   1.0,
	}
}

// Validate checks structural integrity and numerical bounds of the SearchState.
func (s *SearchState) Validate() error {
	if s == nil {
		return errors.New("search state is nil")
	}
	if s.StateID == "" {
		return errors.New("search state missing StateID")
	}
	if s.SimulatedWorldState == nil || s.RemainingDesiredState == nil || s.ActiveConstraints == nil || s.Assumptions == nil {
		return fmt.Errorf("search state %s has nil map fields", s.StateID)
	}
	if s.ExecutedTrajectory == nil || s.ResourceReservations == nil {
		return fmt.Errorf("search state %s has nil slice fields", s.StateID)
	}
	if s.AccumulatedRisk < 0 || s.AccumulatedRisk > 1.0 {
		return fmt.Errorf("search state %s has out of bounds AccumulatedRisk: %f", s.StateID, s.AccumulatedRisk)
	}
	if s.EpistemicConfidence < 0 || s.EpistemicConfidence > 1.0 {
		return fmt.Errorf("search state %s has out of bounds EpistemicConfidence: %f", s.StateID, s.EpistemicConfidence)
	}
	if s.SimulatedClock < 0 {
		return fmt.Errorf("search state %s has negative SimulatedClock: %v", s.StateID, s.SimulatedClock)
	}
	if err := s.AccumulatedCost.Validate(); err != nil {
		return fmt.Errorf("search state %s has invalid AccumulatedCost: %w", s.StateID, err)
	}
	for i, step := range s.ExecutedTrajectory {
		if err := step.Validate(); err != nil {
			return fmt.Errorf("search state %s trajectory step[%d] invalid: %w", s.StateID, i, err)
		}
	}
	for i, res := range s.ResourceReservations {
		if err := res.Validate(); err != nil {
			return fmt.Errorf("search state %s resource reservation[%d] invalid: %w", s.StateID, i, err)
		}
	}
	return nil
}

// Clone creates a completely independent deep copy of the SearchState, ensuring zero shared
// references across maps or slices when expanding new branches during search.
func (s *SearchState) Clone() *SearchState {
	if s == nil {
		return nil
	}
	out := *s
	out.SimulatedWorldState = cloneStringMap(s.SimulatedWorldState)
	out.RemainingDesiredState = cloneStringMap(s.RemainingDesiredState)
	out.ActiveConstraints = cloneStringMap(s.ActiveConstraints)
	out.Assumptions = cloneStringMap(s.Assumptions)

	if s.ExecutedTrajectory != nil {
		stepsCopy := make([]SearchStep, len(s.ExecutedTrajectory))
		for i, step := range s.ExecutedTrajectory {
			stepsCopy[i] = step.Clone()
		}
		out.ExecutedTrajectory = stepsCopy
	} else {
		out.ExecutedTrajectory = make([]SearchStep, 0)
	}

	if s.ResourceReservations != nil {
		resCopy := make([]ResourceRequirement, len(s.ResourceReservations))
		copy(resCopy, s.ResourceReservations)
		out.ResourceReservations = resCopy
	} else {
		out.ResourceReservations = make([]ResourceRequirement, 0)
	}

	return &out
}

// ============================================================================
// Search Graph Topology (`SearchNode`)
// ============================================================================

// SearchNode represents a single vertex within the strategic exploration tree, binding
// a SearchState to its structural graph hierarchy and evaluation metrics.
type SearchNode struct {
	NodeID          string            `json:"node_id"`
	Parent          *SearchNode       `json:"-"` // Omitted from JSON serialization to prevent circular references
	State           *SearchState      `json:"state"`
	IncomingEdge    *SearchEdge       `json:"incoming_edge,omitempty"`
	CostProfile     NodeCostProfile   `json:"cost_profile"`
	PlannerMetadata map[string]string `json:"planner_metadata,omitempty"`
	Status          NodeStatus        `json:"status"`
}

// NewSearchNode constructs a SearchNode wrapped around the given SearchState and incoming edge.
func NewSearchNode(nodeID string, parent *SearchNode, state *SearchState, edge *SearchEdge) *SearchNode {
	return &SearchNode{
		NodeID:          nodeID,
		Parent:          parent,
		State:           state,
		IncomingEdge:    edge,
		CostProfile:     NodeCostProfile{},
		PlannerMetadata: make(map[string]string),
		Status:          NodeStatusUnexpanded,
	}
}

// Validate checks required fields, enum correctness, and nested state integrity of the SearchNode.
func (n *SearchNode) Validate() error {
	if n == nil {
		return errors.New("search node is nil")
	}
	if n.NodeID == "" {
		return errors.New("search node missing NodeID")
	}
	if !n.Status.IsValid() {
		return fmt.Errorf("search node %s has invalid Status: %q", n.NodeID, n.Status)
	}
	if n.State == nil {
		return fmt.Errorf("search node %s missing State", n.NodeID)
	}
	if err := n.State.Validate(); err != nil {
		return fmt.Errorf("search node %s state invalid: %w", n.NodeID, err)
	}
	if n.IncomingEdge != nil {
		if err := n.IncomingEdge.Validate(); err != nil {
			return fmt.Errorf("search node %s incoming edge invalid: %w", n.NodeID, err)
		}
	}
	if err := n.CostProfile.Validate(); err != nil {
		return fmt.Errorf("search node %s cost profile invalid: %w", n.NodeID, err)
	}
	return nil
}

// Clone returns a deep copy of the SearchNode, recursively cloning its State, IncomingEdge,
// and PlannerMetadata map while retaining the read-only Parent back-pointer.
func (n *SearchNode) Clone() *SearchNode {
	if n == nil {
		return nil
	}
	out := *n
	out.State = n.State.Clone()
	out.IncomingEdge = n.IncomingEdge.Clone()
	out.PlannerMetadata = cloneStringMap(n.PlannerMetadata)
	return &out
}

// ============================================================================
// Internal Helpers
// ============================================================================

func cloneStringMap(in map[string]string) map[string]string {
	if in == nil {
		return make(map[string]string)
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
