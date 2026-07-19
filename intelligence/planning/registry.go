package planning

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

// PlanningSpecialistRegistry defines the thread-safe contract for registering,
// sorting, and concurrently executing modular planning specialists.
//
// Architectural Invariant: The registry strictly coordinates lookup, timeout,
// and panic isolation. No planning algorithms or heuristics belong inside the registry.
type PlanningSpecialistRegistry interface {
	// Register adds a PlanningSpecialist to the active roster.
	Register(specialist PlanningSpecialist) error

	// GetSpecialistsForDomain returns all specialists supporting the given domain tag,
	// sorted descending by their weight in the provided PlanningPolicyProfile.
	GetSpecialistsForDomain(domain string, profile *PlanningPolicyProfile) []PlanningSpecialist

	// ExecuteSpecialists runs applicable specialists concurrently, isolating panics and timeouts,
	// returning isolated contributions per specialist and merged step logs.
	ExecuteSpecialists(
		ctx context.Context,
		req *PlanningRequest,
		graph *DependencyGraphSnapshot,
		profile *PlanningPolicyProfile,
		cache *ReflexivePlanningCache,
	) ([]SpecialistContribution, []PlanningStepLog, error)
}

// DefaultSpecialistRegistry implements thread-safe PlanningSpecialistRegistry.
type DefaultSpecialistRegistry struct {
	mu          sync.RWMutex
	specialists map[string]PlanningSpecialist
}

// NewSpecialistRegistry constructs a new clean specialist registry.
func NewSpecialistRegistry() *DefaultSpecialistRegistry {
	return &DefaultSpecialistRegistry{
		specialists: make(map[string]PlanningSpecialist),
	}
}

// Register safely stores a specialist by its canonical name.
func (r *DefaultSpecialistRegistry) Register(specialist PlanningSpecialist) error {
	if specialist == nil {
		return errors.New("cannot register nil PlanningSpecialist")
	}
	name := specialist.Name()
	if name == "" {
		return errors.New("cannot register PlanningSpecialist with empty name")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.specialists[name]; exists {
		return fmt.Errorf("specialist already registered: %s", name)
	}
	r.specialists[name] = specialist
	return nil
}

// GetAllSpecialists returns a snapshot of all registered specialists across all domains.
func (r *DefaultSpecialistRegistry) GetAllSpecialists() []PlanningSpecialist {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]PlanningSpecialist, 0, len(r.specialists))
	for _, spec := range r.specialists {
		list = append(list, spec)
	}
	return list
}

// GetSpecialistsForDomain filters and ranks specialists by domain support and policy profile weights.
func (r *DefaultSpecialistRegistry) GetSpecialistsForDomain(domain string, profile *PlanningPolicyProfile) []PlanningSpecialist {
	r.mu.RLock()
	var candidates []PlanningSpecialist
	for _, spec := range r.specialists {
		for _, d := range spec.SupportedDomains() {
			if d == domain || d == "General" || domain == "" {
				candidates = append(candidates, spec)
				break
			}
		}
	}
	r.mu.RUnlock()

	if profile == nil || profile.SpecialistWeights == nil {
		return candidates
	}

	// Sort descending by policy weight
	sort.Slice(candidates, func(i, j int) bool {
		wi := profile.SpecialistWeights[candidates[i].Name()]
		wj := profile.SpecialistWeights[candidates[j].Name()]
		return wi > wj
	})

	return candidates
}

// specialistResult holds the isolated output of a single concurrent specialist contribution.
type specialistResult struct {
	stepLog    *PlanningStepLog
	subgoals   []Subgoal
	edges      []DependencyEdge
	err        error
	specialist string
}

