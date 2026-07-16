package learning

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"idun/core/memory"
)

// =============================================================================
// ReasoningLearner
// =============================================================================

// ReasoningLearner synthesizes updated reasoning heuristics and strategy proposals
// from historical reasoning traces and post-hoc reflection evaluations.
type ReasoningLearner struct{}

func NewReasoningLearner() *ReasoningLearner {
	return &ReasoningLearner{}
}

func (l *ReasoningLearner) LearnerID() string {
	return "learner-reasoning-heuristics-v1"
}

func (l *ReasoningLearner) LearnerVersion() string {
	return "1.0.0"
}

func (l *ReasoningLearner) LearnerFingerprint() string {
	return "fp-learner-reasoning-heuristics-v1"
}

func (l *ReasoningLearner) Consumes() []string {
	return []string{
		"idun.reasoning.trace.v1",
		"idun.reflection.report.v1",
	}
}

func (l *ReasoningLearner) Produces() []string {
	return []string{
		"idun.reasoning.strategy.v1",
	}
}

func (l *ReasoningLearner) Generate(ctx context.Context, summary *AggregationSummary) ([]*CandidateSnapshot, error) {
	if summary == nil {
		return nil, fmt.Errorf("%w: aggregation summary cannot be nil", ErrValidationFailed)
	}
	if err := summary.Validate(); err != nil {
		return nil, fmt.Errorf("invalid aggregation summary: %w", err)
	}
	if summary.TotalArtifactsIngested == 0 {
		return nil, nil // No candidates synthesized if window is empty
	}

	// Analyze ingested traces
	var traceCount int
	var reportCount int
	for _, rec := range summary.Records {
		switch rec.Type {
		case "idun.reasoning.trace.v1":
			traceCount++
		case "idun.reflection.report.v1":
			reportCount++
		}
	}

	if traceCount == 0 && reportCount == 0 {
		return nil, nil
	}

	// Synthesize refined strategy payload
	payloadObj := map[string]interface{}{
		"heuristics": map[string]float64{
			"deductive_weight":   0.88,
			"abductive_weight":   0.72,
			"confidence_penalty": 0.05,
		},
		"graph_depth_limit": 8,
		"traces_analyzed":   traceCount,
		"reports_analyzed":  reportCount,
		"synthesized_at":    time.Now().UnixNano(),
	}
	payloadBytes, err := json.Marshal(payloadObj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal reasoning strategy payload: %w", err)
	}

	snap := &CandidateSnapshot{
		SnapshotID: fmt.Sprintf("snap-reas-%d", time.Now().UnixNano()),
		SemVer:     l.LearnerVersion(),
		SchemaID:   "idun.reasoning.strategy.v1",
		Lifecycle:  LifecycleDraft,
		Lineage: ReplayMetadata{
			LearningFingerprint: LearningFingerprint(l.LearnerID()),
			PolicyFingerprint:   PolicyFingerprint(summary.AggregationPolicyID),
			LearnerFingerprint:  l.LearnerFingerprint(),
			SourceArtifactHash:  summary.SourceArtifactHash,
		},
		Payload: payloadBytes,
	}

	return []*CandidateSnapshot{snap}, nil
}

// =============================================================================
// PlanningLearner
// =============================================================================

// PlanningLearner synthesizes HTN and hierarchical planning adaptations from
// historical planning execution traces and reflection feedback.
type PlanningLearner struct{}

func NewPlanningLearner() *PlanningLearner {
	return &PlanningLearner{}
}

func (l *PlanningLearner) LearnerID() string {
	return "learner-planning-specialist-v1"
}

func (l *PlanningLearner) LearnerVersion() string {
	return "1.0.0"
}

func (l *PlanningLearner) LearnerFingerprint() string {
	return "fp-learner-planning-specialist-v1"
}

func (l *PlanningLearner) Consumes() []string {
	return []string{
		"idun.planning.trace.v1",
		"idun.reflection.report.v1",
	}
}

func (l *PlanningLearner) Produces() []string {
	return []string{
		"idun.planning.strategy.v1",
	}
}

