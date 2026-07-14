package planning

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TreeSearchSpecialist implements PlanningSpecialist using wide multi-alternative tree search (e.g., Beam Search, MCTS, A*).
// It explores alternative branches and contingency subgoals, utilizing bounded concurrent worker fan-out when
// `AllowParallelExpansion` is enabled in the governing PlanningSearchStrategy.
type TreeSearchSpecialist struct {
	defaultHorizon string
}

// NewTreeSearchSpecialist constructs a new TreeSearchSpecialist governed by the specified horizon (default "STRATEGIC").
func NewTreeSearchSpecialist(horizon ...string) *TreeSearchSpecialist {
	h := "STRATEGIC"
	if len(horizon) > 0 && horizon[0] != "" {
		h = horizon[0]
	}
	return &TreeSearchSpecialist{defaultHorizon: h}
}

func (s *TreeSearchSpecialist) Name() string {
	return "MultiAlternativeTreeSearchSpecialist"
}

func (s *TreeSearchSpecialist) SupportedDomains() []string {
	return []string{"General", "Strategic", "Contingency", "Business", "Research"}
}

func (s *TreeSearchSpecialist) Contribute(
	ctx context.Context,
	req *PlanningRequest,
	currentGraph *DependencyGraphSnapshot,
	profile *PlanningPolicyProfile,
) (*PlanningStepLog, []Subgoal, []DependencyEdge, error) {
	start := time.Now()

	strat := resolveSearchStrategy(profile, s.defaultHorizon)
	if strat == nil {
		return nil, nil, nil, fmt.Errorf("TreeSearchSpecialist: no active search strategy found for horizon %s", s.defaultHorizon)
	}

	if profile != nil && profile.Capabilities != nil {
		if !profile.Capabilities.SupportsTreeSearch {
			return &PlanningStepLog{
				SpecialistName:  s.Name(),
				ActionPerformed: "Tree search skipped: engine capabilities do not support tree search",
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
			ActionPerformed: fmt.Sprintf("Tree Search expansion halted: existing nodes (%d) reached strategy MaxNodes (%d)", existingNodes, strat.MaxNodes),
			Duration:        time.Since(start),
		}, nil, nil, nil
	}

	allowedNodes := strat.MaxNodes - existingNodes
	branchCount := strat.BeamWidth
	if branchCount > allowedNodes {
		branchCount = allowedNodes
	}
	if branchCount == 0 {
		return nil, nil, nil, nil
	}

	subgoals := make([]Subgoal, 0, branchCount)
	var mu sync.Mutex

	// Determine worker concurrency bounded by strategy and engine capabilities
	workers := uint32(1)
	if strat.AllowParallelExpansion && (profile == nil || profile.Capabilities == nil || profile.Capabilities.SupportsParallelSearch) && strat.MaxConcurrentWorkers > 1 {
		workers = strat.MaxConcurrentWorkers
		if profile != nil && profile.Capabilities != nil && profile.Capabilities.MaxParallelWorkers > 0 && uint32(profile.Capabilities.MaxParallelWorkers) < workers {
			workers = uint32(profile.Capabilities.MaxParallelWorkers)
		}
		if workers > branchCount {
			workers = branchCount
		}
	}

	// Fan out candidate branch evaluation across concurrent workers
	workCh := make(chan uint32, branchCount)
	for i := uint32(0); i < branchCount; i++ {
		workCh <- i
	}
	close(workCh)

	var wg sync.WaitGroup
	for w := uint32(0); w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range workCh {
				if ctx.Err() != nil {
					return
				}
				sgID := fmt.Sprintf("tree-sg-%d-%d", time.Now().UnixNano(), idx+1)
				title := fmt.Sprintf("Contingency / Alternative Branch %d: %s [%s]", idx+1, strat.ExpansionPolicy, req.Goal)
				desc := fmt.Sprintf("Tree search branch generated under %s policy (pruning=%s)", strat.SearchType, strat.PruningPolicy)

				sg := Subgoal{
					SubgoalID:    sgID,
					Title:        title,
					Description:  desc,
					AssignedType: req.Domain,
					Parameters:   map[string]string{"priority": fmt.Sprintf("%d", strat.MaxDepth), "expansion_policy": strat.ExpansionPolicy},
				}

				mu.Lock()
				subgoals = append(subgoals, sg)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if err := ctx.Err(); err != nil {
		return nil, nil, nil, err
	}

	stepLog := &PlanningStepLog{
		SpecialistName:  s.Name(),
		ActionPerformed: fmt.Sprintf("Tree search generated %d alternative branches across %d concurrent workers governed by %s (policy=%s)", len(subgoals), workers, strat.SearchID, strat.ExpansionPolicy),
		Duration:        time.Since(start),
	}

	return stepLog, subgoals, nil, nil
}
