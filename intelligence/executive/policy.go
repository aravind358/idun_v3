// Package executive implements IDUN's Intelligence Pillar Executive Functions.
//
// Architecture Version: 2.0.0-FROZEN
//
// Phase 1 defines the immutable domain structures, validation firewalls, fluent builders,
// diagnostic traces, capabilities, and snapshot management that form the hardening
// foundation for Executive Version 2.0.
package executive

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"
)

const (
	// SchemaVersion is the frozen canonical schema version for Phase 1 Executive artifacts.
	SchemaVersion = "2.0.0-FROZEN"

	// SchemaVersion2_0_0 is the explicit alias for the frozen schema version.
	SchemaVersion2_0_0 = "2.0.0-FROZEN"

	// DefaultExecutiveVersion is the default binary build version for Executive.
	DefaultExecutiveVersion = "2.0.0"
)

var (
	// ErrInvalidPolicy is returned when an ExecutivePolicyProfile fails validation.
	ErrInvalidPolicy = errors.New("executive: invalid policy profile")

	// ErrInvalidCapabilities is returned when ExecutiveCapabilities fails validation.
	ErrInvalidCapabilities = errors.New("executive: invalid deployment capabilities")

	// ErrInvalidTrace is returned when ExecutiveTrace fails validation.
	ErrInvalidTrace = errors.New("executive: invalid coordination trace")

	// ErrInvalidResult is returned when ExecutiveResult fails validation.
	ErrInvalidResult = errors.New("executive: invalid coordination result")

	// ErrInvalidReplayMetadata is returned when ReplayMetadata fails validation.
	ErrInvalidReplayMetadata = errors.New("executive: invalid replay metadata")

	// ErrInvalidSummary is returned when ExecutiveCoordinationSummary fails validation.
	ErrInvalidSummary = errors.New("executive: invalid coordination summary")

	// ErrInvalidConfig is returned when Configuration fails validation.
	ErrInvalidConfig = errors.New("executive: invalid configuration")
)

// PriorityPolicies defines immutable prioritization and preemption rules.
type PriorityPolicies struct {
	MaxConcurrentPerBand  map[PriorityBand]int `json:"max_concurrent_per_band"`
	PreemptionAllowed     bool                 `json:"preemption_allowed"`
	EmergencyBand0Timeout time.Duration        `json:"emergency_band_0_timeout"`
}

// Validate checks numerical and structural soundness of PriorityPolicies.
func (p *PriorityPolicies) Validate() error {
	if p.EmergencyBand0Timeout < 0 {
		return fmt.Errorf("%w: emergency_band_0_timeout cannot be negative", ErrInvalidPolicy)
	}
	for band, limit := range p.MaxConcurrentPerBand {
		if band < PriorityBand0CriticalSafety || band > PriorityBand4Idle {
			return fmt.Errorf("%w: invalid priority band %d in max_concurrent_per_band", ErrInvalidPolicy, band)
		}
		if limit < 0 {
			return fmt.Errorf("%w: negative concurrency limit %d for band %d", ErrInvalidPolicy, limit, band)
		}
	}
	return nil
}

// BudgetPolicies defines computational budget limits and escalation multipliers.
type BudgetPolicies struct {
	MaxCycleBudgetUnits      int     `json:"max_cycle_budget_units"`
	MaxFuelPerWorkflow       int     `json:"max_fuel_per_workflow"`
	MaxReflectionDepth       int     `json:"max_reflection_depth"`
	EscalationUnitMultiplier float64 `json:"escalation_unit_multiplier"`
}