func (l *PlanningLearner) Generate(ctx context.Context, summary *AggregationSummary) ([]*CandidateSnapshot, error) {
	if summary == nil {
		return nil, fmt.Errorf("%w: aggregation summary cannot be nil", ErrValidationFailed)
	}
	if err := summary.Validate(); err != nil {
		return nil, fmt.Errorf("invalid aggregation summary: %w", err)
	}
	if summary.TotalArtifactsIngested == 0 {
		return nil, nil
	}

	var traceCount int
	for _, rec := range summary.Records {
		if rec.Type == "idun.planning.trace.v1" || rec.Type == "idun.reflection.report.v1" {
			traceCount++
		}
	}
	if traceCount == 0 {
		return nil, nil
	}

	payloadObj := map[string]interface{}{
		"htn_weights": map[string]float64{
			"subtask_expansion_bias": 0.82,
			"pruning_threshold":      0.15,
		},
		"max_search_depth": 16,
		"traces_analyzed":  traceCount,
		"synthesized_at":   time.Now().UnixNano(),
	}
	payloadBytes, err := json.Marshal(payloadObj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal planning strategy payload: %w", err)
	}

	snap := &CandidateSnapshot{
		SnapshotID: fmt.Sprintf("snap-plan-%d", time.Now().UnixNano()),
		SemVer:     l.LearnerVersion(),
		SchemaID:   "idun.planning.strategy.v1",
		Lifecycle:  LifecycleDraft,
		Lineage: ReplayMetadata{
			LearningFingerprint: LearningFingerprint(l.LearnerID()),
			PolicyFingerprint:   PolicyFingerprint(summary.AggregationPolicyID),
			LearnerFingerprint:  l.LearnerFingerprint(),
			SourceArtifactHash:  summary.SourceArtifactHash,
		},
		Payload: payloadBytes,
	}

	return []*CandidateSnapshot{snap}, nil
}

// =============================================================================
// DecisionLearner
// =============================================================================

// DecisionLearner synthesizes multi-objective decision weights and risk thresholds
// from past decision traces and reflection outcomes.
type DecisionLearner struct{}

func NewDecisionLearner() *DecisionLearner {
	return &DecisionLearner{}
}

func (l *DecisionLearner) LearnerID() string {
	return "learner-decision-weights-v1"
}

func (l *DecisionLearner) LearnerVersion() string {
	return "1.0.0"
}

func (l *DecisionLearner) LearnerFingerprint() string {
	return "fp-learner-decision-weights-v1"
}

func (l *DecisionLearner) Consumes() []string {
	return []string{
		"idun.decision.trace.v1",
		"idun.reflection.report.v1",
	}
}

func (l *DecisionLearner) Produces() []string {
	return []string{
		"idun.decision.strategy.v1",
	}
}

func (l *DecisionLearner) Generate(ctx context.Context, summary *AggregationSummary) ([]*CandidateSnapshot, error) {
	if summary == nil {
		return nil, fmt.Errorf("%w: aggregation summary cannot be nil", ErrValidationFailed)
	}
	if err := summary.Validate(); err != nil {
		return nil, fmt.Errorf("invalid aggregation summary: %w", err)
	}
	if summary.TotalArtifactsIngested == 0 {
		return nil, nil
	}

	var traceCount int
	for _, rec := range summary.Records {
		if rec.Type == "idun.decision.trace.v1" || rec.Type == "idun.reflection.report.v1" {
			traceCount++
		}
	}
	if traceCount == 0 {
		return nil, nil
	}

	payloadObj := map[string]interface{}{
		"objective_weights": map[string]float64{
			"safety_priority":     0.95,
			"efficiency_priority": 0.65,
			"cost_penalty":        0.20,
		},
		"confidence_cutoff": 0.80,
		"traces_analyzed":   traceCount,
		"synthesized_at":    time.Now().UnixNano(),
	}
	payloadBytes, err := json.Marshal(payloadObj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal decision strategy payload: %w", err)
	}

	snap := &CandidateSnapshot{
		SnapshotID: fmt.Sprintf("snap-deci-%d", time.Now().UnixNano()),
		SemVer:     l.LearnerVersion(),
		SchemaID:   "idun.decision.strategy.v1",
		Lifecycle:  LifecycleDraft,
		Lineage: ReplayMetadata{
			LearningFingerprint: LearningFingerprint(l.LearnerID()),
			PolicyFingerprint:   PolicyFingerprint(summary.AggregationPolicyID),
			LearnerFingerprint:  l.LearnerFingerprint(),
			SourceArtifactHash:  summary.SourceArtifactHash,
		},
		Payload: payloadBytes,
	}

	return []*CandidateSnapshot{snap}, nil
}

// ensure _ memory.Record used
var _ memory.Record
