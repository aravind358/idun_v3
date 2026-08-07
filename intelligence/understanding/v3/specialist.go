package v3

import (
	"context"
	"time"

	"idun/boundary/perception"
	"idun/intelligence/infrastructure/inference"
	"idun/intelligence/understanding"
	"idun/intelligence/workspace"
)

// Specialist represents a semantic parsing engine (e.g., Reflexive Grammar, Neural, Deliberative).
// It takes a PerceptionEnvelope and attempts to produce one or more semantic hypotheses.
type Specialist interface {
	Analyze(ctx context.Context, env *perception.PerceptionEnvelope) ([]Hypothesis, error)
}

// LegacyGrammarAdapter wraps a V1 GrammarSpecialist.
type LegacyGrammarAdapter struct {
	v1         understanding.GrammarSpecialist
	normalizer understanding.Normalizer
}

// NewDefaultGrammarSpecialist creates a V3 adapter for the default grammar specialist.
func NewDefaultGrammarSpecialist() Specialist {
	return &LegacyGrammarAdapter{
		v1:         understanding.NewDefaultGrammarSpecialist(),
		normalizer: understanding.NewDefaultNormalizer(),
	}
}

func (a *LegacyGrammarAdapter) Analyze(ctx context.Context, env *perception.PerceptionEnvelope) ([]Hypothesis, error) {
	norm := a.normalizer.Normalize(env.RawInput())
	v1Hyp, matched := a.v1.Evaluate(norm)
	if !matched {
		return []Hypothesis{}, nil
	}
	return convertHyps([]understanding.Hypothesis{v1Hyp}), nil
}

// LegacyNeuralAdapter wraps a V1 NeuralSpecialist.
type LegacyNeuralAdapter struct {
	v1         understanding.NeuralSpecialist
	normalizer understanding.Normalizer
}

// NewDefaultNeuralSpecialist creates a V3 adapter for the default neural specialist.
func NewDefaultNeuralSpecialist() Specialist {
	return &LegacyNeuralAdapter{
		v1:         understanding.NewDefaultNeuralSpecialist(),
		normalizer: understanding.NewDefaultNormalizer(),
	}
}

func (a *LegacyNeuralAdapter) Analyze(ctx context.Context, env *perception.PerceptionEnvelope) ([]Hypothesis, error) {
	norm := a.normalizer.Normalize(env.RawInput())
	v1Hyps, err := a.v1.Evaluate(norm)
	return convertHyps(v1Hyps), err
}

// LegacyDeliberativeAdapter wraps a V1 DeliberativeWorker.
type LegacyDeliberativeAdapter struct {
	v1 *understanding.DeliberativeWorker
}

// NewDeliberativeWorker creates a V3 adapter for the deliberative LLM worker.
func NewDeliberativeWorker(inf *inference.Service, ws *workspace.Engine, timeout time.Duration) Specialist {
	return &LegacyDeliberativeAdapter{v1: understanding.NewDeliberativeWorker(inf, ws, timeout)}
}

func (a *LegacyDeliberativeAdapter) Analyze(ctx context.Context, env *perception.PerceptionEnvelope) ([]Hypothesis, error) {
	frame, err := a.v1.InterpretDeliberative(ctx, string(env.EnvelopeID()), env.RawInput(), "")
	if err != nil {
		return []Hypothesis{}, err
	}
	v1Hyps := []understanding.Hypothesis{frame.PrimaryHypothesis}
	v1Hyps = append(v1Hyps, frame.AmbiguitySet...)
	return convertHyps(v1Hyps), nil
}

func convertHyps(v1Hyps []understanding.Hypothesis) []Hypothesis {
	var hyps []Hypothesis
	for _, h := range v1Hyps {
		var slots []Slot
		for _, s := range h.Slots {
			slots = append(slots, NewSlot(s.Name, s.Value, s.GroundingID, s.Confidence))
		}
		
		layer := LayerReflexiveGrammar
		switch string(h.SourceLayer) {
		case "Understanding.NeuralClassifier":
			layer = LayerNeuralClassifier
		case "Understanding.DeliberativeLLM":
			layer = LayerDeliberativeLLM
		}
		
		hyps = append(hyps, NewHypothesis(h.Intent, h.CalibratedConfidence, h.DeltaFromPrimary, layer, slots))
	}
	return hyps
}

