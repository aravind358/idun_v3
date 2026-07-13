package reasoning

import (
	"context"
	"testing"

	"idun/core/memory"
)

func TestCSPCheckSpecialist_CheckConsistency(t *testing.T) {
	specialist := NewCSPCheckSpecialist()
	if specialist.ID() != StageS3CSPCheck {
		t.Errorf("expected stage ID %s, got %s", StageS3CSPCheck, specialist.ID())
	}

	hyps := []ReasoningHypothesis{
		{ID: "hyp-1", Conclusion: "Subject is admin", CalibratedConfidence: 0.9},
	}

	memRecords := []memory.Record{
		{ID: "contradict/1", Type: "contradiction"},
		{ID: "fact/1", Type: "fact"},
	}

	flags, err := specialist.CheckConsistency(context.Background(), hyps, memRecords)
	if err != nil {
		t.Fatalf("expected consistency check to succeed, got %v", err)
	}
	if len(flags) != 1 {
		t.Fatalf("expected 1 contradiction flag, got %d", len(flags))
	}
	if err := flags[0].Validate(); err != nil {
		t.Fatalf("flagged contradiction failed validation: %v", err)
	}
	if flags[0].DetectedAtStage != StageS3CSPCheck {
		t.Errorf("expected DetectedAtStage S3, got %s", flags[0].DetectedAtStage)
	}
}
