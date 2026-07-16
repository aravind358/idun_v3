package attention

import (
	"errors"

	"idun/core/logger"
)

// Configuration holds structural configuration for building an AttentionService.
type Configuration struct {
	Logger             logger.Writer
	PolicyProfile      *AttentionPolicyProfile
	Capabilities       *AttentionCapabilities
	ReplaySeed         int64
	WorkspacePublisher WorkspacePublisher
}

// Validate checks if Configuration holds valid, fingerprinted profile and capabilities.
func (c *Configuration) Validate() error {
	if c == nil {
		return errors.New("attention: Configuration cannot be nil")
	}
	if c.PolicyProfile == nil {
		return ErrNilProfile
	}
	if err := c.PolicyProfile.Validate(); err != nil {
		return err
	}
	if c.Capabilities == nil {
		return ErrNilCapabilities
	}
	if err := c.Capabilities.Validate(); err != nil {
		return err
	}
	return nil
}

// DefaultAttentionPolicyProfile returns the canonical frozen policy profile with SHA-256 fingerprint.
func DefaultAttentionPolicyProfile() *AttentionPolicyProfile {
	p := &AttentionPolicyProfile{
		PolicyVersion:         "2.0.0",
		SchemaVersion:         "2.0.0",
		Band0Threshold:        100, // By default Band 0 requires SafetyFlag or max score
		Band1Threshold:        85,  // RealTime threshold matching V1 heritage
		Band2Threshold:        50,  // Interactive threshold matching V1 heritage
		Band3Threshold:        20,  // Background threshold matching V1 heritage
		SwitchMargin:          5,   // 5 points hysteresis switch margin
		InterruptMargin:       10,  // 10 points break-in margin
		MaximumTrackedStimuli: 100,
	}
	fp, _ := ComputePolicyFingerprint(p)
	p.PolicyFingerprint = fp
	return p
}

// DefaultAttentionCapabilities returns the canonical deployment capabilities with SHA-256 fingerprint.
func DefaultAttentionCapabilities() *AttentionCapabilities {
	c := &AttentionCapabilities{
		SupportsInterruptions:        true,
		SupportsBackgroundAttention:  true,
		SupportsFocusSwitching:       true,
		SupportsMultimodalAttention:  true,
		SupportsDistributedAttention: false,
		SupportsFocusHistory:         true,
	}
	fp, _ := ComputeCapabilityFingerprint(c)
	c.CapabilityFingerprint = fp
	return c
}

// DefaultConfiguration returns a clean, validated Configuration with canonical defaults.
func DefaultConfiguration() *Configuration {
	return &Configuration{
		PolicyProfile: DefaultAttentionPolicyProfile(),
		Capabilities:  DefaultAttentionCapabilities(),
		ReplaySeed:    0,
	}
}
