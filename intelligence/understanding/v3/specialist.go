package v3

import (
	"context"
	"idun/boundary/perception"
)

// Specialist represents a semantic parsing engine (e.g., Reflexive Grammar, Neural, Deliberative).
// It takes a PerceptionEnvelope and attempts to produce one or more semantic hypotheses.
type Specialist interface {
	Analyze(ctx context.Context, env *perception.PerceptionEnvelope) ([]Hypothesis, error)
}
