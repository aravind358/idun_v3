package decision

import (
	"testing"
	"time"

	"idun/intelligence/communication"
)

func TestEnvelopeFromDecisionRecord_Success(t *testing.T) {
	rec := &DecisionRecord{
		DecisionID:          "dec-1234",
		EpisodeID:           "ep-99",
		SchemaVersion:       "2.0.0-FROZEN",
		Timestamp:           time.Now(),
		StrategyVersion:     "v2.0.0-test",
		DeliberationDepth:   DepthDeliberative,
		SelectedOutcome:     OutcomeCommit,
		SelectedCandidateID: "cand-1",
		Confidence:          0.92,
		Rationale:           "optimal tradeoff selected",
	}

	env, err := EnvelopeFromDecisionRecord(rec, "storage://cas/ref123")
	if err != nil {
		t.Fatalf("EnvelopeFromDecisionRecord error: %v", err)
	}

	if env.Source != "CognitiveAbility.Decision" {
		t.Errorf("expected Source 'CognitiveAbility.Decision', got '%s'", env.Source)
	}
	if env.Topic != communication.TopicEvaluatedOptions {
		t.Errorf("expected Topic '%s', got '%s'", communication.TopicEvaluatedOptions, env.Topic)
	}
	if env.PayloadRef != "storage://cas/ref123" {
		t.Errorf("expected PayloadRef 'storage://cas/ref123', got '%s'", env.PayloadRef)
	}
	if env.RawConfidence != 0.92 {
		t.Errorf("expected RawConfidence 0.92, got %f", env.RawConfidence)
	}
}

func TestMarshalUnmarshalDecisionRecord(t *testing.T) {
	original := &DecisionRecord{
		DecisionID:          "dec-roundtrip",
		EpisodeID:           "ep-rt",
		SchemaVersion:       "2.0.0-FROZEN",
		Timestamp:           time.Now().Truncate(time.Millisecond),
		StrategyVersion:     "v2.0.0",
		DeliberationDepth:   DepthDeliberative,
		SelectedOutcome:     OutcomeCommit,
		SelectedCandidateID: "cand-rt",
		Confidence:          0.88,
		Rationale:           "test roundtrip",
	}

	bytes, err := MarshalDecisionRecord(original)
	if err != nil {
		t.Fatalf("MarshalDecisionRecord error: %v", err)
	}

	decoded, err := UnmarshalDecisionRecord(bytes)
	if err != nil {
		t.Fatalf("UnmarshalDecisionRecord error: %v", err)
	}

	if decoded.DecisionID != original.DecisionID || decoded.Confidence != original.Confidence {
		t.Errorf("mismatch after roundtrip: got %+v", decoded)
	}
}
