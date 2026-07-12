// Package workspace implements the Global Workspace & Leveled Blackboard Engine
// for IDUN V3 Cognitive Communication Architecture Version 2.0.
//
// Architecture Version: 2.0.0-FROZEN
//
// The Workspace provides leveled topic channels where cognitive abilities
// publish and subscribe as symmetric peers, without centralized routing or
// semantic inspection by Executive Functions.
package workspace

import (
	"context"
	"errors"

	"idun/intelligence/communication"
)

// Sentinel errors returned by Workspace operations.
var (
	ErrWorkspaceClosed     = errors.New("workspace: engine is closed")
	ErrInvalidEnvelope     = errors.New("workspace: envelope validation failed")
	ErrSubscriptionClosed  = errors.New("workspace: subscription is closed or invalid")
	ErrSubscriberIDMissing = errors.New("workspace: subscriber ID is required")
	ErrHandlerMissing      = errors.New("workspace: envelope handler is required")
	ErrInvalidTopic        = errors.New("workspace: topic is invalid or unregistered")
)

// SubscriptionID uniquely identifies an active workspace subscription.
type SubscriptionID string

// EnvelopeHandler defines the callback invoked when a matching Envelope is published.
//
// Handlers MUST execute quickly and non-blocking. Long-running cognitive operations
// MUST be scheduled asynchronously.
type EnvelopeHandler func(ctx context.Context, env communication.Envelope) error

// Subscription represents an active subscription on the Global Workspace.
type Subscription interface {
	// ID returns the unique SubscriptionID.
	ID() SubscriptionID

	// Topic returns the subscribed TopicID.
	Topic() communication.TopicID

	// Subscriber returns the subscribing module identifier.
	Subscriber() string

	// Cancel cancels the subscription and stops delivery.
	Cancel() error
}

// PublishOption configures functional options for envelope publication.
type PublishOption func(*publishConfig)

type publishConfig struct {
	broadcastGlobal bool
}

// WithGlobalBroadcast escalates the publication to all subscribers across all topics.
// Reserved for critical constitutional overrides, safety halts, or high-urgency impasses.
func WithGlobalBroadcast(global bool) PublishOption {
	return func(c *publishConfig) {
		c.broadcastGlobal = global
	}
}

// Workspace defines the capability contract for the Global Workspace & Leveled Blackboard Engine.
type Workspace interface {
	// Publish validates and posts an Envelope to its leveled topic channel (or globally if escalated).
	Publish(ctx context.Context, env communication.Envelope, opts ...PublishOption) error

	// Subscribe subscribes a cognitive ability to a specific leveled topic channel.
	Subscribe(topic communication.TopicID, subscriberID string, handler EnvelopeHandler) (Subscription, error)

	// SubscribeAll subscribes a monitor (Executive, Reflection, or Value) to all canonical topics.
	SubscribeAll(subscriberID string, handler EnvelopeHandler) ([]Subscription, error)

	// Unsubscribe cancels an active subscription by ID.
	Unsubscribe(id SubscriptionID) error

	// GetEnvelope retrieves a recently published envelope by its unique ID.
	GetEnvelope(id string) (communication.Envelope, bool)

	// ListTopicEnvelopes retrieves up to limit recent envelopes buffered on a specific topic channel.
	ListTopicEnvelopes(topic communication.TopicID, limit int) []communication.Envelope

	// Name returns the canonical Kernel component name ("Intelligence.Workspace").
	Name() string

	// Start boots the Workspace engine.
	Start() error

	// Close gracefully shuts down the Workspace and cancels all active subscriptions.
	Close() error
}
