package planning

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"idun/intelligence/communication"
)

// Sentinel errors for workspace publishing.
var (
	ErrInvalidPlan           = errors.New("planning: invalid Plan")
	ErrInvalidPlanningTrace  = errors.New("planning: invalid PlanningTrace")
	ErrInvalidPlanningResult = errors.New("planning: invalid PlanningResult")
)

// ShouldPublishToWorkspace determines whether a planning artifact should be broadcast
// over the Global Workspace communication substrate.
// - Reflexive plans (<10ms) skip broadcast unless explicitly escalated.
// - Tactical and Deliberative plans MUST be published to TopicCandidatePlans.
func ShouldPublishToWorkspace(depth PlanningDepth) bool {
	return depth == DepthTactical || depth == DepthStrategic
}

// WorkspacePublisher defines the functional interface required to emit Envelopes to the Global Workspace.
type WorkspacePublisher interface {
	Publish(ctx context.Context, env communication.Envelope) error
}

// PayloadStorer defines the functional interface required to persist payloads to CAS storage.
type PayloadStorer interface {
	Store(ctx context.Context, data []byte) (string, error)
}

// EnvelopeFromPlan packages a completed Plan into a canonical Global Workspace Envelope
// targeted at TopicCandidatePlans (`candidate-plans`).
func EnvelopeFromPlan(plan *Plan, payloadRef string) (communication.Envelope, error) {
	if plan == nil {
		return communication.Envelope{}, ErrInvalidPlan
	}
	if payloadRef == "" {
		return communication.Envelope{}, fmt.Errorf("planning: payloadRef cannot be empty when publishing Plan")
	}

	urgency := 0
	if plan.Status == PlanStatusConstraintConflict || plan.Status == PlanStatusInfeasible {
		urgency = 60
	}

	env := communication.Envelope{
		ID:                fmt.Sprintf("env-plan-%s", plan.PlanID),
		Source:            "CognitiveAbility.Planning",
		Topic:             communication.TopicCandidatePlans,
		PayloadRef:        payloadRef,
		PayloadModality:   "structured-frame",
		RawConfidence:     plan.ConfidenceProfile.OverallConfidence,
		Urgency:           urgency,
		CostEstimateUnits: int(plan.EstimatedCost),
		CreatedAt:         time.Now().UTC(),
	}

	if err := env.Validate(); err != nil {
		return communication.Envelope{}, err
	}
	return env, nil
}

// EnvelopeFromPlanningTrace packages a diagnostic PlanningTrace into a canonical Envelope
// targeted at TopicReflections (`reflections`) for metacognitive consumption.
func EnvelopeFromPlanningTrace(trace *PlanningTrace, payloadRef string) (communication.Envelope, error) {
	if trace == nil {
		return communication.Envelope{}, ErrInvalidPlanningTrace
	}
	if payloadRef == "" {
		return communication.Envelope{}, fmt.Errorf("planning: payloadRef cannot be empty when publishing PlanningTrace")
	}

	env := communication.Envelope{
		ID:                fmt.Sprintf("env-trace-%s", trace.TraceID),
		Source:            "CognitiveAbility.Planning",
		Topic:             communication.TopicReflections,
		PayloadRef:        payloadRef,
		PayloadModality:   "structured-frame",
		RawConfidence:     trace.ConfidenceProfile.OverallConfidence,
		Urgency:           0,
		CostEstimateUnits: 5,
		CreatedAt:         time.Now().UTC(),
	}

	if err := env.Validate(); err != nil {
		return communication.Envelope{}, err
	}
	return env, nil
}

// EnvelopeFromPlanningResult packages an aggregate PlanningResult into a canonical Envelope.
func EnvelopeFromPlanningResult(result *PlanningResult, payloadRef string) (communication.Envelope, error) {
	if result == nil {
		return communication.Envelope{}, ErrInvalidPlanningResult
	}
	if payloadRef == "" {
		return communication.Envelope{}, fmt.Errorf("planning: payloadRef cannot be empty when publishing PlanningResult")
	}

	maxConf := 0.0
	for _, p := range result.Plans {
		if p != nil && p.ConfidenceProfile.OverallConfidence > maxConf {
			maxConf = p.ConfidenceProfile.OverallConfidence
		}
	}

	urgency := 0
	if result.ResultStatus == ResultEscalationRecommended {
		urgency = 50
	}

	env := communication.Envelope{
		ID:                fmt.Sprintf("env-res-%s", result.ResultID),
		Source:            "CognitiveAbility.Planning",
		Topic:             communication.TopicCandidatePlans,
		PayloadRef:        payloadRef,
		PayloadModality:   "structured-frame",
		RawConfidence:     maxConf,
		Urgency:           urgency,
		CostEstimateUnits: 15,
		CreatedAt:         time.Now().UTC(),
	}

	if err := env.Validate(); err != nil {
		return communication.Envelope{}, err
	}
	return env, nil
}

