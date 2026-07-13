package reflection

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestReflectionReport_ValidationAndCardinality(t *testing.T) {
	validReport := ReflectionReport{
		SchemaVersion: SchemaVersion,
		ReportID:      "rep-001",
		EpisodeID:     "ep-100",
		Timestamp:     1700000000,
		Mode:          ModeEpisode,
		SpecialistReports: []SpecialistReport{
			{
				SpecialistID:         "spec-reasoning",
				TargetAbility:        "reasoning",
				Verdict:              VerdictEvaluated,
				WentWell:             []string{"Good symbolic deduction"},
				ReflectionConfidence: 0.9,
				SourceTraceRefs: []TraceReference{
					{
						EnvelopeID:     "env-1",
						SourceAbility:  "reasoning",
						TraceTimestamp: 1700000000,
					},
				},
			},
		},
	}

	if err := validReport.Validate(); err != nil {
		t.Fatalf("expected valid report to pass validation, got: %v", err)
	}

	// Test invalid schema version
	badSchema := validReport.Clone()
	badSchema.SchemaVersion = "1.0.0"
	if err := badSchema.Validate(); err == nil {
		t.Error("expected validation failure on invalid schema version")
	}

	// Test missing ReportID
	missingID := validReport.Clone()
	missingID.ReportID = ""
	if err := missingID.Validate(); err == nil {
		t.Error("expected validation failure on missing ReportID")
	}

	// Test missing EpisodeID in Episode mode
	missingEp := validReport.Clone()
	missingEp.EpisodeID = ""
	if err := missingEp.Validate(); err == nil {
		t.Error("expected validation failure on missing EpisodeID in Episode mode")
	}

	// Test cardinality limits (exceeding MaxSpecialistFindingsPerReport)
	cardinalityFail := validReport.Clone()
	manyFindings := make([]string, MaxSpecialistFindingsPerReport+1)
	for i := range manyFindings {
		manyFindings[i] = "finding"
	}
	cardinalityFail.SpecialistReports[0].WentWell = manyFindings
	if err := cardinalityFail.Validate(); err == nil {
		t.Error("expected validation failure when exceeding specialist findings limit")
	}

	// Test string length limit
	longStrFail := validReport.Clone()
	longStrFail.SpecialistReports[0].WentWell = []string{strings.Repeat("A", MaxStringLength+10)}
	if err := longStrFail.Validate(); err == nil {
		t.Error("expected validation failure when string exceeds MaxStringLength")
	}
}

func TestSpecialistReport_VerdictAndConfidenceValidation(t *testing.T) {
	sr := SpecialistReport{
		SpecialistID:         "spec-1",
		TargetAbility:        "planning",
		Verdict:              VerdictEvaluated,
		ReflectionConfidence: 0.88,
	}
	if err := sr.Validate(); err != nil {
		t.Fatalf("expected valid SpecialistReport to pass, got: %v", err)
	}

	// Invalid verdict
	badVerdict := sr
	badVerdict.Verdict = "UNKNOWN_VERDICT"
	if err := badVerdict.Validate(); err == nil {
		t.Error("expected validation failure on invalid verdict")
	}

	// Out of bounds confidence
	badConf := sr
	badConf.ReflectionConfidence = 1.5
	if err := badConf.Validate(); err == nil {
		t.Error("expected validation failure on out-of-bounds confidence")
	}
}

func TestHistoricalSummary_ValidationAndCloning(t *testing.T) {
	hs := HistoricalSummary{
		SchemaVersion:      SchemaVersion,
		SummaryID:          "sum-001",
		GeneratedTimestamp: 1700000000,
		TimeWindow:         TimeWindowSpec{StartTime: 1000, EndTime: 2000},
		EpisodeCount:       50,
		AverageScores:      map[string]float64{"reasoning": 0.88},
		FailureRates:       map[string]float64{"reasoning": 0.02},
		SummaryConfidence:  0.95,
	}
	if err := hs.Validate(); err != nil {
		t.Fatalf("expected valid HistoricalSummary to pass, got: %v", err)
	}

	clone := hs.Clone()
	clone.AverageScores["reasoning"] = 0.50
	if hs.AverageScores["reasoning"] == 0.50 {
		t.Error("expected Clone to create deep copy of map")
	}
}

func TestReflectionReport_JSONSerialization(t *testing.T) {
	report := ReflectionReport{
		SchemaVersion: SchemaVersion,
		ReportID:      "rep-json",
		EpisodeID:     "ep-json",
		Timestamp:     1700000000,
		Mode:          ModeEpisode,
	}
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("failed to marshal report: %v", err)
	}

	var decoded ReflectionReport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal report: %v", err)
	}
	if err := decoded.Validate(); err != nil {
		t.Fatalf("decoded report failed validation: %v", err)
	}
	if decoded.ReportID != report.ReportID {
		t.Errorf("got ReportID %s, want %s", decoded.ReportID, report.ReportID)
	}
}
