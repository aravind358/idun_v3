package world

import (
	"context"
)

// ============================================================================
// Adapter Identity — Refinement 10
// ============================================================================

// InputAdapter defines the narrow contract for receiving external input and converting
// it to an Interaction. All adapter implementations must carry immutable identity
// information to enable exact replay when adapter implementations evolve over time.
type InputAdapter interface {
	// Receive blocks until a new Interaction is available from the external source,
	// then returns it. The returned Interaction is already normalized.
	// A context cancellation signals the adapter to stop waiting and return an error.
	Receive(ctx context.Context) (*Interaction, error)

	// Name returns the human-readable name of this adapter (e.g., "TextInputAdapter").
	Name() string

	// AdapterVersion returns the implementation version of this adapter.
	// Used in WorldTrace for deterministic replay provenance.
	AdapterVersion() string

	// AdapterFingerprint returns a deterministic SHA-256 identity digest of the adapter.
	// Derived from Name and Version; changes when the adapter implementation changes.
	AdapterFingerprint() string

	// Close releases any resources held by this adapter.
	Close() error
}

// OutputAdapter defines the narrow contract for presenting a Response to the external world.
// All adapter implementations must carry immutable identity information.
type OutputAdapter interface {
	// Send presents the Response to the external world through this adapter's channel.
	// It must return an error if delivery fails. It must not block indefinitely.
	Send(ctx context.Context, response *Response) error

	// Name returns the human-readable name of this adapter (e.g., "TextOutputAdapter").
	Name() string

	// AdapterVersion returns the implementation version of this adapter.
	AdapterVersion() string

	// AdapterFingerprint returns a deterministic SHA-256 identity digest of the adapter.
	AdapterFingerprint() string

	// Close releases any resources held by this adapter.
	Close() error
}

// ============================================================================
// PayloadStorer — narrow bridge to Core.Storage
// ============================================================================

// PayloadStorer defines the narrow interface for persisting large interaction
// payloads to content-addressed storage. World only requires this single capability;
// it does not import or depend on the full storage package.
type PayloadStorer interface {
	// Store persists data and returns its content-addressed key (SHA-256 hex).
	Store(ctx context.Context, data []byte) (string, error)
	// Retrieve fetches data by its content-addressed key.
	Retrieve(ctx context.Context, key string) ([]byte, error)
}

// ============================================================================
// WorldPublisher — narrow bridge to Global Workspace
// ============================================================================

// WorldPublisher defines the narrow interface for publishing envelopes to the
// Global Workspace. World only requires this one capability from Workspace.
type WorldPublisher interface {
	// Publish posts a communication envelope to the Global Workspace.
	Publish(ctx context.Context, env interface{ GetID() string }, opts ...interface{}) error
}

// ============================================================================
// WorldService
// ============================================================================

// WorldService defines the full canonical public interface for the World subsystem.
// World is the boundary between IDUN and the external world.
type WorldService interface {
	// Start boots the World service lifecycle and wires Workspace subscriptions.
	Start(ctx context.Context) error

	// Close gracefully shuts down the World service and its adapters.
	Close() error

	// Name returns the canonical Kernel component name ("World.Service").
	Name() string

	// GetPolicyProfile returns the immutable WorldPolicyProfile governing this service.
	GetPolicyProfile() *WorldPolicyProfile

	// GetCapabilities returns the immutable WorldCapabilities for this deployment.
	GetCapabilities() *WorldCapabilities

	// GetSummary returns a bounded statistical snapshot of World interaction telemetry.
	GetSummary() WorldSummary

	// HandleInteraction publishes an Interaction to the Global Workspace and returns
	// immediately. World is fully event-driven; it does not block waiting for a response.
	HandleInteraction(ctx context.Context, interaction *Interaction) error
}
