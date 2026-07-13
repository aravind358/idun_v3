package reflection

import (
	"testing"
)

func TestReflectionReportBuilder_SuccessAndImmutability(t *testing.T) {
	sr := SpecialistReport{
		SpecialistID:         "spec-understanding",
		TargetAbility:        "understanding",
		Verdict:              VerdictEvaluated,
		ReflectionConfidence: 0.9,
	}

	report, err := NewReflectionReportBuilder("rep-builder-1", ModeEpisode, 1700000000).
		WithEpisodeID("ep-builder-1").
		AddSpecialistReport(sr).
		AddSessionNote("Checked slot framing accuracy").
		Build()

	if err != nil {
		t.Fatalf("expected builder to succeed, got: %v", err)
	}

	if report.ReportID != "rep-builder-1" {
		t.Errorf("got ReportID %s, want rep-builder-1", report.ReportID)
	}
	if len(report.SpecialistReports) != 1 {
		t.Fatalf("expected 1 specialist report, got %d", len(report.SpecialistReports))
	}
}

func TestReflectionReportBuilder_ValidationFailure(t *testing.T) {
	// Missing EpisodeID when Mode is ModeEpisode
	_, err := NewReflectionReportBuilder("rep-fail", ModeEpisode, 1700000000).Build()
	if err == nil {
		t.Error("expected builder to fail validation when missing EpisodeID in Episode mode")
	}
}

func TestHistoricalSummaryBuilder_Success(t *testing.T) {
	hs, err := NewHistoricalSummaryBuilder("sum-b1", 1700000000).
		WithTimeWindow(100, 200).
		WithEpisodeCount(100).
		AddAverageScore("reasoning", 0.91).
		AddFailureRate("planning", 0.05).
		Build()

	if err != nil {
		t.Fatalf("expected HistoricalSummaryBuilder to succeed, got: %v", err)
	}
	if hs.EpisodeCount != 100 {
		t.Errorf("got EpisodeCount %d, want 100", hs.EpisodeCount)
	}
}

func TestHistoricalSummaryRequestBuilder_Success(t *testing.T) {
	req, err := NewHistoricalSummaryRequestBuilder("req-b1").
		WithTimeWindow(100, 200).
		AddTargetAbility("reasoning").
		WithAggregationLevel("DAILY").
		AddRequestedMetric("failure_rates").
		Build()

	if err != nil {
		t.Fatalf("expected HistoricalSummaryRequestBuilder to succeed, got: %v", err)
	}
	if len(req.TargetAbilities) != 1 || req.TargetAbilities[0] != "reasoning" {
		t.Errorf("unexpected target abilities: %v", req.TargetAbilities)
	}
}
