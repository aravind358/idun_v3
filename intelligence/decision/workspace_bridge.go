package decision

import (
	"context"
	"fmt"

	"idun/intelligence/communication"
)

// ShouldPublishToWorkspace determines whether a decision record should be broadcast
// over the Global Workspace communication substrate.
// Per Section 4.1 of Version 2.0.0-FROZEN:
// - Reflexive micro-decisions SKIP workspace broadcast to prevent bus saturation and deadlock.
// - Deliberative macro-decisions MUST be published to TopicEvaluatedOptions.
func ShouldPublishToWorkspace(depth DeliberationDepth) bool {
	return depth == DepthDeliberative
}

// WorkspacePublisher defines the functional interface required to emit Envelopes to the Global Workspace.
type WorkspacePublisher interface {
	Publish(ctx context.Context, env communication.Envelope) error
}

// WorkspaceSubscription manages the lifecycle of an active subscription.
type WorkspaceSubscription interface {
	Cancel() error
}

// WorkspaceSubscriber defines the interface for registering handlers on Workspace topics.
type WorkspaceSubscriber interface {
	Subscribe(topic communication.TopicID, subscriberID string, handler func(ctx context.Context, env communication.Envelope) error) (WorkspaceSubscription, error)
}

// PayloadStorer defines the functional interface required to persist DecisionRecord payloads to CAS storage.
type PayloadStorer interface {
	Store(ctx context.Context, data []byte) (string, error)
	Retrieve(ctx context.Context, key string) ([]byte, error)
}

// PublishDeliberativeDecision serializes a Deliberative DecisionRecord, stores its payload reference,
// packages it into a canonical communication.Envelope, and publishes it to TopicEvaluatedOptions.
func PublishDeliberativeDecision(
	ctx context.Context,
	rec *DecisionRecord,
	storer PayloadStorer,
	publisher WorkspacePublisher,
	parentRefs ...string,
) (communication.Envelope, error) {
	if rec == nil {
		return communication.Envelope{}, ErrInvalidDecisionRecord
	}
	if err := rec.Validate(); err != nil {
		return communication.Envelope{}, fmt.Errorf("decision validation firewall rejected record: %w", err)
	}
	if !ShouldPublishToWorkspace(rec.DeliberationDepth) {
		return communication.Envelope{}, fmt.Errorf("decision: reflexive decision records must not be published to Global Workspace")
	}
	if storer == nil || publisher == nil {
		return communication.Envelope{}, fmt.Errorf("decision: storer and publisher cannot be nil")
	}

	data, err := MarshalDecisionRecord(rec)
	if err != nil {
		return communication.Envelope{}, err
	}

	payloadRef, err := storer.Store(ctx, data)
	if err != nil {
		return communication.Envelope{}, fmt.Errorf("decision: failed to store payload: %w", err)
	}

	env, err := EnvelopeFromDecisionRecord(rec, payloadRef)
	if err != nil {
		return communication.Envelope{}, err
	}

	if len(parentRefs) > 0 && parentRefs[0] != "" {
		env.ParentRef = parentRefs[0]
	}

	if err := publisher.Publish(ctx, env); err != nil {
		return communication.Envelope{}, fmt.Errorf("decision: failed to publish decision record to workspace: %w", err)
	}

	return env, nil
}
