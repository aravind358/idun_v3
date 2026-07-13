package reflection

import (
	"context"
	"testing"

	"idun/intelligence/communication"
)

func TestService_TelemetrySnapshot(t *testing.T) {
	srv := NewService()
	_ = srv.Start()
	defer srv.Close()

	// 1. Episode reflection
	_, _ = srv.ReflectEpisode(context.Background(), "ep-1", []communication.Envelope{
		{ID: "e1", Source: "test"},
	})

	// 2. Periodic reflection
	sum := HistoricalSummary{
		SchemaVersion:      SchemaVersion,
		SummaryID:          "sum-1",
		GeneratedTimestamp: 100,
		TimeWindow:         TimeWindowSpec{StartTime: 0, EndTime: 100},
		AverageScores: map[string]float64{
			"Reasoning": 0.85,
		},
		SummaryConfidence: 0.90,
	}
	_, _ = srv.ReflectPeriodic(context.Background(), sum)

	// 3. Meta reflection
	_, _ = srv.ReflectOnReflection(context.Background(), nil, sum)

	snap := srv.Telemetry()
	if snap.TotalEpisodeReflections != 1 {
		t.Errorf("got %d episode reflections, want 1", snap.TotalEpisodeReflections)
	}
	if snap.TotalPeriodicReflections != 1 {
		t.Errorf("got %d periodic reflections, want 1", snap.TotalPeriodicReflections)
	}
	if snap.TotalMetaReflections != 1 {
		t.Errorf("got %d meta reflections, want 1", snap.TotalMetaReflections)
	}
	if snap.SelfCalibrationRuns != 1 {
		t.Errorf("got %d self calibration runs, want 1", snap.SelfCalibrationRuns)
	}
}
