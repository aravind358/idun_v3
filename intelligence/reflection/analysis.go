package reflection

import (
	"fmt"
	"math"
)

// AnalyzeTrends inspects read-only longitudinal HistoricalSummary metrics from Memory
// and surfaces multi-episode behavioral shifts bounded by MaxTrendFindings.
// It never mutates Memory and never performs Learning parameter adjustments.
func AnalyzeTrends(summary HistoricalSummary) ([]TrendFinding, error) {
	if err := summary.Validate(); err != nil {
		return nil, fmt.Errorf("invalid historical summary for trend analysis: %w", err)
	}

	findings := make([]TrendFinding, 0, MaxTrendFindings)

	// Analyze improving/degrading Reasoning
	if rate, ok := summary.ImprovementRates["Reasoning"]; ok {
		pt := "IMPROVING_REASONING"
		desc := "Deductive inference consistency is trending upward across recent episodes"
		if rate < -0.01 {
			pt = "DEGRADING_REASONING"
			desc = "Reasoning accuracy or consistency has declined across recent episodes"
		}
		findings = append(findings, TrendFinding{
			FindingID:     fmt.Sprintf("trend-reasoning-%s", summary.SummaryID),
			TargetAbility: "Reasoning",
			PatternType:   pt,
			Description:   desc,
			TrendVelocity: rate,
			Confidence:    summary.SummaryConfidence,
		})
	}

	// Analyze degrading conversation / communication
	if rate, ok := summary.ImprovementRates["Communication"]; ok {
		if len(findings) < MaxTrendFindings {
			pt := "IMPROVING_CONVERSATION"
			desc := "Conversational framing and clarity are improving"
			if rate < -0.01 {
				pt = "DEGRADING_CONVERSATION"
				desc = "Conversational coherence or responsiveness is degrading"
			}
			findings = append(findings, TrendFinding{
				FindingID:     fmt.Sprintf("trend-comm-%s", summary.SummaryID),
				TargetAbility: "Communication",
				PatternType:   pt,
				Description:   desc,
				TrendVelocity: rate,
				Confidence:    summary.SummaryConfidence,
			})
		}
	}

	// Analyze recurring planning failures
	if fr, ok := summary.FailureRates["Planning"]; ok && fr > 0.10 {
		if len(findings) < MaxTrendFindings {
			findings = append(findings, TrendFinding{
				FindingID:     fmt.Sprintf("trend-planning-fail-%s", summary.SummaryID),
				TargetAbility: "Planning",
				PatternType:   "RECURRING_PLANNING_FAILURES",
				Description:   fmt.Sprintf("Planning failure rate is elevated at %.1f%% across time window", fr*100.0),
				TrendVelocity: -fr,
				Confidence:    math.Min(1.0, summary.SummaryConfidence),
			})
		}
	}

	// Analyze attention drift
	if drift, ok := summary.TrendMetrics["AttentionDrift"]; ok && drift > 0.05 {
		if len(findings) < MaxTrendFindings {
			findings = append(findings, TrendFinding{
				FindingID:     fmt.Sprintf("trend-attn-drift-%s", summary.SummaryID),
				TargetAbility: "Executive",
				PatternType:   "ATTENTION_DRIFT",
				Description:   fmt.Sprintf("Attentional focus drift observed (velocity %.2f)", drift),
				TrendVelocity: drift,
				Confidence:    summary.SummaryConfidence,
			})
		}
	}

	// Ensure we don't exceed cardinality invariant
	if len(findings) > MaxTrendFindings {
		findings = findings[:MaxTrendFindings]
	}
	return findings, nil
}

// AnalyzeCrossCognitiveEpisode inspects interactions across specialist reports within a single episode
// without modifying cognitive abilities or specialist outputs.
func AnalyzeCrossCognitiveEpisode(reports []SpecialistReport) []CrossCognitiveFinding {
	findings := make([]CrossCognitiveFinding, 0, MaxCrossCognitiveFindings)

	var reasoningReport *SpecialistReport
	var commReport *SpecialistReport

	for i := range reports {
		r := &reports[i]
		if r.TargetAbility == "Reasoning" {
			reasoningReport = r
		} else if r.TargetAbility == "Communication" || r.TargetAbility == "Understanding" {
			commReport = r
		}
	}

	if reasoningReport != nil && commReport != nil {
		if reasoningReport.Verdict == VerdictEvaluated && reasoningReport.ReflectionConfidence >= 0.85 &&
			commReport.ReflectionConfidence < 0.75 {
			findings = append(findings, CrossCognitiveFinding{
				FindingID:        fmt.Sprintf("cc-%s-%s", reasoningReport.SpecialistID, commReport.SpecialistID),
				SourceAbilities:  []string{reasoningReport.TargetAbility, commReport.TargetAbility},
				InteractionIssue: "Reasoning proof quality exceeds Communication/Understanding presentation fidelity",
				Recommendation:   "Align communication framing with underlying logical proof structure",
				SeverityScore:    0.30,
				SourceTraceRefs:  reasoningReport.SourceTraceRefs,
			})
		}
	}

	if len(findings) > MaxCrossCognitiveFindings {
		findings = findings[:MaxCrossCognitiveFindings]
	}
	return findings
}