// Validate checks bounds on BudgetPolicies.
func (b *BudgetPolicies) Validate() error {
	if b.MaxCycleBudgetUnits <= 0 {
		return fmt.Errorf("%w: max_cycle_budget_units must be > 0", ErrInvalidPolicy)
	}
	if b.MaxFuelPerWorkflow <= 0 {
		return fmt.Errorf("%w: max_fuel_per_workflow must be > 0", ErrInvalidPolicy)
	}
	if b.MaxReflectionDepth < 0 {
		return fmt.Errorf("%w: max_reflection_depth cannot be negative", ErrInvalidPolicy)
	}
	if math.IsNaN(b.EscalationUnitMultiplier) || math.IsInf(b.EscalationUnitMultiplier, 0) || b.EscalationUnitMultiplier < 1.0 {
		return fmt.Errorf("%w: escalation_unit_multiplier must be >= 1.0", ErrInvalidPolicy)
	}
	return nil
}

// RetryPolicies governs backoff and retry budgets across cognitive abilities.
type RetryPolicies struct {
	MaxRetries        int           `json:"max_retries"`
	InitialBackoff    time.Duration `json:"initial_backoff"`
	MaxBackoff        time.Duration `json:"max_backoff"`
	BackoffMultiplier float64       `json:"backoff_multiplier"`
}

// Validate checks bounds on RetryPolicies.
func (r *RetryPolicies) Validate() error {
	if r.MaxRetries < 0 {
		return fmt.Errorf("%w: max_retries cannot be negative", ErrInvalidPolicy)
	}
	if r.InitialBackoff < 0 || r.MaxBackoff < 0 {
		return fmt.Errorf("%w: backoff durations cannot be negative", ErrInvalidPolicy)
	}
	if r.MaxBackoff < r.InitialBackoff {
		return fmt.Errorf("%w: max_backoff cannot be less than initial_backoff", ErrInvalidPolicy)
	}
	if math.IsNaN(r.BackoffMultiplier) || math.IsInf(r.BackoffMultiplier, 0) || r.BackoffMultiplier < 1.0 {
		return fmt.Errorf("%w: backoff_multiplier must be >= 1.0", ErrInvalidPolicy)
	}
	return nil
}

// CancellationPolicies governs cooperative cancellation timeouts and propagation.
type CancellationPolicies struct {
	CooperativeTimeout  time.Duration `json:"cooperative_timeout"`
	ForceKillOnTimeout  bool          `json:"force_kill_on_timeout"`
	PropagateToChildren bool          `json:"propagate_to_children"`
}

// Validate checks bounds on CancellationPolicies.
func (c *CancellationPolicies) Validate() error {
	if c.CooperativeTimeout < 0 {
		return fmt.Errorf("%w: cooperative_timeout cannot be negative", ErrInvalidPolicy)
	}
	return nil
}

// EscalationPolicies governs when workflows require tier upgrades or explicit user confirmation.
type EscalationPolicies struct {
	MaxEscalationsPerWorkflow   int        `json:"max_escalations_per_workflow"`
	RequireUserConfirmationTier BudgetTier `json:"require_user_confirmation_tier"`
	AmbiguityThreshold          float64    `json:"ambiguity_threshold"`
}

// Validate checks bounds on EscalationPolicies.
func (e *EscalationPolicies) Validate() error {
	if e.MaxEscalationsPerWorkflow < 0 {
		return fmt.Errorf("%w: max_escalations_per_workflow cannot be negative", ErrInvalidPolicy)
	}
	if math.IsNaN(e.AmbiguityThreshold) || math.IsInf(e.AmbiguityThreshold, 0) || e.AmbiguityThreshold < 0.0 || e.AmbiguityThreshold > 1.0 {
		return fmt.Errorf("%w: ambiguity_threshold out of bounds [0.0, 1.0]", ErrInvalidPolicy)
	}
	return nil
}

// WorkspacePolicies governs Global Workspace admission thresholds and impasse rules.
type WorkspacePolicies struct {
	AdmissionThreshold     float64       `json:"admission_threshold"`
	MaxPendingBidsPerTopic int           `json:"max_pending_bids_per_topic"`
	ImpasseUrgency         int           `json:"impasse_urgency"`
	ImpasseTimeout         time.Duration `json:"impasse_timeout"`
}

