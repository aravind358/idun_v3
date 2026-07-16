// Package executive implements IDUN's Intelligence Pillar Executive Functions.
//
// Architecture Version: 2.0.0-FROZEN
package executive

import (
	"context"

	"idun/intelligence/communication"
)

// ExecutivePayloadStorer defines the interface for persisting and retrieving
// payload blobs through the content-addressable storage layer.
// Executive Functions uses this only to retrieve Decision artifacts by PayloadRef —
// it never constructs, interprets, or transforms cognitive payloads.
type ExecutivePayloadStorer interface {
	// Retrieve fetches a previously stored payload by its CAS key.
	Retrieve(ctx context.Context, key string) ([]byte, error)
}

// ExecutiveWorkspaceSubscription manages the lifecycle of one active workspace subscription.
type ExecutiveWorkspaceSubscription interface {
	// Cancel cancels the subscription and stops delivery.
	Cancel() error
}

// ExecutiveWorkspaceSubscriber defines the interface for subscribing to leveled
// workspace topic channels. Executive uses this to receive TopicEvaluatedOptions events.
type ExecutiveWorkspaceSubscriber interface {
	Subscribe(
		topic communication.TopicID,
		subscriberID string,
		handler func(ctx context.Context, env communication.Envelope) error,
	) (ExecutiveWorkspaceSubscription, error)
}
