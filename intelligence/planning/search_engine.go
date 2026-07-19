// Package planning provides foundational data structures and domain logic for IDUN V3's
// planning capabilities. This file implements the bounded Beam A* strategic search engine,
// including visited-state deduplication, heuristic evaluation, tripartite goal testing,
// and anytime budget preemption.
package planning

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

// ============================================================================
// Search Result & Configuration
// ============================================================================

// SearchResultStatus classifies the termination outcome of a strategic search episode.
type SearchResultStatus string

const (
	// StatusComplete indicates at least one valid terminal goal path was discovered within budget.
	StatusComplete SearchResultStatus = "STATUS_COMPLETE"
	// StatusPartialBudget indicates search was halted by budget or time preemption; partial candidates returned.
	StatusPartialBudget SearchResultStatus = "STATUS_PARTIAL_BUDGET"
	// StatusFailedNoPath indicates the search frontier exhausted without reaching any goal state.
	StatusFailedNoPath SearchResultStatus = "STATUS_FAILED_NO_PATH"
)

// SearchResult encapsulates the outcome of executing the Beam A* search engine.
type SearchResult struct {
	Status                    SearchResultStatus `json:"status"`
	GoalNodes                 []*SearchNode      `json:"goal_nodes,omitempty"`
	PartialNodes              []*SearchNode      `json:"partial_nodes,omitempty"`
	ExpandedCount             int                `json:"expanded_count"`
	PrunedBudgetCount         int                `json:"pruned_budget_count"`
	PrunedConstitutionalCount int                `json:"pruned_constitutional_count"`
	Duration                  time.Duration      `json:"duration"`
}

// BeamAStarConfig specifies the execution ceilings, weights, and pruning parameters for Beam A*.
type BeamAStarConfig struct {
	BeamWidth              int     `json:"beam_width"`               // Maximum OPEN nodes retained per level (default: 15)
	MaxDepth               int     `json:"max_depth"`                // Maximum trajectory depth (default: 50)
	Lambda                 float64 `json:"lambda"`                   // Heuristic weight λ in f(n) = g(n) + λ*h(n) + γ*risk (default: 1.0)
	Gamma                  float64 `json:"gamma"`                    // Risk penalty weight γ (default: 10.0)
	Beta                   float64 `json:"beta"`                     // Confidence degradation multiplier β (default: 2.0)
	MaxRiskCeiling         float64 `json:"max_risk_ceiling"`         // Absolute ceiling for accumulated path risk (default: 0.80)
	SemanticTransitionCost float64 `json:"semantic_transition_cost"` // Base cost prior per remaining desired condition (default: 1.0)
}

// DefaultBeamAStarConfig returns standard defaults aligned with the frozen Stage 5B specification.
func DefaultBeamAStarConfig() BeamAStarConfig {
	return BeamAStarConfig{
		BeamWidth:              15,
		MaxDepth:               50,
		Lambda:                 1.0,
		Gamma:                  10.0,
		Beta:                   2.0,
		MaxRiskCeiling:         0.80,
		SemanticTransitionCost: 1.0,
	}
}

// StrategicSearchEngine defines the core contract for bounded strategic exploration engines.
type StrategicSearchEngine interface {
	Search(ctx context.Context, req *PlanningRequest, rootState *SearchState, operators []*SearchEdge) (*SearchResult, error)
}

// ============================================================================
// Beam A* Search Engine Implementation
// ============================================================================

// BeamAStarEngine implements bounded Beam A* search with O(1) visited-state deduplication,
// tripartite goal testing, multi-signal heuristic evaluation, and anytime budget preemption.
type BeamAStarEngine struct {
	mu     sync.Mutex
	Config BeamAStarConfig
}

// NewBeamAStarEngine constructs a BeamAStarEngine governed by the provided configuration.
func NewBeamAStarEngine(cfg BeamAStarConfig) *BeamAStarEngine {
	if cfg.BeamWidth <= 0 {
		cfg.BeamWidth = 15
	}
	if cfg.MaxDepth <= 0 {
		cfg.MaxDepth = 50
	}
	return &BeamAStarEngine{
		Config: cfg,
	}
}

