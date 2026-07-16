package executive

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// ExecutiveCapabilities describes the deployment engine bounds and architectural limits
// supported by this specific runtime deployment of Executive Functions.
// Executive never modifies these deployment capabilities.
type ExecutiveCapabilities struct {
	SupportsInterrupts            bool   `json:"supports_interrupts"`
	SupportsPauseResume           bool   `json:"supports_pause_resume"`
	SupportsScheduling            bool   `json:"supports_scheduling"`
	SupportsRecovery              bool   `json:"supports_recovery"`
	SupportsCheckpointing         bool   `json:"supports_checkpointing"`
	SupportsBackgroundTasks       bool   `json:"supports_background_tasks"`
	SupportsWorkspaceArbitration  bool   `json:"supports_workspace_arbitration"`
	SupportsCalibration           bool   `json:"supports_calibration"`
	SupportsConstitution          bool   `json:"supports_constitution"`
	MaxConcurrentEpisodes         int    `json:"max_concurrent_episodes"`
	MaxRetryBudget                int    `json:"max_retry_budget"`
	CapabilityFingerprint         string `json:"capability_fingerprint"`
}

// Validate implements the Validation Firewall for ExecutiveCapabilities.
func (c *ExecutiveCapabilities) Validate() error {
	if c == nil {
		return fmt.Errorf("%w: capabilities object is nil", ErrInvalidCapabilities)
	}
	if c.MaxConcurrentEpisodes < 1 {
		return fmt.Errorf("%w: max_concurrent_episodes must be >= 1", ErrInvalidCapabilities)
	}
	if c.MaxRetryBudget < 0 {
		return fmt.Errorf("%w: max_retry_budget cannot be negative", ErrInvalidCapabilities)
	}
	if c.CapabilityFingerprint == "" {
		return fmt.Errorf("%w: missing capability_fingerprint", ErrInvalidCapabilities)
	}
	computed := c.GenerateCapabilityFingerprint()
	if computed != c.CapabilityFingerprint {
		return fmt.Errorf("%w: capability_fingerprint mismatch (got %s, expected %s)", ErrInvalidCapabilities, c.CapabilityFingerprint, computed)
	}
	return nil
}

// GenerateCapabilityFingerprint computes a deterministic SHA-256 hash across all capability flags and limits.
func (c *ExecutiveCapabilities) GenerateCapabilityFingerprint() string {
	if c == nil {
		return ""
	}
	h := sha256.New()
	fmt.Fprintf(h, "Int:%t|Pause:%t|Sched:%t|Rec:%t|Chk:%t|Bg:%t|Wksp:%t|Calib:%t|Const:%t|MaxConc:%d|MaxRetry:%d|",
		c.SupportsInterrupts,
		c.SupportsPauseResume,
		c.SupportsScheduling,
		c.SupportsRecovery,
		c.SupportsCheckpointing,
		c.SupportsBackgroundTasks,
		c.SupportsWorkspaceArbitration,
		c.SupportsCalibration,
		c.SupportsConstitution,
		c.MaxConcurrentEpisodes,
		c.MaxRetryBudget,
	)
	return hex.EncodeToString(h.Sum(nil))
}
