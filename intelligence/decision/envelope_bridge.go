package decision

import (
	"encoding/json"
	"fmt"
	"time"

	"idun/intelligence/communication"
)

// EnvelopeFromDecisionRecord packages a completed DecisionRecord into a canonical Global Workspace Envelope.
// Per Section 4.1 of the frozen specification, only Deliberative macro-decisions are published
// to the Global Workspace on TopicEvaluatedOptions.
func EnvelopeFromDecisionRecord(rec *DecisionRecord, payloadRef string) (communication.Envelope, error) {
	if rec == nil {
		return communication.Envelope{}, ErrInvalidDecisionRecord
	}
	if payloadRef == "" {
		return communication.Envelope{}, fmt.Errorf("decision: payloadRef cannot be empty when publishing DecisionRecord")
	}

	urgency := 0
	if rec.SelectedOutcome == OutcomeEscalateToDeliberative {
		urgency = 50
	} else if len(rec.RejectedCandidates) > 0 {
		for _, rej := range rec.RejectedCandidates {
			if rej.RejectionStage == "TIER_1_CONSTITUTION" {
				urgency = 80 // Elevate urgency if constitutional veto occurred
				break
			}
		}
	}

	env := communication.Envelope{
		ID:                fmt.Sprintf("env-dec-%s", rec.DecisionID),
		Source:            "CognitiveAbility.Decision",
		Topic:             communication.TopicEvaluatedOptions,
		PayloadRef:        payloadRef,
		PayloadModality:   "structured-frame",
		RawConfidence:     rec.Confidence,
		Urgency:           urgency,
		CostEstimateUnits: 10, // nominal base units for evaluated decision processing
		CreatedAt:         time.Now().UTC(),
	}

	if err := env.Validate(); err != nil {
		return communication.Envelope{}, err
	}

	return env, nil
}

// MarshalDecisionRecord serializes a DecisionRecord into canonical JSON bytes.
func MarshalDecisionRecord(rec *DecisionRecord) ([]byte, error) {
	if rec == nil {
		return nil, ErrInvalidDecisionRecord
	}
	return json.Marshal(rec)
}

// UnmarshalDecisionRecord deserializes JSON bytes into a DecisionRecord struct.
func UnmarshalDecisionRecord(data []byte) (*DecisionRecord, error) {
	if len(data) == 0 {
		return nil, ErrInvalidDecisionRecord
	}
	var rec DecisionRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, fmt.Errorf("decision: failed to unmarshal DecisionRecord: %w", err)
	}
	return &rec, nil
}