// AnalyzeCrossCognitivePeriodic evaluates inter-ability interactions across historical summaries.
func AnalyzeCrossCognitivePeriodic(summary HistoricalSummary) []CrossCognitiveFinding {
	findings := make([]CrossCognitiveFinding, 0, MaxCrossCognitiveFindings)

	rScore, rOk := summary.AverageScores["Reasoning"]
	cScore, cOk := summary.AverageScores["Communication"]

	if rOk && cOk && (rScore-cScore) > 0.15 {
		findings = append(findings, CrossCognitiveFinding{
			FindingID:        fmt.Sprintf("cc-period-%s", summary.SummaryID),
			SourceAbilities:  []string{"Reasoning", "Communication"},
			InteractionIssue: fmt.Sprintf("Excellent Reasoning average (%.2f) diverges from Communication average (%.2f)", rScore, cScore),
			Recommendation:   "Focus learning efforts on conversational articulation of complex deductions",
			SeverityScore:    math.Min(1.0, rScore-cScore),
			SourceTraceRefs:  []TraceReference{},
		})
	}

	pScore, pOk := summary.AverageScores["Planning"]
	dScore, dOk := summary.AverageScores["Decision"]

	if pOk && dOk && (pScore-dScore) > 0.15 {
		if len(findings) < MaxCrossCognitiveFindings {
			findings = append(findings, CrossCognitiveFinding{
				FindingID:        fmt.Sprintf("cc-plan-dec-%s", summary.SummaryID),
				SourceAbilities:  []string{"Planning", "Decision"},
				InteractionIssue: fmt.Sprintf("Good Planning (%.2f) but weaker Decision trade-off execution (%.2f)", pScore, dScore),
				Recommendation:   "Recalibrate decision utility evaluation against hierarchical task plans",
				SeverityScore:    math.Min(1.0, pScore-dScore),
				SourceTraceRefs:  []TraceReference{},
			})
		}
	}

	return findings
}

// EstimateGrowthPotentialPeriodic proposes where Learning is likely to obtain the greatest long-term benefit.
// Reflection proposes; Learning decides. Reflection never performs parameter optimization itself.
func EstimateGrowthPotentialPeriodic(summary HistoricalSummary) ([]GrowthPotentialEstimate, []RecommendedLearningSignal) {
	estimates := make([]GrowthPotentialEstimate, 0, MaxRecommendations)
	signals := make([]RecommendedLearningSignal, 0, MaxRecommendations)

	for ability, score := range summary.AverageScores {
		if len(estimates) >= MaxRecommendations {
			break
		}
		// Lower average score or high failure rate indicates higher growth potential ROI
		growthPotential := math.Max(0.1, math.Min(1.0, 1.0-score+0.2))
		rationale := fmt.Sprintf("Targeting %s for skill consolidation has high expected ROI based on historical average %.2f", ability, score)

		estimates = append(estimates, GrowthPotentialEstimate{
			AbilityName:           ability,
			CurrentQualitySignal:  score,
			GrowthPotentialRating: growthPotential,
			Rationale:             rationale,
		})

		signals = append(signals, RecommendedLearningSignal{
			SignalID:      fmt.Sprintf("sig-%s-%s", ability, summary.SummaryID),
			TargetAbility: ability,
			SignalType:    "GROWTH_POTENTIAL_ROI",
			Description:   rationale,
			ExpectedROI:   growthPotential,
		})
	}

	return estimates, signals
}

// EstimateGrowthPotentialEpisode proposes learning targets from single-episode specialist reports.
func EstimateGrowthPotentialEpisode(reports []SpecialistReport, epID string) ([]GrowthPotentialEstimate, []RecommendedLearningSignal) {
	estimates := make([]GrowthPotentialEstimate, 0, MaxRecommendations)
	signals := make([]RecommendedLearningSignal, 0, MaxRecommendations)

	for _, rep := range reports {
		if len(estimates) >= MaxRecommendations {
			break
		}
		roi := math.Max(0.1, math.Min(1.0, 1.0-rep.ReflectionConfidence+0.15))
		estimates = append(estimates, GrowthPotentialEstimate{
			AbilityName:           rep.TargetAbility,
			CurrentQualitySignal:  rep.ReflectionConfidence,
			GrowthPotentialRating: roi,
			Rationale:             fmt.Sprintf("Episode evaluation for %s yielded confidence %.2f", rep.TargetAbility, rep.ReflectionConfidence),
		})
		signals = append(signals, RecommendedLearningSignal{
			SignalID:      fmt.Sprintf("sig-ep-%s-%s", rep.TargetAbility, epID),
			TargetAbility: rep.TargetAbility,
			SignalType:    "EPISODE_REFINEMENT",
			Description:   fmt.Sprintf("Targeting %s episode refinement", rep.TargetAbility),
			ExpectedROI:   roi,
		})
	}

	return estimates, signals
}
