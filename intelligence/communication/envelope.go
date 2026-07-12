package communication

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// Sentinel validation errors for Envelope construction and inspection.
var (
	ErrInvalidEnvelope   = errors.New("communication: invalid envelope")
	ErrMissingID         = errors.New("communication: envelope ID is required")
	ErrMissingSource     = errors.New("communication: envelope Source is required")
	ErrInvalidTopic      = errors.New("communication: envelope Topic is invalid or unregistered")
	ErrMissingPayloadRef = errors.New("communication: envelope PayloadRef is required")
	ErrInvalidConfidence = errors.New("communication: RawConfidence must be within [0.0, 1.0]")
	ErrInvalidUrgency    = errors.New("communication: Urgency must be within [0, 100]")
	ErrNegativeCost      = errors.New("communication: CostEstimateUnits cannot be negative")
)

// Envelope defines the immutable control-plane message wrapper traversing the Global Workspace.
//
// Architectural Invariant: Executive Functions inspects ONLY the fields of this Envelope.
// Executive Functions MUST NEVER dereference, inspect, or parse PayloadRef.
type Envelope struct {
	// ID is the unique content-addressed or randomly generated UUID of this envelope.
	ID string

	// Source identifies the bidding or publishing cognitive ability (e.g., "CognitiveAbility.Reasoning").
	Source string

	// Topic specifies the leveled workspace channel this envelope belongs to.
	Topic TopicID

	// ParentRef references the message ID this envelope responds to (if any).
	ParentRef string

	// PayloadRef is the immutable content-addressed storage URI in idun/core/storage.
	// Executive Functions MUST NEVER inspect or parse the contents of this URI.
	PayloadRef string

	// PayloadModality identifies format hint ("text", "structured-frame", "vector-ref").
	PayloadModality string

	// RawConfidence reports the publishing module's self-assessed epistemic certainty [0.0, 1.0].
	RawConfidence float64

	// Urgency indicates emergency or safety priority override [0 = normal, 100 = critical safety].
	Urgency int

	// CostEstimateUnits reports the estimated computational cost units required downstream.
	CostEstimateUnits int

	// ExecutionDuration records the wall-clock duration taken to formulate this bid/envelope.
	ExecutionDuration time.Duration

	// CreatedAt records UTC creation timestamp.
	CreatedAt time.Time
}

// Validate verifies that the Envelope satisfies all Version 2.0 control-plane invariants.
func (e Envelope) Validate() error {
	if e.ID == "" {
		return ErrMissingID
	}
	if e.Source == "" {
		return ErrMissingSource
	}
	if !e.Topic.IsValid() {
		return ErrInvalidTopic
	}
	if e.PayloadRef == "" {
		return ErrMissingPayloadRef
	}
	if e.RawConfidence < 0.0 || e.RawConfidence > 1.0 {
		return ErrInvalidConfidence
	}
	if e.Urgency < 0 || e.Urgency > 100 {
		return ErrInvalidUrgency
	}
	if e.CostEstimateUnits < 0 {
		return ErrNegativeCost
	}
	return nil
}

// EffectivePriority computes the Calibrated Effective Priority (P_eff) for Executive arbitration.
//
// Formula:
//
//	P_eff = (RawConfidence * calibrationWeight) + (Urgency * alpha) - ((Cost / TotalBudget) * beta)
//
// Where calibrationWeight is supplied by the Epistemic Calibration System.
func (e Envelope) EffectivePriority(calibrationWeight, alpha, beta float64, totalBudget int) float64 {
	if calibrationWeight <= 0 {
		calibrationWeight = 1.0
	}
	costRatio := 0.0
	if totalBudget > 0 {
		costRatio = float64(e.CostEstimateUnits) / float64(totalBudget)
	}
	return (e.RawConfidence * calibrationWeight) + (float64(e.Urgency) * alpha) - (costRatio * beta)
}

// EnvelopeBuilder provides a fluent, thread-safe helper for constructing validated Envelopes.
type EnvelopeBuilder struct {
	env Envelope
}

// NewEnvelopeBuilder creates a new builder with auto-generated ID and current UTC timestamp.
func NewEnvelopeBuilder() *EnvelopeBuilder {
	return &EnvelopeBuilder{
		env: Envelope{
			ID:        generateID(),
			CreatedAt: time.Now().UTC(),
		},
	}
}

// WithID overrides the auto-generated envelope ID.
func (b *EnvelopeBuilder) WithID(id string) *EnvelopeBuilder {
	b.env.ID = id
	return b
}

// WithSource sets the publishing module source identifier.
func (b *EnvelopeBuilder) WithSource(source string) *EnvelopeBuilder {
	b.env.Source = source
	return b
}

// WithTopic sets the leveled workspace TopicID.
func (b *EnvelopeBuilder) WithTopic(topic TopicID) *EnvelopeBuilder {
	b.env.Topic = topic
	return b
}

// WithParentRef sets the optional parent envelope reference ID.
func (b *EnvelopeBuilder) WithParentRef(parentRef string) *EnvelopeBuilder {
	b.env.ParentRef = parentRef
	return b
}

// WithPayloadRef sets the content-addressed storage URI of the domain payload.
func (b *EnvelopeBuilder) WithPayloadRef(payloadRef string) *EnvelopeBuilder {
	b.env.PayloadRef = payloadRef
	return b
}

// WithModality sets the payload modality hint.
func (b *EnvelopeBuilder) WithModality(modality string) *EnvelopeBuilder {
	b.env.PayloadModality = modality
	return b
}

// WithConfidence sets the self-reported raw confidence [0.0, 1.0].
func (b *EnvelopeBuilder) WithConfidence(confidence float64) *EnvelopeBuilder {
	b.env.RawConfidence = confidence
	return b
}

// WithUrgency sets the priority/safety urgency level [0, 100].
func (b *EnvelopeBuilder) WithUrgency(urgency int) *EnvelopeBuilder {
	b.env.Urgency = urgency
	return b
}

// WithCostEstimate sets estimated computational cost units.
func (b *EnvelopeBuilder) WithCostEstimate(units int) *EnvelopeBuilder {
	b.env.CostEstimateUnits = units
	return b
}

// WithExecutionDuration sets the time taken to formulate this bid.
func (b *EnvelopeBuilder) WithExecutionDuration(d time.Duration) *EnvelopeBuilder {
	b.env.ExecutionDuration = d
	return b
}

// Build validates and returns the immutable Envelope struct.
func (b *EnvelopeBuilder) Build() (Envelope, error) {
	if err := b.env.Validate(); err != nil {
		return Envelope{}, fmt.Errorf("build envelope: %w", err)
	}
	return b.env, nil
}

// generateID returns a secure random 16-byte hex ID.
func generateID() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}
