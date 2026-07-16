package attention

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// AttentionPolicyProfile configures immutable, fingerprinted salience thresholds and hysteresis behavior.
// This artifact is owned by Executive Functions and consumed read-only by Attention.
type AttentionPolicyProfile struct {
	PolicyVersion         string `json:"policy_version"`
	SchemaVersion         string `json:"schema_version"`
	Band0Threshold        int    `json:"band0_threshold"`
	Band1Threshold        int    `json:"band1_threshold"`
	Band2Threshold        int    `json:"band2_threshold"`
	Band3Threshold        int    `json:"band3_threshold"`
	SwitchMargin          int    `json:"switch_margin"`
	InterruptMargin       int    `json:"interrupt_margin"`
	MaximumTrackedStimuli int    `json:"maximum_tracked_stimuli"`
	PolicyFingerprint     string `json:"policy_fingerprint"`
}

// Validate verifies that the AttentionPolicyProfile satisfies all structural invariants and thresholds.
func (p *AttentionPolicyProfile) Validate() error {
	if p == nil {
		return ErrNilProfile
	}
	if p.PolicyVersion == "" {
		return ErrMissingPolicyVersion
	}
	if p.SchemaVersion == "" {
		return ErrMissingSchemaVersion
	}
	if p.PolicyFingerprint == "" {
		return ErrMissingFingerprint
	}
	// Check thresholds fall within [0..100] and are logically ordered
	if p.Band0Threshold < 0 || p.Band0Threshold > 100 ||
		p.Band1Threshold < 0 || p.Band1Threshold > 100 ||
		p.Band2Threshold < 0 || p.Band2Threshold > 100 ||
		p.Band3Threshold < 0 || p.Band3Threshold > 100 {
		return ErrInvalidThreshold
	}
	if p.Band1Threshold < p.Band2Threshold || p.Band2Threshold < p.Band3Threshold {
		return fmt.Errorf("attention: thresholds must be monotonic decreasing (Band1 >= Band2 >= Band3)")
	}
	if p.SwitchMargin < 0 || p.InterruptMargin < 0 {
		return ErrInvalidMargin
	}
	if p.MaximumTrackedStimuli < 0 {
		return ErrNegativeTrackedLimit
	}
	return nil
}

// ComputePolicyFingerprint computes the deterministic SHA-256 digest over structural policy settings.
func ComputePolicyFingerprint(p *AttentionPolicyProfile) (string, error) {
	if p == nil {
		return "", ErrNilProfile
	}
	raw := fmt.Sprintf("policy:%s|schema:%s|b0:%d|b1:%d|b2:%d|b3:%d|sm:%d|im:%d|max:%d",
		p.PolicyVersion,
		p.SchemaVersion,
		p.Band0Threshold,
		p.Band1Threshold,
		p.Band2Threshold,
		p.Band3Threshold,
		p.SwitchMargin,
		p.InterruptMargin,
		p.MaximumTrackedStimuli,
	)
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:]), nil
}

// ComputeCapabilityFingerprint computes the deterministic SHA-256 digest over capabilities.
func ComputeCapabilityFingerprint(c *AttentionCapabilities) (string, error) {
	if c == nil {
		return "", ErrNilCapabilities
	}
	raw := fmt.Sprintf("caps|int:%t|bg:%t|switch:%t|multi:%t|dist:%t|hist:%t",
		c.SupportsInterruptions,
		c.SupportsBackgroundAttention,
		c.SupportsFocusSwitching,
		c.SupportsMultimodalAttention,
		c.SupportsDistributedAttention,
		c.SupportsFocusHistory,
	)
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:]), nil
}