// Search executes bounded strategic exploration from rootState using the available abstract operators.
func (e *BeamAStarEngine) Search(
	ctx context.Context,
	req *PlanningRequest,
	rootState *SearchState,
	operators []*SearchEdge,
) (*SearchResult, error) {
	start := time.Now()

	if req == nil {
		return nil, errors.New("BeamAStarEngine: PlanningRequest cannot be nil")
	}
	if rootState == nil {
		return nil, errors.New("BeamAStarEngine: root SearchState cannot be nil")
	}
	if err := rootState.Validate(); err != nil {
		return nil, fmt.Errorf("BeamAStarEngine: root SearchState invalid: %w", err)
	}

	cfg := e.Config
	openQueue := NewOpenQueue()
	closedSet := make(map[string]float64) // StateID -> best scalar path cost g(n) seen

	// Initialize RemainingDesiredState if empty and req.ResolvedGoal is provided
	if len(rootState.RemainingDesiredState) == 0 && req.ResolvedGoal != nil && req.ResolvedGoal.DesiredState != nil {
		for k, v := range req.ResolvedGoal.DesiredState {
			if rootState.SimulatedWorldState[k] != v {
				rootState.RemainingDesiredState[k] = v
			}
		}
	}

	// Initialize hard constraints into rootState if present
	if rootState.ActiveConstraints == nil {
		rootState.ActiveConstraints = make(map[string]string)
	}
	if req.ResolvedGoal != nil && req.ResolvedGoal.Constraints != nil {
		for k, v := range req.ResolvedGoal.Constraints {
			rootState.ActiveConstraints[k] = v
		}
	}

	rootNode := NewSearchNode("node-root", nil, rootState.Clone(), nil)
	e.computeHeuristicAndScore(rootNode, req, cfg)

	result := &SearchResult{
		Status:    StatusFailedNoPath,
		GoalNodes: make([]*SearchNode, 0),
	}

	// Check if root state already satisfies the goal test immediately
	if e.isGoalState(rootNode, req, cfg) {
		rootNode.Status = NodeStatusTerminalGoal
		result.Status = StatusComplete
		result.GoalNodes = append(result.GoalNodes, rootNode)
		result.Duration = time.Since(start)
		return result, nil
	}

	rootNode.Status = NodeStatusOpen
	openQueue.Push(rootNode)

	// Determine execution budget and anytime preemption ceiling (85% of budget)
	budget := req.MaxExecutionBudget
	hasBudget := budget > 0

	for openQueue.Len() > 0 {
		// 1. Check context cancellation and anytime budget preemption ceiling
		if err := ctx.Err(); err != nil {
			return e.assemblePartialResult(openQueue, closedSet, result, start, StatusPartialBudget), nil
		}
		if hasBudget && time.Since(start) >= time.Duration(float64(budget)*0.85) {
			return e.assemblePartialResult(openQueue, closedSet, result, start, StatusPartialBudget), nil
		}

		// 2. Enforce level-based beam width K pruning across the OPEN queue
		if openQueue.Len() > cfg.BeamWidth {
			pruned, count := openQueue.PruneToBeam(cfg.BeamWidth)
			result.PrunedBudgetCount += count
			_ = pruned // Pruned nodes are discarded from further expansion
		}

		current := openQueue.Pop()
		if current == nil {
			break
		}

		// 3. O(1) visited-state deduplication (CLOSED set check)
		scalarG := e.computeScalarCost(current.CostProfile.AccumulatedCost)
		if bestG, exists := closedSet[current.State.StateID]; exists && bestG <= scalarG {
			continue // Skip expansion: we have already visited this state with equal or better cost
		}
		closedSet[current.State.StateID] = scalarG
		current.Status = NodeStatusClosed
		result.ExpandedCount++

		// Enforce maximum trajectory depth
		currentDepth := len(current.State.ExecutedTrajectory)
		if currentDepth >= cfg.MaxDepth {
			continue
		}

		// 4. Node expansion across feasible abstract strategic operators
		for _, op := range operators {
			if op == nil {
				continue
			}

			// Verify operator preconditions hold in SimulatedWorldState
			if !e.checkPreconditions(current.State.SimulatedWorldState, op.Preconditions) {
				continue
			}

			// Verify required assumptions hold in active Assumptions
			if !e.checkAssumptions(current.State.Assumptions, op.RequiredAssumptions) {
				continue
			}

			// Generate child state via deep cloning to guarantee zero shared references
			childState := current.State.Clone()

			// Apply postconditions and update RemainingDesiredState
			for pk, pv := range op.Postconditions {
				childState.SimulatedWorldState[pk] = pv
				if childState.RemainingDesiredState[pk] == pv {
					delete(childState.RemainingDesiredState, pk)
				}
			}

			// Accumulate costs and compound risk
			childState.AccumulatedCost = childState.AccumulatedCost.Add(op.EdgeCost)
			// Compound risk formula: r_new = r_old + (1 - r_old) * r_op
			childState.AccumulatedRisk = childState.AccumulatedRisk + (1.0-childState.AccumulatedRisk)*op.RiskDelta
			childState.SimulatedClock += op.EdgeCost.Time

			// Degrade confidence if operator carries risk
			if op.RiskDelta > 0 {
				childState.EpistemicConfidence = childState.EpistemicConfidence * (1.0 - op.RiskDelta*0.5)
			}

			// Assign deterministic child StateID
			childState.StateID = fmt.Sprintf("state-%s-op_%s-d%d", current.State.StateID, op.EdgeID, currentDepth+1)

			// Record step in trajectory
			step := SearchStep{
				StepID:          fmt.Sprintf("step-%d-%d", currentDepth+1, len(childState.ExecutedTrajectory)+1),
				StepIndex:       len(childState.ExecutedTrajectory),
				AppliedEdgeID:   op.EdgeID,
				OperatorName:    op.OperatorName,
				TransitionCost:  op.EdgeCost,
				RiskIncurred:    op.RiskDelta,
				SimulatedOffset: childState.SimulatedClock,
			}
			childState.ExecutedTrajectory = append(childState.ExecutedTrajectory, step)

			childNodeID := fmt.Sprintf("node-%s-%s", current.NodeID, op.EdgeID)
			childNode := NewSearchNode(childNodeID, current, childState, op.Clone())

			// 5. Check Constitutional invariants and Risk/Confidence viability
			if !e.isConstitutional(childState) {
				childNode.Status = NodeStatusPrunedConstitutional
				result.PrunedConstitutionalCount++
				continue
			}

			if !e.isViable(childState, req, cfg) {
				childNode.Status = NodeStatusPrunedBudget
				result.PrunedBudgetCount++
				continue
			}

			// 6. Compute heuristic and evaluation score f(n)
			e.computeHeuristicAndScore(childNode, req, cfg)

			// 7. Tripartite Goal Test
			if e.isGoalState(childNode, req, cfg) {
				childNode.Status = NodeStatusTerminalGoal
				result.GoalNodes = append(result.GoalNodes, childNode)
				result.Status = StatusComplete
			} else {
				childNode.Status = NodeStatusOpen
				openQueue.Push(childNode)
			}
		}
	}

	if len(result.GoalNodes) > 0 {
		result.Status = StatusComplete
	} else if result.Status == StatusFailedNoPath {
		// Collect top partial candidates if no goal reached
		e.assemblePartialResult(openQueue, closedSet, result, start, StatusFailedNoPath)
	}

	result.Duration = time.Since(start)
	return result, nil
}

