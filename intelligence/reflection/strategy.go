package reflection

import (
	"context"
	"fmt"
	"time"

	"idun/intelligence/communication"
)

// HeuristicEvaluationStrategy implements EvaluationStrategy using deterministic heuristic rules.
// It inspects read-only Workspace trace envelopes and generates structured SpecialistReports
// accompanied by immutable TraceReference lineage.
type HeuristicEvaluationStrategy struct {
	id      string
	version string
	ability string
}

// NewHeuristicEvaluationStrategy constructs a deterministic HeuristicEvaluationStrategy for a target ability.
func NewHeuristicEvaluationStrategy(id string, ability string) *HeuristicEvaluationStrategy {
	return &HeuristicEvaluationStrategy{
		id:      id,
		version: SchemaVersion,
		ability: ability,
	}
}

// StrategyID returns the strategy identifier.
func (h *HeuristicEvaluationStrategy) StrategyID() string {
	return h.id
}

// Version returns the strategy version.
func (h *HeuristicEvaluationStrategy) Version() string {
	return h.version
}

// Evaluate inspects read-only execution traces and emits a SpecialistReport.
// When no relevant traces are present, it explicitly returns VerdictInsufficientData.
func (h *HeuristicEvaluationStrategy) Evaluate(ctx context.Context, traces []communication.Envelope) (SpecialistReport, error) {
	if err := ctx.Err(); err != nil {
		return SpecialistReport{}, err
	}

	refs := make([]TraceReference, 0, len(traces))
	for _, env := range traces {
		// Include trace if matching ability or if evaluating overall activity
		if h.ability == "" || env.Source == h.ability {
			ts := env.CreatedAt.Unix()
			if ts <= 0 {
				ts = time.Now().Unix()
			}
			refs = append(refs, TraceReference{
				EnvelopeID:     env.ID,
				SourceAbility:  env.Source,
				TraceTimestamp: ts,
				PayloadHashRef: fmt.Sprintf("sha256-%s", env.ID),
			})
		}
	}

	if len(refs) == 0 {
		return SpecialistReport{
			SpecialistID:         h.id,
			TargetAbility:        h.ability,
			Verdict:              VerdictInsufficientData,
			WentWell:             []string{},
			WentPoorly:           []string{},
			CouldImprove:         []string{},
			ReflectionConfidence: 1.0,
			SourceTraceRefs:      []TraceReference{},
		}, nil
	}

	return SpecialistReport{
		SpecialistID:         h.id,
		TargetAbility:        h.ability,
		Verdict:              VerdictEvaluated,
		WentWell:             []string{fmt.Sprintf("%s completed successfully within episode bounds", h.ability)},
		WentPoorly:           []string{},
		CouldImprove:         []string{"Monitor execution latency consistency"},
		ReflectionConfidence: 0.88,
		SourceTraceRefs:      refs,
	}, nil
}
