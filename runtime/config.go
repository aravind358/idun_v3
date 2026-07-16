package runtime

import (
	"errors"
)

// RuntimeConfiguration holds runtime-level infrastructure parameters.
// Subsystem-specific policies and hyperparameters remain encapsulated within
// each respective subsystem's local configuration or builder options.
type RuntimeConfiguration struct {
	// RuntimeVersion identifies the target architecture release (e.g., "2.0.0-FROZEN").
	RuntimeVersion string `json:"runtime_version"`

	// StoragePath defines the root file path for persistent records and snapshots.
	StoragePath string `json:"storage_path"`

	// EnableLogging determines whether stdout/stderr structured logging is active.
	EnableLogging bool `json:"enable_logging"`

	// DebugMode activates verbose diagnostic logging and strict boundary checking.
	DebugMode bool `json:"debug_mode"`

	// ReplayMode configures the runtime for deterministic replay simulation.
	ReplayMode bool `json:"replay_mode"`

	// InitialExecutiveBudget defines the starting execution budget allocated to Executive coordinator.
	InitialExecutiveBudget int `json:"initial_executive_budget"`

	// EnabledSubsystems enumerates explicit feature flags for optional cognitive modules.
	// If empty or nil, all standard Layer 1 cognitive subsystems are enabled by default.
	EnabledSubsystems map[string]bool `json:"enabled_subsystems"`
}

// DefaultConfiguration returns safe, production-grade default settings.
func DefaultConfiguration() RuntimeConfiguration {
	return RuntimeConfiguration{
		RuntimeVersion:         "2.0.0-FROZEN",
		StoragePath:            "./data/runtime",
		EnableLogging:          true,
		DebugMode:              false,
		ReplayMode:             false,
		InitialExecutiveBudget: 10000,
		EnabledSubsystems: map[string]bool{
			"understanding": true,
			"reasoning":     true,
			"planning":      true,
			"decision":      true,
			"reflection":    true,
			"learning":      true,
			"attention":     true,
			"executive":     true,
			"world":         true,
		},
	}
}

// Validate verifies runtime configuration parameters.
func (c *RuntimeConfiguration) Validate() error {
	if c.RuntimeVersion == "" {
		return errors.New("runtime: empty runtime version")
	}
	if c.StoragePath == "" {
		return errors.New("runtime: empty storage path")
	}
	if c.InitialExecutiveBudget < 0 {
		return errors.New("runtime: negative initial executive budget")
	}
	return nil
}
