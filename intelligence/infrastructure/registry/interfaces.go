// Package registry implements the Model & Capability Registry for IDUN Intelligence Infrastructure.
//
// Architecture Version: 1.0.0-FROZEN-SPRINT1
//
// The Registry provides the single source of truth mapping stable, semantic
// Logical Model Identifiers (ModelID) known to cognitive abilities to physical
// Backend Descriptors (BackendDescriptor) managed by infrastructure and hardware.
//
// Architectural Invariants:
//   - Interface Segregation: Cognitive abilities and infrastructure execution services
//     receive only the read-only Resolver interface. Only administrative kernel wiring
//     receives the ModelRegistry mutation interface.
//   - Open Extensibility: Physical execution targets use open driver URI schemes
//     (DriverScheme) and opaque configuration maps (DriverConfig).
//   - Zero Cognitive Policy: The registry stores and resolves descriptors; it never
//     interprets cognitive intent or executes inference.
package registry

import (
	"context"
	"errors"
	"time"
)

// Sentinel errors returned by Registry methods.
var (
	ErrModelNotFound      = errors.New("registry: model ID not registered")
	ErrBackendUnavailable = errors.New("registry: backend currently unavailable or unhealthy")
	ErrRegistryClosed     = errors.New("registry: service closed")
	ErrInvalidDescriptor  = errors.New("registry: invalid backend descriptor")
	ErrVersionNotFound    = errors.New("registry: target rollback version not found")
)

// ModelID is a stable, logical identifier representing a computational capability.
// Examples: "language-reasoner", "semantic-embedder", "multimodal-perceiver".
type ModelID string

// BackendHealth indicates the runtime health status of a registered backend.
type BackendHealth string

const (
	// HealthHealthy indicates the backend is ready to accept inference requests.
	HealthHealthy BackendHealth = "HEALTHY"
	// HealthDegraded indicates the backend is operational but experiencing degraded performance.
	HealthDegraded BackendHealth = "DEGRADED"
	// HealthUnhealthy indicates the backend is failing or unreachable.
	HealthUnhealthy BackendHealth = "UNHEALTHY"
)

// BackendDescriptor defines the physical execution target for a logical ModelID.
// It uses an open driver scheme and configuration map for 20+ year hardware extensibility.
type BackendDescriptor struct {
	// ID is the unique physical backend identifier (e.g., "llama-server-01").
	ID string
	// DriverScheme specifies the open execution protocol scheme (e.g., "grpc", "local-bin", "neuromorphic-pci").
	DriverScheme string
	// Endpoint specifies the binary path, network URI, socket, or hardware device node.
	Endpoint string
	// Version specifies the model or engine semantic version.
	Version string
	// MaxConcurrency specifies the maximum concurrent execution streams permitted.
	MaxConcurrency int
	// SupportedBudgets lists supported execution budget tiers ("REFLEXIVE", "STANDARD", "DELIBERATIVE").
	SupportedBudgets []string
	// DriverConfig contains opaque driver-specific configuration parameters.
	DriverConfig map[string]string
	// Health represents the current runtime health status.
	Health BackendHealth
	// HealthReason describes the cause of degraded or unhealthy status.
	HealthReason string
	// RegisteredAt records when this descriptor version was registered.
	RegisteredAt time.Time
}

// Clone returns a deep copy of the BackendDescriptor to prevent data races.
func (bd BackendDescriptor) Clone() BackendDescriptor {
	copyBD := bd
	if bd.SupportedBudgets != nil {
		copyBD.SupportedBudgets = make([]string, len(bd.SupportedBudgets))
		copy(copyBD.SupportedBudgets, bd.SupportedBudgets)
	}
	if bd.DriverConfig != nil {
		copyBD.DriverConfig = make(map[string]string, len(bd.DriverConfig))
		for k, v := range bd.DriverConfig {
			copyBD.DriverConfig[k] = v
		}
	}
	return copyBD
}

// Resolver is the read-only capability interface injected into Inference and Embedding services.
// Cognitive abilities and infrastructure execution engines interact with the registry via Resolver.
type Resolver interface {
	// Resolve returns the active, healthy BackendDescriptor for a logical ModelID.
	// Returns ErrModelNotFound if unregistered or ErrBackendUnavailable if unhealthy.
	Resolve(ctx context.Context, modelID ModelID) (BackendDescriptor, error)
}

// TelemetrySnapshot captures operational health metrics for Host/Kernel monitoring.
type TelemetrySnapshot struct {
	TotalRegisteredModels int
	HealthyModels         int
	UnhealthyModels       int
	TotalResolutions      int64
	FailedResolutions     int64
}

// TelemetryProvider exposes read-only operational telemetry to the Host/Kernel.
// Executive Functions and Cognitive Abilities MUST NEVER import or use TelemetryProvider.
type TelemetryProvider interface {
	GetTelemetry() TelemetrySnapshot
}

// ModelRegistry defines the administrative lifecycle and mutation interface for the registry.
type ModelRegistry interface {
	Resolver
	TelemetryProvider

	// Register binds or updates a logical ModelID to a physical BackendDescriptor atomically.
	// Previous registrations for the same ModelID are preserved in rollback history.
	Register(ctx context.Context, modelID ModelID, backend BackendDescriptor) error

	// Deregister removes the active registration and history for a logical ModelID.
	Deregister(ctx context.Context, modelID ModelID) error

	// Rollback atomically reverts the active backend descriptor for modelID to a previous version.
	Rollback(ctx context.Context, modelID ModelID, version string) error

	// SetHealth updates the runtime health status of a registered model ID.
	SetHealth(ctx context.Context, modelID ModelID, health BackendHealth, reason string) error

	// ListModels returns a snapshot of all currently active logical ModelIDs and descriptors.
	ListModels() map[ModelID]BackendDescriptor

	// ListVersions returns all historical versions registered for a logical ModelID.
	ListVersions(modelID ModelID) []BackendDescriptor

	// Name returns the canonical component name ("Intelligence.Infrastructure.Registry").
	Name() string
	// Start boots the registry lifecycle.
	Start() error
	// Close gracefully shuts down the registry.
	Close() error
}
