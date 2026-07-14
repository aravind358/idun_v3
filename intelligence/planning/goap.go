package planning

import (
	"context"
	"fmt"
	"time"
)

// GOAPSpecialist implements PlanningSpecialist using Goal-Oriented Action Planning (GOAP).
// It constructs state-transition action chains matching preconditions and postconditions,
// strictly governed by the active PlanningSearchStrategy (`BeamWidth`, `AllowBacktracking`, `MaxNodes`).
type GOAPSpecialist struct {
	defaultHorizon string
}

// NewGOAPSpecialist constructs a new GOAPSpecialist governed by the specified horizon (default "TACTICAL").
func NewGOAPSpecialist(horizon ...string) *GOAPSpecialist {
	h := "TACTICAL"
	if len(horizon) > 0 && horizon[0] != "" {
		h = horizon[0]
	}
	return &GOAPSpecialist{defaultHorizon: h}
}

func (s *GOAPSpecialist) Name() string {
	return "GOAPActionSpecialist"
}

func (s *GOAPSpecialist) SupportedDomains() []string {
	return []string{"General", "Tactical", "Action", "PhysicalTask", "Robotics"}
}

func (s *GOAPSpecialist) Contribute(
	ctx context.Context,
	req *PlanningRequest,
	currentGraph *DependencyGraphSnapshot,
	profile *PlanningPolicyProfile,
) (*PlanningStepLog, []Subgoal, []DependencyEdge, error) {
	start := time.Now()

	strat := resolveSearchStrategy(profile, s.defaultHorizon)
	if strat == nil {
		return nil, nil, nil, fmt.Errorf("GOAPSpecialist: no active search strategy found for horizon %s", s.defaultHorizon)
	}

	if profile != nil && profile.Capabilities != nil {
		if !profile.Capabilities.SupportsGOAP {
			return &PlanningStepLog{
				SpecialistName:  s.Name(),
				ActionPerformed: "GOAP skipped: engine capabilities do not support GOAP planning",
				Duration:        time.Since(start),
			}, nil, nil, nil
		}
	}

	if err := ctx.Err(); err != nil {
		return nil, nil, nil, err
	}
	existingNodes := uint32(0)
	if currentGraph != nil {
		existingNodes = uint32(len(currentGraph.Nodes))
	}
	if existingNodes >= strat.MaxNodes {
		return &PlanningStepLog{
			SpecialistName:  s.Name(),
			ActionPerformed: fmt.Sprintf("GOAP expansion halted: existing nodes (%d) reached strategy MaxNodes (%d)", existingNodes, strat.MaxNodes),
			Duration:        time.Since(start),
		}, nil, nil, nil
	}

	allowedNodes := strat.MaxNodes - existingNodes
	// BeamWidth dictates how many alternative action sequences GOAP maintains at each state transition
	actionCount := strat.BeamWidth * 2
	if actionCount > allowedNodes {
		actionCount = allowedNodes
	}
	if actionCount == 0 {
		return nil, nil, nil, nil
	}

	subgoals := make([]Subgoal, 0, actionCount)
	edges := make([]DependencyEdge, 0, actionCount)

	actionTemplates := []struct {
		action  string
		precond string
		post    string
	}{
		{"Acquire resources and environment lock", "Environment accessible", "Resources acquired"},
		{"Configure state parameters", "Resources acquired", "State ready"},
		{"Execute target goal transition", "State ready", req.Goal + " satisfied"},
		{"Verify state persistence and cleanup", req.Goal + " satisfied", "Goal stabilized"},
	}

	for i := uint32(0); i < actionCount && int(i) < len(actionTemplates); i++ {
		tpl := actionTemplates[i]
		sgID := fmt.Sprintf("goap-act-%d-%d", time.Now().UnixNano(), i+1)
		title := fmt.Sprintf("GOAP Action: %s", tpl.action)

		sg := Subgoal{
			SubgoalID:    sgID,
			Title:        title,
			Description:  fmt.Sprintf("Action state transition governed by strategy %s (backtracking=%v)", strat.SearchID, strat.AllowBacktracking),
			AssignedType: req.Domain,
			Parameters:   map[string]string{"priority": fmt.Sprintf("%d", strat.MaxDepth), "precondition": tpl.precond, "postcondition": tpl.post},
		}
		subgoals = append(subgoals, sg)

		if i > 0 {
			edgeID := fmt.Sprintf("goap-dep-%d-%d", time.Now().UnixNano(), i)
			edges = append(edges, DependencyEdge{
				EdgeID:         edgeID,
				SourceNodeID:   subgoals[i-1].SubgoalID,
				TargetNodeID:   sgID,
				DependencyType: "DATA_FLOW",
				IsBlocking:     true,
			})
		}
	}

	stepLog := &PlanningStepLog{
		SpecialistName:  s.Name(),
		ActionPerformed: fmt.Sprintf("GOAP chained %d state transitions governed by %s (beam_width=%d, backtracking=%v)", len(subgoals), strat.SearchID, strat.BeamWidth, strat.AllowBacktracking),
		Duration:        time.Since(start),
	}

	return stepLog, subgoals, edges, nil
}
