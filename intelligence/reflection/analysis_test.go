package reflection

import (
	"testing"
	"time"
)

func TestAnalyzeTrends_SuccessAndCardinality(t *testing.T) {
	summary := HistoricalSummary{
		SchemaVersion:      SchemaVersion,
		SummaryID:          "sum-trends",
		GeneratedTimestamp: time.Now().UnixNano(),
		AverageScores: map[string]float64{
			"Reasoning":     0.91,
			"Communication": 0.72,
		},
		TrendMetrics: map[string]float64{
			"AttentionDrift": 0.12,
		},
		FailureRates: map[string]float64{
			"Planning": 0.18,
		},
		ImprovementRates: map[string]float64{
			"Reasoning":     0.08,
			"Communication": -0.06,
		},
		SummaryConfidence: 0.95,
	}

	findings, err := AnalyzeTrends(summary)
	if err != nil {
		t.Fatalf("AnalyzeTrends failed: %v", err)
	}

	if len(findings) == 0 {
		t.Fatal("expected trend findings to be surfaced")
	}
	for _, tf := range findings {
		if err := tf.Validate(); err != nil {
			t.Errorf("invalid trend finding %s: %v", tf.FindingID, err)
		}
	}
}

func TestAnalyzeCrossCognitive_PeriodicAndEpisode(t *testing.T) {
	summary := HistoricalSummary{
		SchemaVersion:      SchemaVersion,
		SummaryID:          "sum-cc",
		GeneratedTimestamp: time.Now().UnixNano(),
		AverageScores: map[string]float64{
			"Reasoning":     0.94,
			"Communication": 0.70,
			"Planning":      0.88,
			"Decision":      0.65,
		},
		SummaryConfidence: 0.90,
	}

	ccPeriodic := AnalyzeCrossCognitivePeriodic(summary)
	if len(ccPeriodic) != 2 {
		t.Errorf("expected 2 cross-cognitive findings, got %d", len(ccPeriodic))
	}
	for _, cf := range ccPeriodic {
		if err := cf.Validate(); err != nil {
			t.Errorf("invalid cross cognitive finding: %v", err)
		}
	}

	reports := []SpecialistReport{
		{
			SpecialistID:         "spec-r",
			TargetAbility:        "Reasoning",
			Verdict:              VerdictEvaluated,
			ReflectionConfidence: 0.92,
		},
		{
			SpecialistID:         "spec-c",
			TargetAbility:        "Communication",
			Verdict:              VerdictEvaluated,
			ReflectionConfidence: 0.65,
		},
	}

	ccEpisode := AnalyzeCrossCognitiveEpisode(reports)
	if len(ccEpisode) != 1 {
		t.Errorf("expected 1 cross cognitive finding in episode, got %d", len(ccEpisode))
	}
}

func TestEstimateGrowthPotential_PeriodicAndEpisode(t *testing.T) {
	summary := HistoricalSummary{
		SchemaVersion:      SchemaVersion,
		SummaryID:          "sum-gp",
		GeneratedTimestamp: time.Now().UnixNano(),
		AverageScores: map[string]float64{
			"Reasoning":     0.90,
			"Communication": 0.65,
		},
		SummaryConfidence: 0.90,
	}

	estimates, signals := EstimateGrowthPotentialPeriodic(summary)
	if len(estimates) != 2 || len(signals) != 2 {
		t.Errorf("unexpected counts: estimates=%d, signals=%d", len(estimates), len(signals))
	}
	for _, gp := range estimates {
		if err := gp.Validate(); err != nil {
			t.Errorf("invalid growth potential estimate: %v", err)
		}
	}
	for _, s := range signals {
		if err := s.Validate(); err != nil {
			t.Errorf("invalid recommended learning signal: %v", err)
		}
	}
}
