package reflection

import (
	"context"
	"fmt"
	"math"
	"time"

	"idun/intelligence/communication"
)

// CompareHistoricalOutcomes inspects prior ReflectionReports and compares predicted outcomes against
// actual performance in currentSummary to produce a validated diagnostic SelfCalibrationReport.
// It is read-only and never performs parameter adjustments or strategy self-modifications.
func CompareHistoricalOutcomes(priorReports []ReflectionReport, currentSummary HistoricalSummary) (*SelfCalibrationReport, error) {
	if err := currentSummary.Validate(); err != nil {
		return nil, fmt.Errorf("invalid current summary for historical comparison: %w", err)
	}

	var minTime, maxTime int64
	minTime = math.MaxInt64
	var totalConfidence float64
	var reportCount int
	biasIndicators := make([]string, 0, MaxSpecialistFindingsPerReport)

	for i := range priorReports {
		rep := &priorReports[i]
		if err := rep.Validate(); err != nil {
			return nil, fmt.Errorf("invalid prior reflection report %s: %w", rep.ReportID, err)
		}
		if rep.Timestamp < minTime {
			minTime = rep.Timestamp
		}
		if rep.Timestamp > maxTime {
			maxTime = rep.Timestamp
		}
		for _, sr := range rep.SpecialistReports {
			totalConfidence += sr.ReflectionConfidence
			reportCount++

			// Check if prior specialist was highly confident but summary average score is poor
			if score, ok := currentSummary.AverageScores[sr.TargetAbility]; ok {
				if sr.ReflectionConfidence >= 0.88 && score < 0.65 {
					bias := fmt.Sprintf("OVERESTIMATION_BIAS_%s", sr.TargetAbility)
					found := false
					for _, b := range biasIndicators {
						if b == bias {
							found = true
							break
						}
					}
					if !found && len(biasIndicators) < MaxSpecialistFindingsPerReport {
						biasIndicators = append(biasIndicators, bias)
					}
				}
			}
		}
	}

	if minTime == math.MaxInt64 {
		minTime = currentSummary.TimeWindow.StartTime
	}
	if maxTime == 0 {
		maxTime = currentSummary.TimeWindow.EndTime
	}

	avgPriorConf := 0.85
	if reportCount > 0 {
		avgPriorConf = totalConfidence / float64(reportCount)
	}

	// Compute overall average score across current summary
	var sumScore float64
	var scoreCount int
	for _, s := range currentSummary.AverageScores {
		sumScore += s
		scoreCount++
	}
	avgActualScore := 0.85
	if scoreCount > 0 {
		avgActualScore = sumScore / float64(scoreCount)
	}

	predError := math.Abs(avgPriorConf - avgActualScore)
	reliability := math.Max(0.0, math.Min(1.0, 1.0-predError))

	trend := "ACCURATE_CALIBRATION"
	if (avgPriorConf - avgActualScore) > 0.15 {
		trend = "SYSTEMATIC_OVERESTIMATION"
	} else if (avgActualScore - avgPriorConf) > 0.15 {
		trend = "SYSTEMATIC_UNDERESTIMATION"
	}

	reportID := fmt.Sprintf("selfcal-%s-%d", currentSummary.SummaryID, time.Now().UnixNano())
	summaryText := fmt.Sprintf("Diagnostic self-calibration: prior avg confidence %.2f vs actual avg performance %.2f (%s)",
		avgPriorConf, avgActualScore, trend)

	sc := &SelfCalibrationReport{
		ReportID:              reportID,
		EvaluatedPeriodStart:  minTime,
		EvaluatedPeriodEnd:    maxTime,
		PriorPredictionError:  predError,
		ReflectionReliability: reliability,
		CalibrationTrend:      trend,
		BiasIndicators:        biasIndicators,
		Summary:               summaryText,
	}

	if err := sc.Validate(); err != nil {
		return nil, fmt.Errorf("generated self calibration report failed validation: %w", err)
	}
	return sc, nil
}

// CompositeEvaluationStrategy supports composing multiple EvaluationStrategies (heuristic, statistical,
// neural, or LLM-based) into a hybrid evaluator without modifying any public API.
type CompositeEvaluationStrategy struct {
	id         string
	version    string
	strategies []EvaluationStrategy
}

// NewCompositeEvaluationStrategy initializes a hybrid composable evaluation strategy.
func NewCompositeEvaluationStrategy(id, version string, strategies ...EvaluationStrategy) *CompositeEvaluationStrategy {
	return &CompositeEvaluationStrategy{
		id:         id,
		version:    version,
		strategies: append([]EvaluationStrategy(nil), strategies...),
	}
}

// Evaluate runs all underlying strategies sequentially or hierarchically and synthesizes a combined SpecialistReport.
func (c *CompositeEvaluationStrategy) Evaluate(ctx context.Context, traces []communication.Envelope) (SpecialistReport, error) {
	if len(c.strategies) == 0 {
		return SpecialistReport{
			SpecialistID:         c.id,
			TargetAbility:        "Composite",
			Verdict:              VerdictAbstain,
			ReflectionConfidence: 0.0,
		}, nil
	}

	// Run first strategy as primary and augment findings from subsequent strategies
	primary, err := c.strategies[0].Evaluate(ctx, traces)
	if err != nil {
		return SpecialistReport{}, err
	}

	for i := 1; i < len(c.strategies); i++ {
		sec, secErr := c.strategies[i].Evaluate(ctx, traces)
		if secErr == nil && sec.Verdict == VerdictEvaluated {
			totalFindings := len(primary.WentWell) + len(primary.WentPoorly) + len(primary.CouldImprove)
			for _, w := range sec.WentWell {
				if totalFindings < MaxSpecialistFindingsPerReport {
					primary.WentWell = append(primary.WentWell, w)
					totalFindings++
				}
			}
			for _, p := range sec.WentPoorly {
				if totalFindings < MaxSpecialistFindingsPerReport {
					primary.WentPoorly = append(primary.WentPoorly, p)
					totalFindings++
				}
			}
		}
	}
	return primary, nil
}

// StrategyID returns the composite strategy ID.
func (c *CompositeEvaluationStrategy) StrategyID() string {
	return c.id
}

// Version returns the composite strategy version.
func (c *CompositeEvaluationStrategy) Version() string {
	return c.version
}
