package learning

import (
	"context"
	"testing"
	"time"
)

func FuzzReplayMetadataValidation(f *testing.F) {
	f.Add("fp-learn-1", "fp-pol-1", "hash-src-1", uint64(12345), "snap-parent", "snap-ancestor", uint32(2), "fp-learner")
	f.Add("", "fp-pol-1", "hash-src-1", uint64(0), "", "", uint32(0), "")

	f.Fuzz(func(t *testing.T, lfp, pfp, srcHash string, seed uint64, parent, ancestor string, depth uint32, lnrFp string) {
		meta := &ReplayMetadata{
			LearningFingerprint: LearningFingerprint(lfp),
			PolicyFingerprint:   PolicyFingerprint(pfp),
			SourceArtifactHash:  SourceArtifactHash(srcHash),
			ReplaySeed:          seed,
			ParentSnapshot:      parent,
			AncestorSnapshot:    ancestor,
			GenerationDepth:     depth,
			LearnerFingerprint:  lnrFp,
		}
		_ = meta.Validate()
	})
}

func FuzzCandidateSnapshotValidation(f *testing.F) {
	f.Add("snap-fuzz-1", "1.0.0", "idun.reasoning.strategy.v1", []byte(`{"weights":{"a":1}}`), "fp-pol", "hash-src", uint32(1))
	f.Add("", "invalid-ver", "", []byte(`invalid-json`), "", "", uint32(0))

	f.Fuzz(func(t *testing.T, id, ver, schemaID string, payload []byte, pfp, srcHash string, depth uint32) {
		cand := &CandidateSnapshot{
			SnapshotID: id,
			SemVer:     ver,
			SchemaID:   schemaID,
			Lifecycle:  LifecycleDraft,
			Payload:    payload,
			Lineage: ReplayMetadata{
				LearningFingerprint: "fp-learn",
				PolicyFingerprint:   PolicyFingerprint(pfp),
				SourceArtifactHash:  SourceArtifactHash(srcHash),
				GenerationDepth:     depth,
			},
		}
		_ = cand.Validate()
	})
}

func FuzzLearningTraceValidation(f *testing.F) {
	f.Add("trace-fuzz-1", "req-fuzz", "idun.reasoning.trace.v1", "fp-pol", string(StatusPublished), string(ReasonSuccess))
	f.Add("", "", "", "", "INVALID_STATUS", "INVALID_REASON")

	f.Fuzz(func(t *testing.T, tid, rid, schemaID, pfp, statusStr, reasonStr string) {
		trace := &LearningTrace{
			TraceID:           tid,
			RequestID:         rid,
			DomainSchemaID:    schemaID,
			TraceTimestamp:    time.Now(),
			Status:            LearningResultStatus(statusStr),
			TerminationReason: LearningTerminationReason(reasonStr),
			Lineage: ReplayMetadata{
				LearningFingerprint: "fp-learn",
				PolicyFingerprint:   PolicyFingerprint(pfp),
			},
		}
		_ = trace.Validate()
	})
}

func FuzzAggregationPolicyValidation(f *testing.F) {
	f.Add("policy-fuzz-1", "1.0.0", "fp-fuzz", "COGNITIVE_PERFORMANCE_DEFAULT", int64(time.Hour), string(SamplingMethodDeterministic), string(OrderingStrategyChronologicalAsc), string(MergePolicyAppend), uint32(100), uint64(1024))
	f.Add("", "", "", "INVALID_STRAT", int64(-10), "INVALID_SAMP", "INVALID_ORD", "INVALID_MERGE", uint32(0), uint64(0))

	f.Fuzz(func(t *testing.T, id, ver, fp, strat string, winDur int64, samp, ord, merge string, maxArt uint32, maxMem uint64) {
		policy := &AggregationPolicy{
			PolicyID:           id,
			PolicyVersion:      ver,
			PolicyFingerprint:  fp,
			Strategy:           strat,
			WindowDuration:     time.Duration(winDur),
			SamplingMethod:     SamplingMethod(samp),
			OrderingStrategy:   OrderingStrategy(ord),
			MergePolicy:        MergePolicy(merge),
			MaximumArtifacts:   maxArt,
			MaximumMemoryBytes: maxMem,
		}
		_ = policy.Validate()
	})
}

func FuzzExperimentProfileValidation(f *testing.F) {
	f.Add("exp-fuzz-1", "idun.reasoning.strategy.v1", "snap-fuzz-target", float64(0.10), float64(0.01), int64(24*time.Hour), uint64(123))
	f.Add("", "", "", float64(-1.0), float64(2.0), int64(0), uint64(0))

	f.Fuzz(func(t *testing.T, expID, schemaID, targetID string, shadow, canary float64, dur int64, seed uint64) {
		profile := &ExperimentProfile{
			ExperimentID:     expID,
			DomainSchemaID:   schemaID,
			TargetSnapshotID: targetID,
			ShadowRatio:      shadow,
			CanaryRatio:      canary,
			MaxDuration:      time.Duration(dur),
			ReplaySeed:       seed,
		}
		_ = profile.Validate()
	})
}

func FuzzLearningCampaignSummaryValidation(f *testing.F) {
	f.Add("camp-fuzz-1", uint64(10), uint64(8), uint64(2), float64(0.85), float64(0.90), int64(time.Hour))
	f.Add("", uint64(5), uint64(10), uint64(0), float64(-1.0), float64(2.0), int64(0))

	f.Fuzz(func(t *testing.T, id string, cycles, cands, val uint64, conf, fidelity float64, dur int64) {
		summary := &LearningCampaignSummary{
			CampaignID:                  id,
			TotalCycles:                 cycles,
			TotalCandidates:             cands,
			TotalValidated:              val,
			AverageValidationConfidence: conf,
			AverageReplayFidelity:       fidelity,
			CampaignDuration:            time.Duration(dur),
		}
		_ = summary.Validate()
	})
}

func FuzzValidationPipelineWithPayload(f *testing.F) {
	f.Add([]byte(`{"weights":{"alpha":0.5}}`), float64(0.1), float64(0.9))
	f.Add([]byte(`invalid-json`), float64(-0.5), float64(2.0))

	vp := NewDefaultValidationPipeline(nil)
	ctx := context.Background()
	profile := DefaultLearningPolicyProfile()
	summary := &AggregationSummary{
		TotalArtifactsIngested: 10,
		SourceArtifactHash:     "hash-fuzz",
	}

	f.Fuzz(func(t *testing.T, payload []byte, driftScore, coverage float64) {
		cand := &CandidateSnapshot{
			SnapshotID: "cand-fuzz",
			SemVer:     "1.0.0",
			SchemaID:   "idun.reasoning.strategy.v1",
			Lifecycle:  LifecycleDraft,
			Payload:    payload,
			Lineage: ReplayMetadata{
				PolicyFingerprint:  profile.PolicyFingerprint,
				SourceArtifactHash: "hash-fuzz",
			},
		}
		_, _, _ = vp.ValidateCandidate(ctx, cand, summary, profile)
	})
}
