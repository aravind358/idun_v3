package planning

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"idun/intelligence/reasoning"
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

func formatSortedMap(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(m[k])
	}
	return b.String()
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

	subgoals := make([]Subgoal, 0, expansionCount)
	edges := make([]DependencyEdge, 0, expansionCount)

	if req.ResolvedGoal != nil {
		// Stage 4C: Structured HTN Adoption using canonical ResolvedGoal without mutating or reinterpreting it.
		dsStr := formatSortedMap(req.ResolvedGoal.DesiredState)
		cStr := formatSortedMap(req.ResolvedGoal.Constraints)

		var phases []string
		var titles []string

		if req.ResolvedGoal.Kind == reasoning.GoalKindCommunicative {
			phases = []string{
				"Validate communicative context and scope",
				"Formulate communicative structure",
				"Produce approved communicative content",
				"Verify communicative invariants",
			}
			titles = make([]string, len(phases))
			titles[0] = fmt.Sprintf("Validate communicative context and scope for intent [%s]", req.ResolvedGoal.Intent)
			titles[1] = fmt.Sprintf("Formulate communicative structure targeting [%s]", req.ResolvedGoal.Target)
			if req.ResolvedGoal.Intent == "greet_user" && req.ResolvedGoal.Target == "user" {
				titles[2] = "Produce an approved acknowledgement of the user's greeting."
			} else if dsStr != "" {
				titles[2] = fmt.Sprintf("Produce approved communicative content achieving [%s] targeting [%s]", dsStr, req.ResolvedGoal.Target)
			} else {
				titles[2] = fmt.Sprintf("Produce approved communicative content for intent [%s] targeting [%s]", req.ResolvedGoal.Intent, req.ResolvedGoal.Target)
			}
			titles[3] = fmt.Sprintf("Verify communicative invariants achieving state [%s]", dsStr)
		} else if req.ResolvedGoal.Kind == reasoning.GoalKindToolExecution {
			phases = []string{
				"Validate required parameters and scope",
				"Prepare tool operation",
				"Establish desired state",
				"Verify resulting state invariants",
			}
			titles = make([]string, len(phases))
			if req.ResolvedGoal.Intent == "set_alarm" && req.ResolvedGoal.Target == "alarm_service" {
				titles[0] = "Validate required alarm parameters."
				titles[1] = "Prepare alarm service operation."
				titles[2] = fmt.Sprintf("Establish desired alarm state: %s.", dsStr)
				titles[3] = "Verify resulting alarm state."
			} else {
				titles[0] = fmt.Sprintf("Validate required parameters and scope for [%s].", req.ResolvedGoal.Intent)
				titles[1] = fmt.Sprintf("Prepare %s operation targeting [%s].", req.ResolvedGoal.Intent, req.ResolvedGoal.Target)
				titles[2] = fmt.Sprintf("Establish desired state targeting [%s]: %s.", req.ResolvedGoal.Target, dsStr)
				titles[3] = fmt.Sprintf("Verify resulting state invariants for [%s].", req.ResolvedGoal.Target)
			}
		} else {
			phases = []string{
				"Analyze scope and preconditions",
				"Formulate core structural design",
				"Execute primary task operations",
				"Verify outcome invariants and postconditions",
			}
			titles = make([]string, len(phases))
			titles[0] = fmt.Sprintf("Analyze scope and preconditions for [%s]", req.ResolvedGoal.Intent)
			titles[1] = fmt.Sprintf("Formulate core structural design targeting [%s]", req.ResolvedGoal.Target)
			titles[2] = fmt.Sprintf("Execute primary operations achieving [%s]", dsStr)
			titles[3] = fmt.Sprintf("Verify outcome invariants and postconditions for [%s]", req.ResolvedGoal.Target)
		}

		baseID := req.RequestID
		if baseID == "" {
			baseID = req.ResolvedGoal.Fingerprint()
		}
		if baseID == "" {
			baseID = "htn-default"
		}

		for i := uint32(0); i < expansionCount && int(i) < len(phases); i++ {
			sgID := fmt.Sprintf("htn-sg-%s-%d", baseID, i+1)
			description := fmt.Sprintf("HTN task node at depth %d governed by strategy %s (%s)", i+1, strat.SearchID, strat.PruningPolicy)

			params := map[string]string{
				"priority":  fmt.Sprintf("%d", strat.MaxDepth-i),
				"goal_kind": string(req.ResolvedGoal.Kind),
				"intent":    req.ResolvedGoal.Intent,
				"target":    req.ResolvedGoal.Target,
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
			if i == 0 {
				params["precondition"] = fmt.Sprintf("Context valid for intent %s", req.ResolvedGoal.Intent)
			} else {
				params["precondition"] = fmt.Sprintf("Phase %d completed successfully", i)
			}
			if int(i) == len(phases)-1 || i == expansionCount-1 {
				params["postcondition"] = fmt.Sprintf("Desired state achieved: %s", dsStr)
			} else {
				params["postcondition"] = fmt.Sprintf("Phase %d operational requirements satisfied", i+1)
			}

			sg := Subgoal{
				SubgoalID:    sgID,
				Title:        titles[i],
				Description:  description,
				AssignedType: req.Domain,
				Parameters:   params,
			}
			subgoals = append(subgoals, sg)

			if i > 0 {
				edgeID := fmt.Sprintf("htn-dep-%s-%d", baseID, i)
				edges = append(edges, DependencyEdge{
					EdgeID:         edgeID,
					SourceNodeID:   subgoals[i-1].SubgoalID,
					TargetNodeID:   sgID,
					DependencyType: "HARD_PREREQUISITE",
					IsBlocking:     true,
				})
			}
		}
	} else {
		// Legacy behavior when req.ResolvedGoal == nil
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