// ReconstructPath extracts the chronological root-to-terminal path of SearchNodes.
func ReconstructPath(node *SearchNode) []*SearchNode {
	if node == nil {
		return []*SearchNode{}
	}
	path := make([]*SearchNode, 0)
	for curr := node; curr != nil; curr = curr.Parent {
		path = append(path, curr)
	}
	// Reverse slice in place
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}

// ============================================================================
// Tripartite Goal Test & Evaluation Methods
// ============================================================================

// isGoalState verifies that all desired states are achieved, constraints unbreached, and viability met.
func (e *BeamAStarEngine) isGoalState(node *SearchNode, req *PlanningRequest, cfg BeamAStarConfig) bool {
	if node == nil || node.State == nil {
		return false
	}
	s := node.State
	if len(s.RemainingDesiredState) > 0 {
		return false
	}
	return e.isConstitutional(s) && e.isViable(s, req, cfg)
}

func (e *BeamAStarEngine) isConstitutional(state *SearchState) bool {
	if state == nil {
		return false
	}
	for k, expectedV := range state.ActiveConstraints {
		if actualV, present := state.SimulatedWorldState[k]; present && actualV != expectedV {
			return false // Hard constraint violated by simulated world state
		}
	}
	return true
}

func (e *BeamAStarEngine) isViable(state *SearchState, req *PlanningRequest, cfg BeamAStarConfig) bool {
	if state == nil {
		return false
	}
	if state.AccumulatedRisk > cfg.MaxRiskCeiling {
		return false
	}
	if req != nil {
		if req.MinConfidenceFloor > 0 && state.EpistemicConfidence < req.MinConfidenceFloor {
			return false
		}
		if state.AccumulatedRisk > (1.0 - req.MinConfidenceFloor) {
			return false
		}
	}
	return true
}

