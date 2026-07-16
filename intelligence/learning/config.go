package learning

import (
	"fmt"
	"time"
)

// Config holds operational parameters and immutable baseline profiles for the Learning Service.
type Config struct {
	ServiceVersion        string                 `json:"service_version"`
	PolicyProfile         *LearningPolicyProfile `json:"policy_profile"`
	Capabilities          *LearningCapabilities  `json:"capabilities"`
	AggregationPolicy     *AggregationPolicy     `json:"aggregation_policy,omitempty"`
	PublishToWorkspace    bool                   `json:"publish_to_workspace"`
	ValidateFirewall      bool                   `json:"validate_firewall"`
	MaxWorkers            int                    `json:"max_workers"`
	DefaultCooldown       time.Duration          `json:"default_cooldown"`
	MinSampleFloorDefault int                    `json:"min_sample_floor_default"`
}

// DefaultLearningPolicyProfile constructs a safe, conservative default profile owned by Executive.
func DefaultLearningPolicyProfile() *LearningPolicyProfile {
	return &LearningPolicyProfile{
		ProfileID:         "prof-learning-default-v2.0.0",
		PolicyVersion:     "2.0.0",
		PolicyFingerprint: "fp-policy-default-sha256",
		Author:            "Executive",
		LearningRate:      0.01,
		MinimumSampleSize: 100,
		CooldownPeriod:    24 * time.Hour,
		ValidationThresholds: map[string]float64{
			"min_confidence":        0.85,
			"max_regression_sigma":  0.05,
			"min_contribution":      0.10,
		},
		DriftDetectionThresholds: map[string]float64{
			"distribution_shift_p_val": 0.01,
			"latency_regression_ms":    15.0,
		},
		DomainWeights: map[string]float64{
			"idun.reasoning.strategy.v1":  1.0,
			"idun.planning.htn.v1":        1.0,
			"idun.decision.weights.v1":    1.0,
		},
		LearnerWeights: map[string]float64{
			"statistical-pattern-learner": 1.0,
			"rule-consolidation-learner":  1.0,
		},
		ExperimentLimits: map[string]int{
			"max_concurrent_shadow": 3,
			"max_concurrent_canary": 1,
		},
	}
}

// DefaultLearningCapabilities constructs the baseline structural capabilities advertised by this deployment.
func DefaultLearningCapabilities() *LearningCapabilities {
	return &LearningCapabilities{
		SupportsCalibrationLearning:   true,
		SupportsPreferenceLearning:    true,
		SupportsStatisticalLearning:   true,
		SupportsReinforcementLearning: false, // Disabled by default until Phase 3
		SupportsOnlineLearning:        false, // Strictly offline by default
		SupportsOfflineLearning:       true,
		SupportsExperimentation:       true,
		SupportsRollback:              true,
		SupportsShadowDeployment:      true,
		CapabilityFingerprint:         "fp-cap-default-sha256",
	}
}

// DefaultAggregationPolicy returns the standard Executive-governed aggregation policy.
func DefaultAggregationPolicy() *AggregationPolicy {
	return &AggregationPolicy{
		PolicyID:           "agg-policy-default-v2.0.0",
		PolicyVersion:      SchemaVersion,
		PolicyFingerprint:  "sha256-agg-policy-default-v2.0.0",
		Strategy:           "COGNITIVE_PERFORMANCE_DEFAULT",
		WindowDuration:     24 * time.Hour,
		SamplingMethod:     SamplingMethodDeterministic,
		DomainPriorities:   map[string]float64{"idun.reasoning.strategy.v1": 1.0, "idun.planning.policy.v1": 1.0, "idun.decision.weights.v1": 1.0},
		MaximumArtifacts:   500,
		MaximumMemoryBytes: MaxPayloadBytes,
		OrderingStrategy:   OrderingStrategyChronologicalAsc,
		MergePolicy:        MergePolicyDeduplicateHash,
	}
}

// DefaultConfig returns the baseline configuration for the Learning Service.
func DefaultConfig() *Config {
	return &Config{
		ServiceVersion:        SchemaVersion,
		PolicyProfile:         DefaultLearningPolicyProfile(),
		Capabilities:          DefaultLearningCapabilities(),
		AggregationPolicy:     DefaultAggregationPolicy(),
		PublishToWorkspace:    true,
		ValidateFirewall:      true,
		MaxWorkers:            4,
		DefaultCooldown:       24 * time.Hour,
		MinSampleFloorDefault: 100,
	}
}

// Validate verifies the configuration integrity and underlying profiles.
func (c *Config) Validate() error {
	if c.ServiceVersion == "" {
		return fmt.Errorf("%w: missing service_version", ErrMissingID)
	}
	if c.ServiceVersion != SchemaVersion {
		return fmt.Errorf("%w: service version %q != %q", ErrInvalidSchemaVersion, c.ServiceVersion, SchemaVersion)
	}
	if c.MaxWorkers < 1 {
		return fmt.Errorf("%w: max_workers must be >= 1", ErrValidationFailed)
	}
	if c.MinSampleFloorDefault < 1 {
		return fmt.Errorf("%w: min_sample_floor_default must be >= 1", ErrValidationFailed)
	}
	if c.PolicyProfile == nil || c.Capabilities == nil {
		return fmt.Errorf("%w: policy_profile and capabilities must not be nil", ErrValidationFailed)
	}
	if err := c.PolicyProfile.Validate(); err != nil {
		return fmt.Errorf("config policy validation failed: %w", err)
	}
	if err := c.Capabilities.Validate(); err != nil {
		return fmt.Errorf("config capabilities validation failed: %w", err)
	}
	if c.AggregationPolicy != nil {
		if err := c.AggregationPolicy.Validate(); err != nil {
			return fmt.Errorf("config aggregation policy validation failed: %w", err)
		}
	}
	return nil
}
