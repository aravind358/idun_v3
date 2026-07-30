package v3

import (
	"context"
	"idun/boundary/perception"
)

// UnderstandingService defines the primary entry point for the Understanding layer.
type UnderstandingService interface {
	Analyze(ctx context.Context, env *perception.PerceptionEnvelope) (*SemanticInterpretation, error)
}

// Orchestrator coordinates the execution of specialists, evaluates their output,
// and synthesizes the final SemanticInterpretation.
type Orchestrator struct {
	grammar      Specialist
	neural       Specialist
	deliberative Specialist
}

// NewOrchestrator creates a new Orchestrator with the given specialists.
func NewOrchestrator(grammar, neural, deliberative Specialist) *Orchestrator {
	return &Orchestrator{
		grammar:      grammar,
		neural:       neural,
		deliberative: deliberative,
	}
}

// Analyze processes a PerceptionEnvelope through the cascade of specialists.
func (o *Orchestrator) Analyze(ctx context.Context, env *perception.PerceptionEnvelope) (*SemanticInterpretation, error) {
	// 1. Cascade execution
	hyps := o.cascadeAnalyze(ctx, env)

	// 2. Evaluate hypotheses
	primary, ambiguitySet, status := EvaluateHypotheses(hyps)

	// 3. Synthesize final interpretation
	return Synthesize(env, primary, ambiguitySet, status)
}

func (o *Orchestrator) cascadeAnalyze(ctx context.Context, env *perception.PerceptionEnvelope) []Hypothesis {
	// Attempt Grammar
	if o.grammar != nil {
		if hyps, err := o.grammar.Analyze(ctx, env); err == nil && len(hyps) > 0 {
			// If Grammar is highly confident, we can return early, else we might still cascade.
			// The spec says if confident, return. We assume any valid hypotheses from grammar are returned.
			return hyps
		}
	}

	// Attempt Neural if Grammar failed or wasn't present
	if o.neural != nil {
		if hyps, err := o.neural.Analyze(ctx, env); err == nil && len(hyps) > 0 {
			return hyps
		}
	}

	// Attempt Deliberative if Neural failed
	if o.deliberative != nil {
		if hyps, err := o.deliberative.Analyze(ctx, env); err == nil && len(hyps) > 0 {
			return hyps
		}
	}

	// Return empty if all failed
	return []Hypothesis{}
}