// ExecuteSpecialists runs applicable specialists concurrently with per-specialist panic recovery and timeout isolation.
func (r *DefaultSpecialistRegistry) ExecuteSpecialists(
	ctx context.Context,
	req *PlanningRequest,
	graph *DependencyGraphSnapshot,
	profile *PlanningPolicyProfile,
	cache *ReflexivePlanningCache,
) ([]SpecialistContribution, []PlanningStepLog, error) {
	if req == nil {
		return nil, nil, errors.New("cannot execute specialists with nil request")
	}

	specialists := r.GetSpecialistsForDomain(req.Domain, profile)
	if len(specialists) == 0 {
		return nil, nil, nil
	}

	var wg sync.WaitGroup
	resultCh := make(chan specialistResult, len(specialists))

	for i, spec := range specialists {
		wg.Add(1)
		go func(idx int, s PlanningSpecialist) {
			defer wg.Done()
			start := time.Now()

			// Panic Isolation
			defer func() {
				if rcv := recover(); rcv != nil {
					latencyUs := uint32(time.Since(start).Microseconds())
					if cache != nil {
						cache.RecordEvaluation(false, 0, 0, 1, 0, 1, latencyUs)
					}
					resultCh <- specialistResult{
						stepLog: &PlanningStepLog{
							StepIndex:       idx + 1,
							SpecialistName:  s.Name(),
							ActionPerformed: "PANIC_RECOVERED",
							Duration:        time.Since(start),
							Status:          "ERROR_PANIC",
							Metadata:        map[string]string{"panic": fmt.Sprintf("%v", rcv)},
						},
						specialist: s.Name(),
						err:        fmt.Errorf("specialist %s panicked: %v", s.Name(), rcv),
					}
				}
			}()

			// Check if outer context already cancelled
			if err := ctx.Err(); err != nil {
				resultCh <- specialistResult{
					stepLog: &PlanningStepLog{
						StepIndex:       idx + 1,
						SpecialistName:  s.Name(),
						ActionPerformed: "CANCELLED",
						Duration:        time.Since(start),
						Status:          "CANCELLED",
					},
					specialist: s.Name(),
					err:        err,
				}
				return
			}

			stepLog, subgoals, edges, err := s.Contribute(ctx, req, graph, profile)
			latencyUs := uint32(time.Since(start).Microseconds())
			if cache != nil {
				nodesAdded := len(subgoals)
				cache.RecordEvaluation(err == nil, nodesAdded, 0, 0, 0, 1, latencyUs)
			}

			if stepLog == nil {
				status := "DONE"
				if err != nil {
					status = "ERROR"
				}
				stepLog = &PlanningStepLog{
					StepIndex:       idx + 1,
					SpecialistName:  s.Name(),
					ActionPerformed: "CONTRIBUTE",
					Duration:        time.Since(start),
					NodesAdded:      len(subgoals),
					EdgesAdded:      len(edges),
					Status:          status,
				}
			} else {
				stepLog.StepIndex = idx + 1
				if stepLog.SpecialistName == "" {
					stepLog.SpecialistName = s.Name()
				}
				if stepLog.NodesAdded == 0 {
					stepLog.NodesAdded = len(subgoals)
				}
				if stepLog.EdgesAdded == 0 {
					stepLog.EdgesAdded = len(edges)
				}
				if stepLog.Status == "" {
					if err != nil {
						stepLog.Status = "ERROR"
					} else {
						stepLog.Status = "DONE"
					}
				}
			}

			resultCh <- specialistResult{
				stepLog:    stepLog,
				subgoals:   subgoals,
				edges:      edges,
				err:        err,
				specialist: s.Name(),
			}
		}(i, spec)
	}

	wg.Wait()
	close(resultCh)

	var allSteps []PlanningStepLog
	var allContribs []SpecialistContribution
	var firstErr error

	for res := range resultCh {
		if res.stepLog != nil {
			allSteps = append(allSteps, *res.stepLog)
		}
		if res.err != nil && firstErr == nil && !errors.Is(res.err, context.Canceled) {
			firstErr = res.err
		}
		allContribs = append(allContribs, SpecialistContribution{
			SpecialistName: res.specialist,
			StepLog:        res.stepLog,
			Subgoals:       res.subgoals,
			Edges:          res.edges,
			Error:          res.err,
		})
	}

	// Sort step logs by StepIndex for deterministic ordering
	sort.Slice(allSteps, func(i, j int) bool {
		return allSteps[i].StepIndex < allSteps[j].StepIndex
	})

	// Sort contributions deterministically by StepLog.StepIndex (or SpecialistName as fallback)
	sort.Slice(allContribs, func(i, j int) bool {
		idxI, idxJ := 999999, 999999
		if allContribs[i].StepLog != nil {
			idxI = allContribs[i].StepLog.StepIndex
		}
		if allContribs[j].StepLog != nil {
			idxJ = allContribs[j].StepLog.StepIndex
		}
		if idxI == idxJ {
			return allContribs[i].SpecialistName < allContribs[j].SpecialistName
		}
		return idxI < idxJ
	})

	return allContribs, allSteps, firstErr
}