// Validate checks bounds on WorkspacePolicies.
func (w *WorkspacePolicies) Validate() error {
	if math.IsNaN(w.AdmissionThreshold) || math.IsInf(w.AdmissionThreshold, 0) || w.AdmissionThreshold < 0.0 || w.AdmissionThreshold > 1.0 {
		return fmt.Errorf("%w: admission_threshold out of bounds [0.0, 1.0]", ErrInvalidPolicy)
	}
	if w.MaxPendingBidsPerTopic < 1 {
		return fmt.Errorf("%w: max_pending_bids_per_topic must be >= 1", ErrInvalidPolicy)
	}
	if w.ImpasseUrgency < 0 || w.ImpasseUrgency > 100 {
		return fmt.Errorf("%w: impasse_urgency out of bounds [0, 100]", ErrInvalidPolicy)
	}
	if w.ImpasseTimeout < 0 {
		return fmt.Errorf("%w: impasse_timeout cannot be negative", ErrInvalidPolicy)
	}
	return nil
}

// HomeostasisPolicies governs periodic consolidation intervals and idle detection.
type HomeostasisPolicies struct {
	IdleThreshold         time.Duration `json:"idle_threshold"`
	ConsolidationInterval time.Duration `json:"consolidation_interval"`
	MaxPeriodicFuel       int           `json:"max_periodic_fuel"`
}

// Validate checks bounds on HomeostasisPolicies.
func (h *HomeostasisPolicies) Validate() error {
	if h.IdleThreshold < 0 || h.ConsolidationInterval < 0 {
		return fmt.Errorf("%w: homeostasis durations cannot be negative", ErrInvalidPolicy)
	}
	if h.MaxPeriodicFuel < 0 {
		return fmt.Errorf("%w: max_periodic_fuel cannot be negative", ErrInvalidPolicy)
	}
	return nil
}

// ExecutivePolicyProfile defines the immutable, read-only policy snapshot that dictates
// all coordination rules, thresholds, priority limits, and retry governance for Executive.
// Executive strictly consumes this profile; only Learning may publish future versions.
type ExecutivePolicyProfile struct {
	ProfileID            string               `json:"profile_id"`
	ProfileVersion       string               `json:"profile_version"`
	SchemaVersion        string               `json:"schema_version"`
	PolicyFingerprint    string               `json:"policy_fingerprint"`
	PolicySource         string               `json:"policy_source"`
	PriorityPolicies     PriorityPolicies     `json:"priority_policies"`
	BudgetPolicies       BudgetPolicies       `json:"budget_policies"`
	RetryPolicies        RetryPolicies        `json:"retry_policies"`
	CancellationPolicies CancellationPolicies `json:"cancellation_policies"`
	EscalationPolicies   EscalationPolicies   `json:"escalation_policies"`
	WorkspacePolicies    WorkspacePolicies    `json:"workspace_policies"`
	HomeostasisPolicies  HomeostasisPolicies  `json:"homeostasis_policies"`
}

// Validate implements the Validation Firewall for ExecutivePolicyProfile.
func (p *ExecutivePolicyProfile) Validate() error {
	if p == nil {
		return fmt.Errorf("%w: profile is nil", ErrInvalidPolicy)
	}
	if p.ProfileID == "" || p.ProfileVersion == "" || p.PolicyFingerprint == "" || p.PolicySource == "" {
		return fmt.Errorf("%w: missing required profile metadata fields", ErrInvalidPolicy)
	}
	if p.SchemaVersion != SchemaVersion {
		return fmt.Errorf("%w: schema_version %q != required %q", ErrInvalidPolicy, p.SchemaVersion, SchemaVersion)
	}
	if err := p.PriorityPolicies.Validate(); err != nil {
		return err
	}
	if err := p.BudgetPolicies.Validate(); err != nil {
		return err
	}
	if err := p.RetryPolicies.Validate(); err != nil {
		return err
	}
	if err := p.CancellationPolicies.Validate(); err != nil {
		return err
	}
	if err := p.EscalationPolicies.Validate(); err != nil {
		return err
	}
	if err := p.WorkspacePolicies.Validate(); err != nil {
		return err
	}
	if err := p.HomeostasisPolicies.Validate(); err != nil {
		return err
	}
	// Verify exact fingerprint match
	computed := p.GeneratePolicyFingerprint()
	if computed != p.PolicyFingerprint {
		return fmt.Errorf("%w: policy_fingerprint mismatch (got %s, expected %s)", ErrInvalidPolicy, p.PolicyFingerprint, computed)
	}
	return nil
}

