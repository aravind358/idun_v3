package reflection

import (
	"context"
	"sync"
	"testing"

	"idun/intelligence/communication"
	"idun/intelligence/workspace"
)

func TestCompareHistoricalOutcomes_OverestimationBias(t *testing.T) {
	prior1 := ReflectionReport{
		SchemaVersion: SchemaVersion,
		ReportID:      "refl-prior-1",
		EpisodeID:     "ep-1",
		Timestamp:     1000,
		Mode:          ModeEpisode,
		SpecialistReports: []SpecialistReport{
			{
				SpecialistID:         "spec-r-1",
				TargetAbility:        "Reasoning",
				Verdict:              VerdictEvaluated,
				ReflectionConfidence: 0.95,
			},
		},
	}
	prior2 := ReflectionReport{
		SchemaVersion: SchemaVersion,
		ReportID:      "refl-prior-2",
		EpisodeID:     "ep-2",
		Timestamp:     1500,
		Mode:          ModeEpisode,
		SpecialistReports: []SpecialistReport{
			{
				SpecialistID:         "spec-r-2",
				TargetAbility:        "Reasoning",
				Verdict:              VerdictEvaluated,
				ReflectionConfidence: 0.92,
			},
		},
	}

	summary := HistoricalSummary{
		SchemaVersion:      SchemaVersion,
		SummaryID:          "sum-meta",
		GeneratedTimestamp: 2000,
		TimeWindow:         TimeWindowSpec{StartTime: 1000, EndTime: 2000},
		EpisodeCount:       10,
		AverageScores: map[string]float64{
			"Reasoning": 0.60,
		},
		SummaryConfidence: 0.95,
	}

	selfCal, err := CompareHistoricalOutcomes([]ReflectionReport{prior1, prior2}, summary)
	if err != nil {
		t.Fatalf("CompareHistoricalOutcomes failed: %v", err)
	}

	if selfCal.CalibrationTrend != "SYSTEMATIC_OVERESTIMATION" {
		t.Errorf("got trend %s, want SYSTEMATIC_OVERESTIMATION", selfCal.CalibrationTrend)
	}
	if len(selfCal.BiasIndicators) == 0 || selfCal.BiasIndicators[0] != "OVERESTIMATION_BIAS_Reasoning" {
		t.Errorf("unexpected bias indicators: %v", selfCal.BiasIndicators)
	}
	if selfCal.ReflectionReliability >= 0.75 {
		t.Errorf("expected reduced reliability due to overestimation error, got %.2f", selfCal.ReflectionReliability)
	}
}

func TestService_ReflectOnReflectionPipeline(t *testing.T) {
	ws := workspace.NewEngine()
	srv := NewService(WithWorkspace(ws))
	_ = srv.Start()
	defer srv.Close()

	prior := ReflectionReport{
		SchemaVersion: SchemaVersion,
		ReportID:      "refl-p1",
		EpisodeID:     "ep-10",
		Timestamp:     100,
		Mode:          ModeEpisode,
		SpecialistReports: []SpecialistReport{
			{
				SpecialistID:         "spec-1",
				TargetAbility:        "Reasoning",
				Verdict:              VerdictEvaluated,
				ReflectionConfidence: 0.90,
			},
		},
	}

	summary := HistoricalSummary{
		SchemaVersion:      SchemaVersion,
		SummaryID:          "sum-meta-test",
		GeneratedTimestamp: 500,
		TimeWindow:         TimeWindowSpec{StartTime: 100, EndTime: 500},
		AverageScores: map[string]float64{
			"Reasoning": 0.65,
		},
		SummaryConfidence: 0.90,
	}

	report, err := srv.ReflectOnReflection(context.Background(), []ReflectionReport{prior}, summary)
	if err != nil {
		t.Fatalf("ReflectOnReflection failed: %v", err)
	}

	if err := report.Validate(); err != nil {
		t.Fatalf("generated reflection-on-reflection report failed validation: %v", err)
	}
	if report.SelfCalibration == nil {
		t.Fatal("expected SelfCalibration report to be populated")
	}
	if len(report.SessionNotes) == 0 {
		t.Error("expected session notes summarizing metacognitive audit")
	}
	if len(report.RecommendedLearningSignals) == 0 {
		t.Error("expected recommended learning signal for self-calibration refinement")
	}
}

func TestCompositeEvaluationStrategy_HybridFramework(t *testing.T) {
	strat1 := NewHeuristicEvaluationStrategy("s1", "Reasoning")
	strat2 := NewHeuristicEvaluationStrategy("s2", "Reasoning")
	hybrid := NewCompositeEvaluationStrategy("hybrid-strat", "2.0", strat1, strat2)

	if hybrid.StrategyID() != "hybrid-strat" || hybrid.Version() != "2.0" {
		t.Errorf("unexpected hybrid identity: %s / %s", hybrid.StrategyID(), hybrid.Version())
	}

	rep, err := hybrid.Evaluate(context.Background(), []communication.Envelope{
		{ID: "e1", Source: "test"},
	})
	if err != nil {
		t.Fatalf("hybrid Evaluate failed: %v", err)
	}
	if rep.SpecialistID != "s1" {
		t.Errorf("got specialist ID %s, want s1", rep.SpecialistID)
	}
}

func TestService_ReflectOnReflection_ConcurrentSafety(t *testing.T) {
	srv := NewService()
	_ = srv.Start()

	summary := HistoricalSummary{
		SchemaVersion:      SchemaVersion,
		SummaryID:          "sum-conc",
		GeneratedTimestamp: 100,
		TimeWindow:         TimeWindowSpec{StartTime: 0, EndTime: 100},
		AverageScores: map[string]float64{
			"Reasoning": 0.88,
		},
		SummaryConfidence: 0.90,
	}

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = srv.ReflectOnReflection(context.Background(), nil, summary)
		}()
	}
	wg.Wait()
	_ = srv.Close()
}