func (e *BeamAStarEngine) checkPreconditions(worldState map[string]string, preconditions map[string]string) bool {
	if len(preconditions) == 0 {
		return true
	}
	if worldState == nil {
		return false
	}
	for k, requiredVal := range preconditions {
		if actualVal, ok := worldState[k]; !ok || actualVal != requiredVal {
			return false
		}
	}
	return true
}

func (e *BeamAStarEngine) checkAssumptions(assumptions map[string]string, requiredAssumptions map[string]string) bool {
	if len(requiredAssumptions) == 0 {
		return true
	}
	if assumptions == nil {
		return false
	}
	for k, requiredVal := range requiredAssumptions {
		if actualVal, ok := assumptions[k]; !ok || actualVal != requiredVal {
			return false
		}
	}
	return true
}

// computeHeuristicAndScore evaluates h(n) = max(h_semantic, h_resource) * (1 + β * (1 - confidence))
// and composite evaluation score f(n) = g(n) + λ * h(n) + γ * risk.
func (e *BeamAStarEngine) computeHeuristicAndScore(node *SearchNode, req *PlanningRequest, cfg BeamAStarConfig) {
	if node == nil || node.State == nil {
		return
	}
	s := node.State

	// h_semantic: remaining desired key count multiplied by base transition cost
	hSemantic := float64(len(s.RemainingDesiredState)) * cfg.SemanticTransitionCost

	// h_resource: estimated resources gap
	hResource := s.AccumulatedCost.Resources * 0.2

	hBase := math.Max(hSemantic, hResource)
	confidencePenalty := 1.0 + cfg.Beta*(1.0-s.EpistemicConfidence)
	hFinal := hBase * confidencePenalty

	// Set cost profiles
	node.CostProfile.AccumulatedCost = s.AccumulatedCost
	node.CostProfile.EstimatedRemainingCost = CostVector{
		Time:      time.Duration(hFinal * float64(time.Second)),
		Resources: hFinal,
	}

	scalarG := e.computeScalarCost(s.AccumulatedCost)
	fScore := scalarG + cfg.Lambda*hFinal + cfg.Gamma*s.AccumulatedRisk + s.AccumulatedCost.IrreversibilityPenalty
	node.CostProfile.EvaluationScore = fScore
}

func (e *BeamAStarEngine) computeScalarCost(c CostVector) float64 {
	return c.Resources + c.MonetaryCost + float64(c.Time.Milliseconds())*0.001
}

func (e *BeamAStarEngine) assemblePartialResult(
	queue *OpenQueue,
	closedSet map[string]float64,
	result *SearchResult,
	start time.Time,
	targetStatus SearchResultStatus,
) *SearchResult {
	result.Status = targetStatus
	result.Duration = time.Since(start)

	// Snapshot best nodes remaining in OPEN
	snapshot := queue.Snapshot()
	limit := 5
	if len(snapshot) < limit {
		limit = len(snapshot)
	}
	result.PartialNodes = snapshot[:limit]
	return result
}
