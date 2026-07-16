package learning

import (
	"math"
)

// ComputeTraceStatisticalSummary computes bounded O(1) statistical telemetry across
// the ingested corpus and validation results of a single learning cycle.
func ComputeTraceStatisticalSummary(summary *AggregationSummary, candidates []*CandidateSnapshot) *TraceStatisticalSummary {
	if summary == nil {
		return &TraceStatisticalSummary{
			MeanValidationScore: 0.0,
			MinValidationScore:  0.0,
			MaxValidationScore:  0.0,
			ReplayFidelityRatio: 1.0,
		}
	}

	stats := &TraceStatisticalSummary{
		TotalArtifactsAnalyzed: summary.TotalArtifactsIngested,
		EstimatedDriftScore:    computeDriftScore(summary),
	}

	if len(summary.DomainSchemaIDs) > 0 {
		stats.DomainCoverageRatio = math.Min(1.0, float64(len(summary.DomainSchemaIDs))/10.0)
	} else {
		stats.DomainCoverageRatio = 0.1
	}

	if len(candidates) == 0 {
		stats.MeanValidationScore = 0.0
		stats.MinValidationScore = 0.0
		stats.MaxValidationScore = 0.0
		stats.ReplayFidelityRatio = 1.0
		return stats
	}

	var sumScore, minScore, maxScore float64
	var totalChecks, passedReplays int
	minScore = math.MaxFloat64

	for _, cand := range candidates {
		if cand == nil {
			continue
		}
		if cand.Lineage.SourceArtifactHash == summary.SourceArtifactHash && summary.SourceArtifactHash != "" {
			passedReplays++
		}
		for _, vr := range cand.ValidationRecords {
			sumScore += vr.Score
			if vr.Score < minScore {
				minScore = vr.Score
			}
			if vr.Score > maxScore {
				maxScore = vr.Score
			}
			totalChecks++
		}
	}

	if totalChecks > 0 {
		stats.MeanValidationScore = sumScore / float64(totalChecks)
		if minScore == math.MaxFloat64 {
			minScore = 0.0
		}
		stats.MinValidationScore = minScore
		stats.MaxValidationScore = maxScore
	} else {
		stats.MinValidationScore = 0.0
		stats.MaxValidationScore = 0.0
	}

	if len(candidates) > 0 {
		stats.ReplayFidelityRatio = float64(passedReplays) / float64(len(candidates))
	} else {
		stats.ReplayFidelityRatio = 1.0
	}

	stats.RejectionSummary = &CandidateRejectionSummary{}
	return stats
}

// ComputeTraceStatisticalSummaryWithRejections computes statistical summary and attaches the exact rejection summary.
func ComputeTraceStatisticalSummaryWithRejections(summary *AggregationSummary, candidates []*CandidateSnapshot, rejections *CandidateRejectionSummary) *TraceStatisticalSummary {
	stats := ComputeTraceStatisticalSummary(summary, candidates)
	if rejections != nil {
		stats.RejectionSummary = rejections
	}
	return stats
}

// UpdateLearnerPerformanceSummary updates bounded O(1) historical metrics for a learner
// using the results of a newly completed learning cycle.
func UpdateLearnerPerformanceSummary(existing *LearnerPerformanceSummary, usage LearnerUsage, version, fingerprint string) *LearnerPerformanceSummary {
	var target LearnerPerformanceSummary
	if existing == nil {
		target = LearnerPerformanceSummary{
			LearnerID:          usage.LearnerID,
			LearnerVersion:     version,
			LearnerFingerprint: fingerprint,
		}
	} else {
		target = *existing
		if version != "" {
			target.LearnerVersion = version
		}
		if fingerprint != "" {
			target.LearnerFingerprint = fingerprint
		}
	}

	target.Executions++
	target.AcceptedCandidates += uint64(usage.CandidatesAccepted)
	rejected := uint64(usage.CandidatesProduced - usage.CandidatesAccepted)
	if usage.CandidatesProduced < usage.CandidatesAccepted {
		rejected = 0
	}
	target.RejectedCandidates += rejected

	prevWeight := float64(target.Executions - 1)
	currWeight := float64(target.Executions)
	target.AverageValidationScore = (target.AverageValidationScore*prevWeight + usage.ContributionScore) / currWeight
	target.AverageExecutionTimeUs = uint64((float64(target.AverageExecutionTimeUs)*prevWeight + float64(usage.ExecutionTime.Microseconds())) / currWeight)

	totalProduced := target.AcceptedCandidates + target.RejectedCandidates
	if totalProduced > 0 {
		target.SuccessRatio = float64(target.AcceptedCandidates) / float64(totalProduced)
	} else {
		target.SuccessRatio = 1.0
	}

	return &target
}

// ComputeCandidateRejectionSummary aggregates rejected candidate reasons from check IDs.
func ComputeCandidateRejectionSummary(failedCheckIDs []string) *CandidateRejectionSummary {
	summary := &CandidateRejectionSummary{
		TotalRejected: uint64(len(failedCheckIDs)),
	}
	for _, checkID := range failedCheckIDs {
		switch checkID {
		case "STAT_SAMPLE_FLOOR":
			summary.StatisticalValidationFailures++
		case "SCHEMA_PAYLOAD_CHECK", "STRUCT_COMPATIBILITY_CHECK", "STRUCTURAL_CHECK":
			summary.StructuralValidationFailures++
		case "REPLAY_LINEAGE_CHECK", "REPLAY_CHECK":
			summary.ReplayValidationFailures++
		case "CAPABILITY_CHECK":
			summary.CapabilityValidationFailures++
		case "CONST_SAFETY_BOUNDS", "CONSTITUTIONAL_CHECK":
			summary.ConstitutionalValidationFailures++
		case "DUPLICATE_CHECK":
			summary.DuplicateCandidateFailures++
		default:
			summary.OtherFailures++
		}
	}
	return summary
}
