package reflection

import (
	"context"
	"fmt"
	"sync"

	"idun/intelligence/communication"
	"idun/intelligence/executive"
)

// DefaultSpecialistEvaluator wraps an EvaluationStrategy to evaluate a specific cognitive ability.
type DefaultSpecialistEvaluator struct {
	id       string
	ability  executive.CognitiveAbility
	strategy EvaluationStrategy
}

// NewSpecialistEvaluator creates a new DefaultSpecialistEvaluator.
func NewSpecialistEvaluator(id string, ability executive.CognitiveAbility, strategy EvaluationStrategy) *DefaultSpecialistEvaluator {
	return &DefaultSpecialistEvaluator{
		id:       id,
		ability:  ability,
		strategy: strategy,
	}
}

// ID returns the unique specialist ID.
func (d *DefaultSpecialistEvaluator) ID() string {
	return d.id
}

// TargetAbility identifies the cognitive ability evaluated by this specialist.
func (d *DefaultSpecialistEvaluator) TargetAbility() executive.CognitiveAbility {
	return d.ability
}

// EvaluateEpisode delegates execution to the underlying EvaluationStrategy and ensures identifiers match.
func (d *DefaultSpecialistEvaluator) EvaluateEpisode(ctx context.Context, traces []communication.Envelope) (SpecialistReport, error) {
	report, err := d.strategy.Evaluate(ctx, traces)
	if err != nil {
		return SpecialistReport{}, err
	}
	report.SpecialistID = d.id
	report.TargetAbility = string(d.ability)
	return report, nil
}

// SpecialistRegistry manages registration and safe concurrent execution of SpecialistEvaluator instances.
type SpecialistRegistry struct {
	mu          sync.RWMutex
	specialists map[string]SpecialistEvaluator
	order       []string
}

// NewSpecialistRegistry initializes an empty SpecialistRegistry.
func NewSpecialistRegistry() *SpecialistRegistry {
	return &SpecialistRegistry{
		specialists: make(map[string]SpecialistEvaluator),
		order:       make([]string, 0),
	}
}

// Register adds a specialist evaluator to the registry.
func (sr *SpecialistRegistry) Register(eval SpecialistEvaluator) error {
	if eval == nil || eval.ID() == "" {
		return ErrInvalidSpecialist
	}

	sr.mu.Lock()
	defer sr.mu.Unlock()

	if _, exists := sr.specialists[eval.ID()]; exists {
		return fmt.Errorf("%w: %s", ErrSpecialistAlreadyRegistered, eval.ID())
	}
	sr.specialists[eval.ID()] = eval
	sr.order = append(sr.order, eval.ID())
	return nil
}

// All returns a snapshot list of registered specialists in deterministic registration order.
func (sr *SpecialistRegistry) All() []SpecialistEvaluator {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	res := make([]SpecialistEvaluator, 0, len(sr.order))
	for _, id := range sr.order {
		if spec, ok := sr.specialists[id]; ok {
			res = append(res, spec)
		}
	}
	return res
}

// EvaluateAll executes all registered specialists over the provided read-only traces.
// It isolates errors and abstaining specialists so one specialist failure or abstention
// never prevents remaining specialists from completing.
func (sr *SpecialistRegistry) EvaluateAll(ctx context.Context, traces []communication.Envelope) ([]SpecialistReport, error) {
	specialists := sr.All()
	if len(specialists) == 0 {
		return []SpecialistReport{}, nil
	}

	reports := make([]SpecialistReport, len(specialists))
	var wg sync.WaitGroup

	for i, spec := range specialists {
		wg.Add(1)
		go func(idx int, s SpecialistEvaluator) {
			defer wg.Done()

			// Isolate panics
			defer func() {
				if r := recover(); r != nil {
					reports[idx] = SpecialistReport{
						SpecialistID:         s.ID(),
						TargetAbility:        string(s.TargetAbility()),
						Verdict:              VerdictAbstain,
						WentWell:             []string{},
						WentPoorly:           []string{},
						CouldImprove:         []string{},
						ReflectionConfidence: 0.0,
						SourceTraceRefs:      []TraceReference{},
					}
				}
			}()

			rep, err := s.EvaluateEpisode(ctx, traces)
			if err != nil {
				// Record error isolation as ABSTAIN
				reports[idx] = SpecialistReport{
					SpecialistID:         s.ID(),
					TargetAbility:        string(s.TargetAbility()),
					Verdict:              VerdictAbstain,
					WentWell:             []string{},
					WentPoorly:           []string{},
					CouldImprove:         []string{},
					ReflectionConfidence: 0.0,
					SourceTraceRefs:      []TraceReference{},
				}
				return
			}
			reports[idx] = rep
		}(i, spec)
	}

	wg.Wait()
	return reports, nil
}
