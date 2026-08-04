package v3

import (
	"context"
	"time"
	"idun/boundary/perception"
	"idun/core/foundation"
	"idun/intelligence/understanding/v3/splitter"
)

// UnderstandingService defines the primary entry point for the Understanding layer.
type UnderstandingService interface {
	Analyze(ctx context.Context, env *perception.PerceptionEnvelope) (*UnderstandingBatch, error)
}

// Orchestrator coordinates the execution of specialists, evaluates their output,
// and synthesizes the final SemanticInterpretation.
type Orchestrator struct {
	grammar      Specialist
	neural       Specialist
	deliberative Specialist
	
	extractors  ExtractorRunner
	normalizers NormalizerRunner
	composers   ComposerRunner
	splitter    splitter.Splitter
}

// ExtractorRunner coordinates semantic extraction from a hypothesis.
type ExtractorRunner interface {
	Run(hyp Hypothesis, b *Builder)
}

// NormalizerRunner coordinates semantic normalization.
type NormalizerRunner interface {
	Run(b *Builder)
}

// ComposerRunner coordinates semantic composition.
type ComposerRunner interface {
	Run(b *Builder)
}

// NewOrchestrator creates a new Orchestrator with the given specialists.
func NewOrchestrator(grammar, neural, deliberative Specialist, extractors ExtractorRunner, norms NormalizerRunner, comps ComposerRunner, spl splitter.Splitter) *Orchestrator {
	return &Orchestrator{
		grammar:      grammar,
		neural:       neural,
		deliberative: deliberative,
		extractors:   extractors,
		normalizers:  norms,
		composers:    comps,
		splitter:     spl,
	}
}

// Analyze processes a PerceptionEnvelope through the cascade of specialists.
func (o *Orchestrator) Analyze(ctx context.Context, env *perception.PerceptionEnvelope) (*UnderstandingBatch, error) {
	rawInput := env.RawInput()
	
	// Split utterance into distinct goals
	isValidGoal := func(chunk string) bool {
		// Create a temporary envelope for the chunk
		chunkEnv, _ := perception.NewBuilder().
			EnvelopeID(string(env.EnvelopeID())).
			ArtifactID(string(env.ArtifactID())).
			Version(string(env.Version())).
			Timestamp(time.Time(env.Timestamp())).
			RawInput(chunk).
			Build()
		hyps := o.cascadeAnalyze(ctx, chunkEnv)
		primary, _, _ := EvaluateHypotheses(hyps)
		return primary.Intent() != "" && primary.Intent() != "unresolved_intent"
	}
	
	var chunks []string
	if o.splitter != nil {
		chunks = o.splitter.Split(rawInput, isValidGoal)
	} else {
		chunks = []string{rawInput}
	}
	
	totalGoals := len(chunks)
	var interps []*SemanticInterpretation
	
	for i, chunk := range chunks {
		// Process each chunk
		chunkEnv, _ := perception.NewBuilder().
			EnvelopeID(string(env.EnvelopeID())).
			ArtifactID(string(env.ArtifactID())).
			Version(string(env.Version())).
			Timestamp(time.Time(env.Timestamp())).
			RawInput(chunk).
			Build()
			
		hyps := o.cascadeAnalyze(ctx, chunkEnv)
		primary, ambiguitySet, status := EvaluateHypotheses(hyps)

		b := NewBuilder()
		o.extractors.Run(primary, b)
		
		if o.normalizers != nil {
			o.normalizers.Run(b)
		}

		if o.composers != nil {
			o.composers.Run(b)
		}

		entities := b.obj.Entities()
		refs := b.obj.References()
		anchors := b.obj.TemporalAnchors()
		composed := b.GetComposedTimestamps()
		
		var secondary []SecondaryIntent
		
		interp, _ := Synthesize(chunkEnv, primary, ambiguitySet, status, entities, refs, anchors, composed, secondary)
		
		// Inject lightweight metadata
		interp.goalIndex = i + 1
		interp.totalGoals = totalGoals
		
		interps = append(interps, interp)
	}
	
	batch := NewUnderstandingBatch(env.EnvelopeID(), foundation.ParentArtifactID(env.ArtifactID()), rawInput, interps)
	return batch, nil
}

func (o *Orchestrator) cascadeAnalyze(ctx context.Context, env *perception.PerceptionEnvelope) []Hypothesis {
	if o.grammar != nil {
		if hyps, err := o.grammar.Analyze(ctx, env); err == nil && len(hyps) > 0 {
			return hyps
		}
	}

	if o.neural != nil {
		if hyps, err := o.neural.Analyze(ctx, env); err == nil && len(hyps) > 0 {
			return hyps
		}
	}

	if o.deliberative != nil {
		if hyps, err := o.deliberative.Analyze(ctx, env); err == nil && len(hyps) > 0 {
			return hyps
		}
	}

	return []Hypothesis{}
}
