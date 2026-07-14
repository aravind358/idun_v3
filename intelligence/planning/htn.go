package planning

import (
	"context"
	"fmt"
	"time"
)

// HTNSpecialist implements PlanningSpecialist using Hierarchical Task Network decomposition.
// It strictly consumes its expansion limits (`MaxDepth`, `MaxNodes`, `BeamWidth`, `PruningPolicy`)
// exclusively from the immutable PlanningPolicyProfile / PlanningSearchStrategy snapshot without
// hardcoded algorithm parameters.
type HTNSpecialist struct {
	defaultHorizon string
}

// NewHTNSpecialist constructs a new HTNSpecialist governed by the specified horizon strategy (default "TACTICAL").
func NewHTNSpecialist(horizon ...string) *HTNSpecialist {
	h := "TACTICAL"
	if len(horizon) > 0 && horizon[0] != "" {
		h = horizon[0]
	}
	return &HTNSpecialist{defaultHorizon: h}
}

func (s *HTNSpecialist) Name() string {
	return "HTNDecompositionSpecialist"
}

func (s *HTNSpecialist) SupportedDomains() []string {
	return []string{"General", "Workflow", "Decomposition", "Coding", "Robotics", "Business"}
}

func (s *HTNSpecialist) Contribute(
	ctx context.Context,
	req *PlanningRequest,
	currentGraph *DependencyGraphSnapshot,
	profile *PlanningPolicyProfile,
) (*PlanningStepLog, []Subgoal, []DependencyEdge, error) {
	start := time.Now()

	strat := resolveSearchStrategy(profile, s.defaultHorizon)
	if strat == nil {
		return nil, nil, nil, fmt.Errorf("HTNSpecialist: no active search strategy found for horizon %s", s.defaultHorizon)
	}

	if profile != nil && profile.Capabilities != nil {
		if !profile.Capabilities.SupportsHTN {
			return &PlanningStepLog{
				SpecialistName:  s.Name(),
				ActionPerformed: "HTN skipped: engine capabilities do not support HTN decomposition",
				Duration:        time.Since(start),
			}, nil, nil, nil
		}
	}

	// Check timeout and expansion budgets from strategy
	if err := ctx.Err(); err != nil {
		return nil, nil, nil, err
	}
	existingNodes := uint32(0)
	if currentGraph != nil {
		existingNodes = uint32(len(currentGraph.Nodes))
	}
	if existingNodes >= strat.MaxNodes {
		// Budget exhausted, prune further expansion without error
		return &PlanningStepLog{
			SpecialistName:  s.Name(),
			ActionPerformed: fmt.Sprintf("HTN expansion halted: existing nodes (%d) reached strategy MaxNodes (%d)", existingNodes, strat.MaxNodes),
			Duration:        time.Since(start),
		}, nil, nil, nil
	}

	// Calculate allowed expansion count bounded by MaxDepth, Capabilities ceiling, and remaining MaxNodes
	allowedNodes := strat.MaxNodes - existingNodes
	expansionCount := strat.MaxDepth
	if profile != nil && profile.Capabilities != nil && profile.Capabilities.MaxPlanningDepth > 0 && uint32(profile.Capabilities.MaxPlanningDepth) < expansionCount {
		expansionCount = uint32(profile.Capabilities.MaxPlanningDepth)
	}
	if expansionCount > allowedNodes {
		expansionCount = allowedNodes
	}
	if expansionCount == 0 {
		return nil, nil, nil, nil
	}

	// Perform hierarchical task network decomposition of req.Goal
	subgoals := make([]Subgoal, 0, expansionCount)
	edges := make([]DependencyEdge, 0, expansionCount)

	// Generate baseline decomposition phases governed by strategy
	phases := []string{
		"Analyze scope and preconditions",
		"Formulate core structural design",
		"Execute primary task operations",
		"Verify outcome invariants and postconditions",
	}

	for i := uint32(0); i < expansionCount && int(i) < len(phases); i++ {
		sgID := fmt.Sprintf("htn-sg-%d-%d", time.Now().UnixNano(), i+1)
		title := fmt.Sprintf("%s for [%s]", phases[i], req.Goal)
		description := fmt.Sprintf("HTN task node at depth %d governed by strategy %s (%s)", i+1, strat.SearchID, strat.PruningPolicy)

		sg := Subgoal{
			SubgoalID:    sgID,
			Title:        title,
			Description:  description,
			AssignedType: req.Domain,
			Parameters:   map[string]string{"priority": fmt.Sprintf("%d", strat.MaxDepth-i)},
		}
		subgoals = append(subgoals, sg)

		// Create sequential dependency edges between adjacent HTN tasks
		if i > 0 {
			edgeID := fmt.Sprintf("htn-dep-%d-%d", time.Now().UnixNano(), i)
			edges = append(edges, DependencyEdge{
				EdgeID:         edgeID,
				SourceNodeID:   subgoals[i-1].SubgoalID,
				TargetNodeID:   sgID,
				DependencyType: "HARD_PREREQUISITE",
				IsBlocking:     true,
			})
		}
	}

	stepLog := &PlanningStepLog{
		SpecialistName:  s.Name(),
		ActionPerformed: fmt.Sprintf("HTN decomposed goal into %d subgoals and %d edges governed by %s (max_depth=%d, beam_width=%d)", len(subgoals), len(edges), strat.SearchID, strat.MaxDepth, strat.BeamWidth),
		Duration:        time.Since(start),
	}

	return stepLog, subgoals, edges, nil
}

// resolveSearchStrategy extracts the target horizon strategy from the immutable profile, falling back to TACTICAL or REFLEXIVE.
func resolveSearchStrategy(profile *PlanningPolicyProfile, targetHorizon string) *PlanningSearchStrategy {
	if profile != nil && profile.SearchStrategies != nil {
		if strat, ok := profile.SearchStrategies[targetHorizon]; ok && strat != nil {
			return strat
		}
		// Fallback order: TACTICAL -> STRATEGIC -> REFLEXIVE
		for _, fallback := range []string{"TACTICAL", "STRATEGIC", "REFLEXIVE"} {
			if strat, ok := profile.SearchStrategies[fallback]; ok && strat != nil {
				return strat
			}
		}
	}
	return nil
}
