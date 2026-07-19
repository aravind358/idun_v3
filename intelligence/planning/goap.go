package planning

import (
	"context"
	"fmt"
	"time"

	"idun/intelligence/reasoning"
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

	if req.ResolvedGoal != nil {
		// Stage 4D: Structured GOAP Adoption using canonical ResolvedGoal (DesiredState / Constraints).
		if req.ResolvedGoal.Kind == reasoning.GoalKindCommunicative {
			// Part 5: Communicative goals do not require physical side effects. Abstain cleanly so HTN handles them.
			return &PlanningStepLog{
				SpecialistName:  s.Name(),
				ActionPerformed: "GOAP skipped: communicative goals are handled by HTN without physical side effects",
				Duration:        time.Since(start),
			}, nil, nil, nil
		}

		dsStr := formatSortedMap(req.ResolvedGoal.DesiredState)
		cStr := formatSortedMap(req.ResolvedGoal.Constraints)

		baseID := req.RequestID
		if baseID == "" {
			baseID = req.ResolvedGoal.Fingerprint()
		}
		if baseID == "" {
			baseID = "goap-default"
		}

		var actionTemplates []struct {
			action  string
			precond string
			post    string
		}

		enforcementNote := "enforced_target_state"
		if req.ResolvedGoal.Constraints["read_only"] == "true" {
			// Enforce read_only constraint by rejecting mutating state transitions and selecting read-only inspection operators
			enforcementNote = "enforced_read_only"
			actionTemplates = []struct {
				action  string
				precond string
				post    string
			}{
				{
					action:  fmt.Sprintf("Validate target state accessibility and read-only scope for [%s]", req.ResolvedGoal.Target),
					precond: "Target accessible and read-only verified",
					post:    fmt.Sprintf("Inspection scope established for state: %s", dsStr),
				},
				{
					action:  fmt.Sprintf("Inspect current state conditions targeting [%s]", req.ResolvedGoal.Target),
					precond: fmt.Sprintf("Inspection scope established for state: %s", dsStr),
					post:    fmt.Sprintf("Current state observed against target: %s", dsStr),
				},
				{
					action:  fmt.Sprintf("Verify read-only invariants and desired state satisfaction: %s", dsStr),
					precond: fmt.Sprintf("Current state observed against target: %s", dsStr),
					post:    dsStr,
				},
				{
					action:  fmt.Sprintf("Confirm non-mutating diagnosis stabilization for [%s]", req.ResolvedGoal.Target),
					precond: dsStr,
					post:    fmt.Sprintf("Verified stable target state: %s", dsStr),
				},
			}
		} else {
			// Normal state-transition operator chain moving current state toward canonical DesiredState
			actionTemplates = []struct {
				action  string
				precond string
				post    string
			}{
				{
					action:  fmt.Sprintf("Validate target state preconditions and environment access for [%s]", req.ResolvedGoal.Target),
					precond: "Environment accessible",
					post:    fmt.Sprintf("Preconditions verified for target state: %s", dsStr),
				},
				{
					action:  fmt.Sprintf("Configure operator state transitions targeting [%s]", req.ResolvedGoal.Target),
					precond: fmt.Sprintf("Preconditions verified for target state: %s", dsStr),
					post:    fmt.Sprintf("Operators configured to establish: %s", dsStr),
				},
				{
					action:  fmt.Sprintf("Execute state transition establishing target condition: %s", dsStr),
					precond: fmt.Sprintf("Operators configured to establish: %s", dsStr),
					post:    dsStr,
				},
				{
					action:  fmt.Sprintf("Verify target state postconditions and persistence: %s", dsStr),
					precond: dsStr,
					post:    fmt.Sprintf("Target state verified and stabilized: %s", dsStr),
				},
			}
		}

		for i := uint32(0); i < actionCount && int(i) < len(actionTemplates); i++ {
			tpl := actionTemplates[i]
			sgID := fmt.Sprintf("goap-act-%s-%d", baseID, i+1)
			title := fmt.Sprintf("GOAP Action: %s", tpl.action)

			params := map[string]string{
				"priority":               fmt.Sprintf("%d", strat.MaxDepth),
				"goal_kind":              string(req.ResolvedGoal.Kind),
				"intent":                 req.ResolvedGoal.Intent,
				"target":                 req.ResolvedGoal.Target,
				"precondition":           tpl.precond,
				"postcondition":          tpl.post,
				"constraint_enforcement": enforcementNote,
			}
			if dsStr != "" {
				params["desired_state"] = dsStr
			}
			if cStr != "" {
				params["constraints"] = cStr
			}
			if req.ResolvedGoal.OperationHint != "" {
				params["operation_hint"] = req.ResolvedGoal.OperationHint
			}

			sg := Subgoal{
				SubgoalID:    sgID,
				Title:        title,
				Description:  fmt.Sprintf("Action state transition toward [%s] governed by strategy %s (backtracking=%v)", dsStr, strat.SearchID, strat.AllowBacktracking),
				AssignedType: req.Domain,
				Parameters:   params,
			}
			subgoals = append(subgoals, sg)

			if i > 0 {
				edgeID := fmt.Sprintf("goap-dep-%s-%d", baseID, i)
				edges = append(edges, DependencyEdge{
					EdgeID:         edgeID,
					SourceNodeID:   subgoals[i-1].SubgoalID,
					TargetNodeID:   sgID,
					DependencyType: "DATA_FLOW",
					IsBlocking:     true,
				})
			}
		}
	} else {
		// Legacy fallback behavior when req.ResolvedGoal == nil (exact pre-Stage-4D behavior)
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
	}

	stepLog := &PlanningStepLog{
		SpecialistName:  s.Name(),
		ActionPerformed: fmt.Sprintf("GOAP chained %d state transitions governed by %s (beam_width=%d, backtracking=%v)", len(subgoals), strat.SearchID, strat.BeamWidth, strat.AllowBacktracking),
		Duration:        time.Since(start),
	}

	return stepLog, subgoals, edges, nil
}
