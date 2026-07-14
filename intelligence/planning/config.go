package planning

import (
	"errors"
	"fmt"
	"time"
)

// Config holds service-level configuration options for CognitiveAbility.Planning.
type Config struct {
	// DefaultProfile is the fallback PlanningPolicyProfile if no external snapshot is active.
	DefaultProfile *PlanningPolicyProfile `json:"default_profile"`

	// MaxTraceRetention limits the maximum number of PlanningTrace objects cached locally in ring buffers.
	MaxTraceRetention int `json:"max_trace_retention"`

	// EnableReflexiveCache controls whether Stage 1 exact/memoized cache lookups are active.
	EnableReflexiveCache bool `json:"enable_reflexive_cache"`

	// CacheTTL defines the time-to-live for reflexive cache entries.
	CacheTTL time.Duration `json:"cache_ttl"`

	// MaxConcurrentPlans limits the maximum number of simultaneous active planning workflows.
	MaxConcurrentPlans int `json:"max_concurrent_plans"`

	// Capabilities defines the engine-level capability boundaries for this deployment.
	Capabilities *PlanningCapabilities `json:"capabilities"`
}

// Validate verifies that service configuration parameters are within legal bounds.
func (c *Config) Validate() error {
	if c.MaxTraceRetention <= 0 {
		return errors.New("MaxTraceRetention must be positive")
	}
	if c.MaxConcurrentPlans <= 0 {
		return errors.New("MaxConcurrentPlans must be positive")
	}
	if c.CacheTTL <= 0 {
		return errors.New("CacheTTL must be positive")
	}
	if c.DefaultProfile == nil {
		return errors.New("DefaultProfile cannot be nil")
	}
	if err := c.DefaultProfile.Validate(); err != nil {
		return fmt.Errorf("DefaultProfile validation failed: %w", err)
	}
	if c.Capabilities != nil {
		if err := c.Capabilities.Validate(); err != nil {
			return fmt.Errorf("Capabilities validation failed: %w", err)
		}
	}
	return nil
}

// DefaultPlanningCapabilities initializes the standard production engine capabilities.
func DefaultPlanningCapabilities() *PlanningCapabilities {
	caps := &PlanningCapabilities{
		SupportsHTN:              true,
		SupportsGOAP:             true,
		SupportsTreeSearch:       true,
		SupportsParallelSearch:   true,
		MaxParallelWorkers:       16,
		MaxPlanningDepth:         128,
		MaxSupportedAlternatives: 32,
		SupportsConstraintSolve:  true,
		SupportsTemporalPlanning: true,
		SupportsContingencies:    true,
	}
	caps.CapabilityFingerprint = ComputeCapabilityFingerprint(caps)
	return caps
}

// DefaultPlanningPolicyProfile initializes the standard baseline immutable policy profile.
func DefaultPlanningPolicyProfile() *PlanningPolicyProfile {
	return &PlanningPolicyProfile{
		ProfileID:         "PROFILE_GENERAL_BASE",
		ProfileVersion:    "1.0",
		PolicyFingerprint: "E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855",
		PolicySource:      "CognitiveAbility.Learning.Bootstrap",
		PlanningDepthLimits: map[string]int{
			"REFLEXIVE": 10,
			"TACTICAL":  100,
			"STRATEGIC": 500,
		},
		SpecialistWeights: map[string]float64{
			"GoalDecomposition":   1.0,
			"TaskSequencing":      1.0,
			"DependencyAnalysis":  0.9,
			"ResourcePlanning":    0.8,
			"TimePlanning":        0.8,
			"RiskPlanning":        0.7,
			"ConstraintPlanning":  0.9,
			"ContingencyPlanning": 0.6,
			"AcquisitionPlanning": 0.7,
		},
		DomainWeights: map[string]float64{
			"General":      1.0,
			"Coding":       1.0,
			"Robotics":     1.0,
			"Conversation": 1.0,
			"Business":     1.0,
			"Research":     1.0,
			"PhysicalTask": 1.0,
		},
		EscalationThresholds: map[string]float64{
			"AmbiguityThreshold": 0.15,
			"ConfidenceFloor":    0.70,
		},
		SearchBudgets: map[string]int{
			"MaxNodesExpanded": 5000,
			"MaxNodesPruned":   10000,
		},
		MaxPlanningTime:     500 * time.Millisecond,
		MaxPlanningNodes:    5000,
		MaxAlternatives:     8,
		RiskPreferences: map[string]float64{
			"RiskTolerance": 0.20,
		},
		CalibrationWeight:   1.0,
		MaxBeamWidth:        3,
		MaxBranchDepth:      16,
		MaxInfoRequirements: 8,
		SearchStrategies: map[string]*PlanningSearchStrategy{
			"REFLEXIVE": {
				SearchID:               "SEARCH_REFLEXIVE_BASE",
				SearchVersion:          "1.0",
				SearchFingerprint:      "FP_REFLEXIVE",
				SearchType:             "TEMPLATE_MATCH",
				Description:            "Fast reflexive template lookup without deep tree search",
				MaxDepth:               2,
				MaxNodes:               10,
				BeamWidth:              1,
				AllowParallelExpansion: false,
				AllowBacktracking:      false,
				AllowReplanning:        false,
				ExpansionPolicy:        "DIRECT",
				PruningPolicy:          "NONE",
				MaxConcurrentWorkers:   1,
				SearchBudgetMs:         10,
			},
			"TACTICAL": {
				SearchID:               "SEARCH_TACTICAL_BASE",
				SearchVersion:          "1.0",
				SearchFingerprint:      "FP_TACTICAL",
				SearchType:             "BEAM_SEARCH",
				Description:            "Bounded beam search with HTN task decomposition",
				MaxDepth:               16,
				MaxNodes:               1000,
				BeamWidth:              3,
				AllowParallelExpansion: true,
				AllowBacktracking:      true,
				AllowReplanning:        true,
				ExpansionPolicy:        "BEST_FIRST",
				PruningPolicy:          "ALPHA_BETA",
				MaxConcurrentWorkers:   4,
				SearchBudgetMs:         200,
			},
			"STRATEGIC": {
				SearchID:               "SEARCH_STRATEGIC_BASE",
				SearchVersion:          "1.0",
				SearchFingerprint:      "FP_STRATEGIC",
				SearchType:             "MCTS",
				Description:            "Deep multi-alternative tree exploration with contingency planning",
				MaxDepth:               32,
				MaxNodes:               5000,
				BeamWidth:              5,
				AllowParallelExpansion: true,
				AllowBacktracking:      true,
				AllowReplanning:        true,
				ExpansionPolicy:        "UCT",
				PruningPolicy:          "PARETO_DOMINANCE",
				MaxConcurrentWorkers:   8,
				SearchBudgetMs:         1000,
			},
		},
		Capabilities: DefaultPlanningCapabilities(),
	}
}

// DefaultConfig initializes the default production configuration for PlanningService.
func DefaultConfig() *Config {
	return &Config{
		DefaultProfile:       DefaultPlanningPolicyProfile(),
		MaxTraceRetention:    1024,
		EnableReflexiveCache: true,
		CacheTTL:             30 * time.Minute,
		MaxConcurrentPlans:   64,
		Capabilities:         DefaultPlanningCapabilities(),
	}
}