// PublishPlan validates, serializes, stores, and publishes a Plan to the Global Workspace.
// Enforces Validation Firewall: an invalid Plan is rejected prior to storage or broadcast.
func PublishPlan(
	ctx context.Context,
	plan *Plan,
	storer PayloadStorer,
	publisher WorkspacePublisher,
) (communication.Envelope, error) {
	if plan == nil {
		return communication.Envelope{}, ErrInvalidPlan
	}
	if err := plan.Validate(); err != nil {
		return communication.Envelope{}, fmt.Errorf("planning validation firewall rejected Plan: %w", err)
	}
	if storer == nil || publisher == nil {
		return communication.Envelope{}, errors.New("planning: storer and publisher cannot be nil")
	}

	data, err := MarshalPlan(plan)
	if err != nil {
		return communication.Envelope{}, err
	}

	payloadRef, err := storer.Store(ctx, data)
	if err != nil {
		return communication.Envelope{}, fmt.Errorf("planning: failed to store payload: %w", err)
	}

	env, err := EnvelopeFromPlan(plan, payloadRef)
	if err != nil {
		return communication.Envelope{}, err
	}

	if err := publisher.Publish(ctx, env); err != nil {
		return communication.Envelope{}, fmt.Errorf("planning: failed to publish envelope to workspace: %w", err)
	}

	return env, nil
}

// PublishPlanningTrace validates, serializes, stores, and publishes a diagnostic PlanningTrace.
func PublishPlanningTrace(
	ctx context.Context,
	trace *PlanningTrace,
	storer PayloadStorer,
	publisher WorkspacePublisher,
) (communication.Envelope, error) {
	if trace == nil {
		return communication.Envelope{}, ErrInvalidPlanningTrace
	}
	if err := trace.Validate(); err != nil {
		return communication.Envelope{}, fmt.Errorf("planning validation firewall rejected PlanningTrace: %w", err)
	}
	if storer == nil || publisher == nil {
		return communication.Envelope{}, errors.New("planning: storer and publisher cannot be nil")
	}

	data, err := MarshalPlanningTrace(trace)
	if err != nil {
		return communication.Envelope{}, err
	}

	payloadRef, err := storer.Store(ctx, data)
	if err != nil {
		return communication.Envelope{}, fmt.Errorf("planning: failed to store trace payload: %w", err)
	}

	env, err := EnvelopeFromPlanningTrace(trace, payloadRef)
	if err != nil {
		return communication.Envelope{}, err
	}

	if err := publisher.Publish(ctx, env); err != nil {
		return communication.Envelope{}, fmt.Errorf("planning: failed to publish trace envelope: %w", err)
	}

	return env, nil
}

// PublishPlanningResult validates, serializes, stores, and publishes an aggregate PlanningResult.
func PublishPlanningResult(
	ctx context.Context,
	result *PlanningResult,
	storer PayloadStorer,
	publisher WorkspacePublisher,
) (communication.Envelope, error) {
	if result == nil {
		return communication.Envelope{}, ErrInvalidPlanningResult
	}
	if err := result.Validate(); err != nil {
		return communication.Envelope{}, fmt.Errorf("planning validation firewall rejected PlanningResult: %w", err)
	}
	if storer == nil || publisher == nil {
		return communication.Envelope{}, errors.New("planning: storer and publisher cannot be nil")
	}

	data, err := MarshalPlanningResult(result)
	if err != nil {
		return communication.Envelope{}, err
	}

	payloadRef, err := storer.Store(ctx, data)
	if err != nil {
		return communication.Envelope{}, fmt.Errorf("planning: failed to store result payload: %w", err)
	}

	env, err := EnvelopeFromPlanningResult(result, payloadRef)
	if err != nil {
		return communication.Envelope{}, err
	}

	if err := publisher.Publish(ctx, env); err != nil {
		return communication.Envelope{}, fmt.Errorf("planning: failed to publish result envelope: %w", err)
	}

	return env, nil
}

// Marshal & Unmarshal Helpers

func MarshalPlan(p *Plan) ([]byte, error) {
	if p == nil {
		return nil, ErrInvalidPlan
	}
	return json.Marshal(p)
}

func UnmarshalPlan(data []byte) (*Plan, error) {
	if len(data) == 0 {
		return nil, ErrInvalidPlan
	}
	var p Plan
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("planning: failed to unmarshal Plan: %w", err)
	}
	return &p, nil
}

func MarshalPlanningTrace(t *PlanningTrace) ([]byte, error) {
	if t == nil {
		return nil, ErrInvalidPlanningTrace
	}
	return json.Marshal(t)
}

func UnmarshalPlanningTrace(data []byte) (*PlanningTrace, error) {
	if len(data) == 0 {
		return nil, ErrInvalidPlanningTrace
	}
	var t PlanningTrace
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, fmt.Errorf("planning: failed to unmarshal PlanningTrace: %w", err)
	}
	return &t, nil
}

func MarshalPlanningResult(r *PlanningResult) ([]byte, error) {
	if r == nil {
		return nil, ErrInvalidPlanningResult
	}
	return json.Marshal(r)
}

func UnmarshalPlanningResult(data []byte) (*PlanningResult, error) {
	if len(data) == 0 {
		return nil, ErrInvalidPlanningResult
	}
	var r PlanningResult
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("planning: failed to unmarshal PlanningResult: %w", err)
	}
	return &r, nil
}
