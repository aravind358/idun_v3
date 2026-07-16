package learning

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// CrossDomainLearner discovers relationships across distinct cognitive domains (Planning, Reasoning, Decision)
// and proposes cross-domain candidate snapshots (Draft) for offline validation without ever directly activating them.
type CrossDomainLearner struct{}

func NewCrossDomainLearner() *CrossDomainLearner {
	return &CrossDomainLearner{}
}

func (l *CrossDomainLearner) LearnerID() string {
	return "learner-cross-domain-v1"
}

func (l *CrossDomainLearner) LearnerVersion() string {
	return "1.0.0"
}

func (l *CrossDomainLearner) LearnerFingerprint() string {
	return "fp-learner-cross-domain-v1"
}

func (l *CrossDomainLearner) Consumes() []string {
	return []string{
		"idun.planning.trace.v1",
		"idun.reasoning.trace.v1",
		"idun.reflection.report.v1",
	}
}

func (l *CrossDomainLearner) Produces() []string {
	return []string{
		"idun.decision.strategy.v1",
		"idun.planning.heuristics.v1",
		"idun.decision.policy.v1",
	}
}

func (l *CrossDomainLearner) Generate(ctx context.Context, summary *AggregationSummary) ([]*CandidateSnapshot, error) {
	if summary == nil {
		return nil, fmt.Errorf("%w: aggregation summary cannot be nil", ErrValidationFailed)
	}
	if err := summary.Validate(); err != nil {
		return nil, fmt.Errorf("invalid aggregation summary: %w", err)
	}
	if summary.TotalArtifactsIngested == 0 {
		return nil, nil
	}

	var planningCount, reasoningCount, reflectionCount int
	for _, rec := range summary.Records {
		switch rec.Type {
		case "idun.planning.trace.v1":
			planningCount++
		case "idun.reasoning.trace.v1":
			reasoningCount++
		case "idun.reflection.report.v1":
			reflectionCount++
		}
	}

	if planningCount == 0 && reasoningCount == 0 && reflectionCount == 0 {
		return nil, nil
	}

	var candidates []*CandidateSnapshot
	nowNano := time.Now().UnixNano()

	// 1. Planning -> Decision: If planning bottleneck detected, propose decision risk/exploration adjustment
	if planningCount > 0 {
		payloadObj := map[string]interface{}{
			"cross_domain_source": "idun.planning.trace.v1",
			"target_adaptation":   "decision_strategy_exploration_boost",
			"exploration_bonus":   0.15,
			"records_analyzed":    planningCount,
			"synthesized_at":      nowNano,
		}
		payloadBytes, _ := json.Marshal(payloadObj)
		candidates = append(candidates, &CandidateSnapshot{
			SnapshotID: fmt.Sprintf("snap-xd-deci-%d", nowNano),
			SemVer:     l.LearnerVersion(),
			SchemaID:   "idun.decision.strategy.v1",
			Lifecycle:  LifecycleDraft,
			Lineage: ReplayMetadata{
				LearningFingerprint: LearningFingerprint(l.LearnerID()),
				PolicyFingerprint:   PolicyFingerprint(summary.AggregationPolicyID),
				LearnerFingerprint:  l.LearnerFingerprint(),
				SourceArtifactHash:  summary.SourceArtifactHash,
				ReplaySeed:          uint64(nowNano),
			},
			Payload: payloadBytes,
			Provenance: &CandidateLineage{
				ParentSnapshot:   "",
				AncestorSnapshot: fmt.Sprintf("snap-xd-deci-%d", nowNano),
				GenerationDepth:  0,
			},
		})
	}

	// 2. Reasoning -> Planning: If deep reasoning steps observed, propose refined planning heuristics
	if reasoningCount > 0 {
		payloadObj := map[string]interface{}{
			"cross_domain_source": "idun.reasoning.trace.v1",
			"target_adaptation":   "planning_heuristic_depth_alignment",
			"heuristic_weight":    0.85,
			"records_analyzed":    reasoningCount,
			"synthesized_at":      nowNano + 1,
		}
		payloadBytes, _ := json.Marshal(payloadObj)
		candidates = append(candidates, &CandidateSnapshot{
			SnapshotID: fmt.Sprintf("snap-xd-plan-%d", nowNano+1),
			SemVer:     l.LearnerVersion(),
			SchemaID:   "idun.planning.heuristics.v1",
			Lifecycle:  LifecycleDraft,
			Lineage: ReplayMetadata{
				LearningFingerprint: LearningFingerprint(l.LearnerID()),
				PolicyFingerprint:   PolicyFingerprint(summary.AggregationPolicyID),
				LearnerFingerprint:  l.LearnerFingerprint(),
				SourceArtifactHash:  summary.SourceArtifactHash,
				ReplaySeed:          uint64(nowNano + 1),
			},
			Payload: payloadBytes,
			Provenance: &CandidateLineage{
				ParentSnapshot:   "",
				AncestorSnapshot: fmt.Sprintf("snap-xd-plan-%d", nowNano+1),
				GenerationDepth:  0,
			},
		})
	}

	// 3. Reflection -> Decision Policy: If systemic anomalies across episodes, propose conservative policy updates
	if reflectionCount > 0 {
		payloadObj := map[string]interface{}{
			"cross_domain_source": "idun.reflection.report.v1",
			"target_adaptation":   "decision_policy_conservative_buffer",
			"safety_buffer":       0.25,
			"records_analyzed":    reflectionCount,
			"synthesized_at":      nowNano + 2,
		}
		payloadBytes, _ := json.Marshal(payloadObj)
		candidates = append(candidates, &CandidateSnapshot{
			SnapshotID: fmt.Sprintf("snap-xd-pol-%d", nowNano+2),
			SemVer:     l.LearnerVersion(),
			SchemaID:   "idun.decision.policy.v1",
			Lifecycle:  LifecycleDraft,
			Lineage: ReplayMetadata{
				LearningFingerprint: LearningFingerprint(l.LearnerID()),
				PolicyFingerprint:   PolicyFingerprint(summary.AggregationPolicyID),
				LearnerFingerprint:  l.LearnerFingerprint(),
				SourceArtifactHash:  summary.SourceArtifactHash,
				ReplaySeed:          uint64(nowNano + 2),
			},
			Payload: payloadBytes,
			Provenance: &CandidateLineage{
				ParentSnapshot:   "",
				AncestorSnapshot: fmt.Sprintf("snap-xd-pol-%d", nowNano+2),
				GenerationDepth:  0,
			},
		})
	}

	return candidates, nil
}