// GeneratePolicyFingerprint calculates a deterministic SHA-256 hash across all policy fields
// excluding PolicyFingerprint itself.
func (p *ExecutivePolicyProfile) GeneratePolicyFingerprint() string {
	if p == nil {
		return ""
	}
	h := sha256.New()
	fmt.Fprintf(h, "ProfileID:%s|Version:%s|Schema:%s|Source:%s|", p.ProfileID, p.ProfileVersion, p.SchemaVersion, p.PolicySource)

	// Priority
	var bands []int
	for b := range p.PriorityPolicies.MaxConcurrentPerBand {
		bands = append(bands, int(b))
	}
	sort.Ints(bands)
	for _, b := range bands {
		fmt.Fprintf(h, "Band%d:%d|", b, p.PriorityPolicies.MaxConcurrentPerBand[PriorityBand(b)])
	}
	fmt.Fprintf(h, "Preempt:%t|EmgTimeout:%d|", p.PriorityPolicies.PreemptionAllowed, p.PriorityPolicies.EmergencyBand0Timeout)

	// Budget
	fmt.Fprintf(h, "MaxCycle:%d|MaxFuel:%d|MaxRefl:%d|EscMult:%.6f|",
		p.BudgetPolicies.MaxCycleBudgetUnits, p.BudgetPolicies.MaxFuelPerWorkflow,
		p.BudgetPolicies.MaxReflectionDepth, p.BudgetPolicies.EscalationUnitMultiplier)

	// Retry
	fmt.Fprintf(h, "MaxRetries:%d|InitBO:%d|MaxBO:%d|BOMult:%.6f|",
		p.RetryPolicies.MaxRetries, p.RetryPolicies.InitialBackoff,
		p.RetryPolicies.MaxBackoff, p.RetryPolicies.BackoffMultiplier)

	// Cancellation
	fmt.Fprintf(h, "CoopTO:%d|ForceKill:%t|Propagate:%t|",
		p.CancellationPolicies.CooperativeTimeout, p.CancellationPolicies.ForceKillOnTimeout, p.CancellationPolicies.PropagateToChildren)

	// Escalation
	fmt.Fprintf(h, "MaxEsc:%d|ReqConfTier:%d|AmbThresh:%.6f|",
		p.EscalationPolicies.MaxEscalationsPerWorkflow, p.EscalationPolicies.RequireUserConfirmationTier, p.EscalationPolicies.AmbiguityThreshold)

	// Workspace
	fmt.Fprintf(h, "AdmThresh:%.6f|MaxPending:%d|ImpUrgency:%d|ImpTO:%d|",
		p.WorkspacePolicies.AdmissionThreshold, p.WorkspacePolicies.MaxPendingBidsPerTopic,
		p.WorkspacePolicies.ImpasseUrgency, p.WorkspacePolicies.ImpasseTimeout)

	// Homeostasis
	fmt.Fprintf(h, "IdleThresh:%d|ConsInt:%d|MaxPerFuel:%d|",
		p.HomeostasisPolicies.IdleThreshold, p.HomeostasisPolicies.ConsolidationInterval, p.HomeostasisPolicies.MaxPeriodicFuel)

	return hex.EncodeToString(h.Sum(nil))
}
