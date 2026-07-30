package v3

import (
	"encoding/json"
	"idun/core/foundation"
	"testing"
	"time"
)

func TestDecisionRecord_Validation(t *testing.T) {
	artID, _ := foundation.NewUUID()
	parentID, _ := foundation.NewUUID()
	envID, _ := foundation.NewUUID()

	validBuilder := NewBuilder().
		ArtifactID(foundation.ArtifactID(artID)).
		ParentArtifactID(foundation.ParentArtifactID(parentID)).
		EnvelopeID(foundation.EnvelopeID(envID)).
		Timestamp(foundation.Timestamp(time.Now())).
		Resolution(StatusApproved)

	if _, err := validBuilder.Build(); err != nil {
		t.Errorf("expected valid artifact, got: %v", err)
	}

	badResBuilder := NewBuilder().
		ArtifactID(foundation.ArtifactID(artID)).
		ParentArtifactID(foundation.ParentArtifactID(parentID)).
		EnvelopeID(foundation.EnvelopeID(envID)).
		Timestamp(foundation.Timestamp(time.Now())).
		Resolution("NOT_A_STATUS")
	if _, err := badResBuilder.Build(); err == nil {
		t.Errorf("expected validation err for bad status")
	}
}

func TestDecisionRecord_Serialization(t *testing.T) {
	artID, _ := foundation.NewUUID()
	parentID, _ := foundation.NewUUID()
	envID, _ := foundation.NewUUID()

	f := NewDecisionFinding("Safety", "node-1", false, "ERR_NUCLEAR", "Nuclear launch rejected")

	rec, _ := NewBuilder().
		ArtifactID(foundation.ArtifactID(artID)).
		ParentArtifactID(foundation.ParentArtifactID(parentID)).
		EnvelopeID(foundation.EnvelopeID(envID)).
		Timestamp(foundation.Timestamp(time.Now())).
		Resolution(StatusRejected).
		Reason("Safety check failed").
		EvaluationFlags(false, true, true, true).
		Findings([]DecisionFinding{f}).
		Build()

	data, err := json.Marshal(rec)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var r2 DecisionRecord
	if err := json.Unmarshal(data, &r2); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if r2.Resolution() != StatusRejected {
		t.Errorf("mismatch resolution")
	}
	if len(r2.Findings()) != 1 || r2.Findings()[0].Code() != "ERR_NUCLEAR" {
		t.Errorf("mismatch findings")
	}
	if r2.SafetyPassed() != false || r2.PolicyPassed() != true {
		t.Errorf("mismatch evaluation flags")
	}
}
