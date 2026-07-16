package executive

import (
	"fmt"
	"time"
)

// DefaultExecutivePolicyProfile constructs and returns a fully validated, default
// immutable ExecutivePolicyProfile.
func DefaultExecutivePolicyProfile() *ExecutivePolicyProfile {
	profile, err := NewExecutivePolicyProfileBuilder().
		WithProfileID("default-profile-v2").
		WithVersion("1.0.0").
		WithSource("SystemBoot").
		Build()
	if err != nil {
		panic(fmt.Sprintf("executive: failed to build default policy profile: %v", err))
	}
	return profile
}

// DefaultExecutiveCapabilities constructs and returns fully validated, default
// deployment capabilities.
func DefaultExecutiveCapabilities() *ExecutiveCapabilities {
	caps, err := NewExecutiveCapabilitiesBuilder().Build()
	if err != nil {
		panic(fmt.Sprintf("executive: failed to build default capabilities: %v", err))
	}
	return caps
}

// Configuration defines the complete runtime configuration for Executive Version 2.0,
// grouping the V1 service parameters with the Phase 1 immutable policy profile and capabilities.
type Configuration struct {
	// ServiceConfig contains base V1 settings (Logger, DefaultMaxFuel, etc.).
	ServiceConfig Config `json:"service_config"`

	// Policy holds the immutable profile snapshot consumed by Executive.
	Policy *ExecutivePolicyProfile `json:"policy"`

	// Capabilities holds the immutable deployment boundaries of the engine.
	Capabilities *ExecutiveCapabilities `json:"capabilities"`

	// ExecutiveVersion records the version string of the coordinating engine.
	ExecutiveVersion string `json:"executive_version"`
}

// DefaultConfiguration returns a fully populated and validated default Configuration.
func DefaultConfiguration() Configuration {
	return Configuration{
		ServiceConfig: Config{
			IdleThreshold:        30 * time.Second,
			DefaultMaxFuel:       500,
			DefaultMaxReflection: 3,
		},
		Policy:           DefaultExecutivePolicyProfile(),
		Capabilities:     DefaultExecutiveCapabilities(),
		ExecutiveVersion: DefaultExecutiveVersion,
	}
}

// Validate implements the Validation Firewall for Configuration.
func (c *Configuration) Validate() error {
	if c == nil {
		return fmt.Errorf("%w: configuration object is nil", ErrInvalidConfig)
	}
	if c.ServiceConfig.IdleThreshold < 0 {
		return fmt.Errorf("%w: service_config.idle_threshold cannot be negative", ErrInvalidConfig)
	}
	if c.ServiceConfig.DefaultMaxFuel < 0 {
		return fmt.Errorf("%w: service_config.default_max_fuel cannot be negative", ErrInvalidConfig)
	}
	if c.ServiceConfig.DefaultMaxReflection < 0 {
		return fmt.Errorf("%w: service_config.default_max_reflection cannot be negative", ErrInvalidConfig)
	}
	if c.Policy == nil {
		return fmt.Errorf("%w: policy cannot be nil", ErrInvalidConfig)
	}
	if err := c.Policy.Validate(); err != nil {
		return fmt.Errorf("%w: invalid policy: %v", ErrInvalidConfig, err)
	}
	if c.Capabilities == nil {
		return fmt.Errorf("%w: capabilities cannot be nil", ErrInvalidConfig)
	}
	if err := c.Capabilities.Validate(); err != nil {
		return fmt.Errorf("%w: invalid capabilities: %v", ErrInvalidConfig, err)
	}
	if c.ExecutiveVersion == "" {
		return fmt.Errorf("%w: missing executive_version", ErrInvalidConfig)
	}
	return nil
}
