package capabilities

import (
	"time"
)

// CapabilityCategory defines the authorized World surface taxonomies.
type CapabilityCategory string

const (
	CategorySystem           CapabilityCategory = "System"
	CategoryFiles            CapabilityCategory = "Files"
	CategoryDevicesSensors   CapabilityCategory = "Devices & Sensors" // Includes Location
	CategoryCommunication    CapabilityCategory = "Communication"
	CategoryMedia            CapabilityCategory = "Media"
	CategoryNetwork          CapabilityCategory = "Network"
	CategoryExternalServices CapabilityCategory = "External Services" // Reserved for Plugin/Cloud Capabilities
	CategoryAutomation       CapabilityCategory = "Automation"
)

// LifecycleState defines the strict lifecycle states for any capability.
type LifecycleState string

const (
	LifecycleDiscovered  LifecycleState = "DISCOVERED"
	LifecycleRegistered  LifecycleState = "REGISTERED"
	LifecycleLoaded      LifecycleState = "LOADED"
	LifecycleInitialized LifecycleState = "INITIALIZED"
	LifecycleHealthy     LifecycleState = "HEALTHY"
	LifecycleExecuting   LifecycleState = "EXECUTING"
	LifecycleIdle        LifecycleState = "IDLE"
	LifecycleDisabled    LifecycleState = "DISABLED"
	LifecycleUnloaded    LifecycleState = "UNLOADED"
)

// OperationalStatus provides granular health visibility.
type OperationalStatus string

const (
	StatusHealthy     OperationalStatus = "HEALTHY"
	StatusBusy        OperationalStatus = "BUSY"
	StatusUnavailable OperationalStatus = "UNAVAILABLE"
	StatusDegraded    OperationalStatus = "DEGRADED"
	StatusDisabled    OperationalStatus = "DISABLED"
)

// CapabilityMetadata contains immutable definitions of a capability.
type CapabilityMetadata struct {
	Name                    string
	Category                CapabilityCategory
	Description             string
	Version                 string
	APIVersion              string
	MinimumFrameworkVersion string
	SupportedPlatforms      []string
	PermissionCategory      string
	ImplementationType      string
	Author                  string
	Tags                    []string
}

// CapabilityState contains mutable status information.
type CapabilityState struct {
	Lifecycle   LifecycleState
	Operational OperationalStatus
	LastUpdated time.Time
}

// CapabilityRequest encapsulates the payload passed to a capability for execution.
type CapabilityRequest struct {
	RequirementID string
	Parameters    map[string]string
	ContextID     string
}

// RealizationStrategy defines how the Response Type Router should realize the result.
type RealizationStrategy int

const (
	Deterministic RealizationStrategy = iota
	Generative
)

// CapabilityResult is the normalized, semantically neutral output from execution.
type CapabilityResult struct {
	RequirementID string                 `json:"requirement_id"`
	Success       bool                   `json:"success"`
	Realization   RealizationStrategy    `json:"realization"`
	ResponseType  string                 `json:"response_type,omitempty"`
	Data          map[string]interface{} `json:"data"`
	Error         *CapabilityError       `json:"error,omitempty"`
	Duration      time.Duration          `json:"duration"`
}

// CapabilityError represents standardized failure modes without polluting semantic understanding.
type CapabilityError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Retry   bool   `json:"retry"`
}
