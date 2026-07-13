package reflection

import (
	"context"
	"testing"

	"idun/intelligence/communication"
)

func FuzzReflectionReportValidate(f *testing.F) {
	f.Add(SchemaVersion, "refl-1", "ep-1", int64(100), string(ModeEpisode), 0.95, string(VerdictEvaluated))
	f.Add("invalid-ver", "", "", int64(-10), "INVALID_MODE", 1.5, "INVALID_VERDICT")

	f.Fuzz(func(t *testing.T, schemaVer, reportID, epID string, ts int64, modeStr string, conf float64, verdictStr string) {
		rep := ReflectionReport{
			SchemaVersion: schemaVer,
			ReportID:      reportID,
			EpisodeID:     epID,
			Timestamp:     ts,
			Mode:          ReflectionMode(modeStr),
			SpecialistReports: []SpecialistReport{
				{
					SpecialistID:         "spec-fuzz",
					TargetAbility:        "Reasoning",
					Verdict:              EvaluationVerdict(verdictStr),
					ReflectionConfidence: conf,
				},
			},
		}
		// Validate must return an error or nil without ever panicking
		_ = rep.Validate()
	})
}

func FuzzHistoricalSummaryValidate(f *testing.F) {
	f.Add(SchemaVersion, "sum-1", int64(100), int64(0), int64(100), 0.95)
	f.Add("bad-ver", "", int64(-5), int64(200), int64(100), -0.2)

	f.Fuzz(func(t *testing.T, schemaVer, sumID string, genTs, startTs, endTs int64, conf float64) {
		sum := HistoricalSummary{
			SchemaVersion:      schemaVer,
			SummaryID:          sumID,
			GeneratedTimestamp: genTs,
			TimeWindow:         TimeWindowSpec{StartTime: startTs, EndTime: endTs},
			SummaryConfidence:  conf,
		}
		_ = sum.Validate()
	})
}

func FuzzReflectEpisodeTraces(f *testing.F) {
	f.Add("ep-fuzz-1", "trace-id-1", "source-1", "payload")
	f.Add("", "", "", "")

	srv := NewService()
	_ = srv.Start()
	defer srv.Close()

	f.Fuzz(func(t *testing.T, epID, envID, source, payload string) {
		traces := []communication.Envelope{
			{ID: envID, Source: source, PayloadRef: payload},
		}
		_, _ = srv.ReflectEpisode(context.Background(), epID, traces)
	})
}

func FuzzReflectOnReflectionInputs(f *testing.F) {
	f.Add("sum-fuzz-meta", 0.95, 0.60)

	srv := NewService()
	_ = srv.Start()
	defer srv.Close()

	f.Fuzz(func(t *testing.T, sumID string, priorConf, actualScore float64) {
		prior := ReflectionReport{
			SchemaVersion: SchemaVersion,
			ReportID:      "refl-prior",
			EpisodeID:     "ep-1",
			Timestamp:     100,
			Mode:          ModeEpisode,
			SpecialistReports: []SpecialistReport{
				{
					SpecialistID:         "spec-1",
					TargetAbility:        "Reasoning",
					Verdict:              VerdictEvaluated,
					ReflectionConfidence: priorConf,
				},
			},
		}
		summary := HistoricalSummary{
			SchemaVersion:      SchemaVersion,
			SummaryID:          sumID,
			GeneratedTimestamp: 200,
			TimeWindow:         TimeWindowSpec{StartTime: 0, EndTime: 200},
			AverageScores: map[string]float64{
				"Reasoning": actualScore,
			},
			SummaryConfidence: 0.90,
		}
		_, _ = srv.ReflectOnReflection(context.Background(), []ReflectionReport{prior}, summary)
	})
}
