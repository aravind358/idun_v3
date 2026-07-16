package attention

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"idun/intelligence/communication"
)

// WorkspacePublisher defines the interface required to emit Envelopes to the Global Workspace.
type WorkspacePublisher interface {
	Publish(ctx context.Context, env communication.Envelope) error
}

// PayloadStorer defines the interface required to persist payloads to CAS storage.
type PayloadStorer interface {
	Store(ctx context.Context, data []byte) (string, error)
}

// FocusChangedPayload represents the event payload for focus target changes.
type FocusChangedPayload struct {
	PreviousFocus     string    `json:"previous_focus"`
	CurrentFocus      string    `json:"current_focus"`
	SwitchReason      string    `json:"switch_reason"`
	PolicyFingerprint string    `json:"policy_fingerprint"`
	Timestamp         time.Time `json:"timestamp"`
}

// InterruptPayload represents the event payload for interrupt acceptance or rejection.
type InterruptPayload struct {
	StimulusID        string           `json:"stimulus_id"`
	StimulusSource    string           `json:"stimulus_source"`
	Accepted          bool             `json:"accepted"`
	AssignedBand      PriorityBand     `json:"assigned_band"`
	Decision          SalienceDecision `json:"decision"`
	PolicyFingerprint string           `json:"policy_fingerprint"`
	Timestamp         time.Time        `json:"timestamp"`
}

// SalienceUpdatedPayload represents the event payload when salience evaluations occur.
type SalienceUpdatedPayload struct {
	StimulusID        string           `json:"stimulus_id"`
	SalienceScore     int              `json:"salience_score"`
	Decision          SalienceDecision `json:"decision"`
	AssignedBand      PriorityBand     `json:"assigned_band"`
	PolicyFingerprint string           `json:"policy_fingerprint"`
	Timestamp         time.Time        `json:"timestamp"`
}

// EnvelopeFromFocusChange formats a focus transition into a Global Workspace Envelope
// directed to TopicActiveGoals.
func EnvelopeFromFocusChange(ctx context.Context, storer PayloadStorer, payload FocusChangedPayload) (communication.Envelope, error) {
	if storer == nil {
		return communication.Envelope{}, fmt.Errorf("attention: PayloadStorer cannot be nil when publishing events")
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return communication.Envelope{}, err
	}
	payloadRef, err := storer.Store(ctx, data)
	if err != nil {
		return communication.Envelope{}, err
	}
	return communication.Envelope{
		ID:                fmt.Sprintf("env-focus-%d", time.Now().UnixNano()),
		Source:            "CognitiveAbility.Attention",
		Topic:             communication.TopicActiveGoals,
		PayloadRef:        payloadRef,
		PayloadModality:   "structured-frame",
		RawConfidence:     1.0,
		Urgency:           50,
		CostEstimateUnits: 0,
		CreatedAt:         time.Now().UTC(),
	}, nil
}

// EnvelopeFromInterrupt formats an interrupt event into a Global Workspace Envelope
// directed to TopicPerception.
func EnvelopeFromInterrupt(ctx context.Context, storer PayloadStorer, payload InterruptPayload) (communication.Envelope, error) {
	if storer == nil {
		return communication.Envelope{}, fmt.Errorf("attention: PayloadStorer cannot be nil when publishing events")
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return communication.Envelope{}, err
	}
	payloadRef, err := storer.Store(ctx, data)
	if err != nil {
		return communication.Envelope{}, err
	}
	urgency := 30
	if payload.Accepted && payload.AssignedBand <= PriorityBand1RealTime {
		urgency = 80
	}
	return communication.Envelope{
		ID:                fmt.Sprintf("env-interrupt-%d", time.Now().UnixNano()),
		Source:            "CognitiveAbility.Attention",
		Topic:             communication.TopicPerception,
		PayloadRef:        payloadRef,
		PayloadModality:   "structured-frame",
		RawConfidence:     1.0,
		Urgency:           urgency,
		CostEstimateUnits: 0,
		CreatedAt:         time.Now().UTC(),
	}, nil
}

// EnvelopeFromSalience formats a salience evaluation event into a Global Workspace Envelope
// directed to TopicPerception.
func EnvelopeFromSalience(ctx context.Context, storer PayloadStorer, payload SalienceUpdatedPayload) (communication.Envelope, error) {
	if storer == nil {
		return communication.Envelope{}, fmt.Errorf("attention: PayloadStorer cannot be nil when publishing events")
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return communication.Envelope{}, err
	}
	payloadRef, err := storer.Store(ctx, data)
	if err != nil {
		return communication.Envelope{}, err
	}
	return communication.Envelope{
		ID:                fmt.Sprintf("env-salience-%d", time.Now().UnixNano()),
		Source:            "CognitiveAbility.Attention",
		Topic:             communication.TopicPerception,
		PayloadRef:        payloadRef,
		PayloadModality:   "structured-frame",
		RawConfidence:     1.0,
		Urgency:           20,
		CostEstimateUnits: 0,
		CreatedAt:         time.Now().UTC(),
	}, nil
}
