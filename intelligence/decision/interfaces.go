package decision

import (
	"context"

	"idun/intelligence/executive"
)

// ============================================================================
// Core Decision Service Interface
// ============================================================================

// DecisionService defines the public contract for CognitiveAbility.Decision.
// It implements executive.DecisionAbility and provides explicit structured evaluation
// across Reflexive (<2ms) and Deliberative (50-500ms) execution surfaces.
type DecisionService interface {
	executive.DecisionAbility

	// Start boots the Decision service lifecycle and initializes Tier 1/2 engines.
	Start() error

	// Close gracefully shuts down the Decision service.
	Close() error

	// EvaluateReflexive performs fast-path linear utility commitment or emits an
	// EscalationRecommendation if uncertainty/margin thresholds are breached.
	EvaluateReflexive(ctx context.Context, cs CandidateSet) (*DecisionRecord, error)

	// EvaluateDeliberative performs rigorous Multi-Criteria Decision Analysis (MCDA)
	// or Pareto trade-off optimization across candidate set C.
	EvaluateDeliberative(ctx context.Context, cs CandidateSet) (*DecisionRecord, error)

	// GetEpisodeTrace retrieves the O(1) memory-bounded ReflexiveDecisionTrace for an episode.
	GetEpisodeTrace(episodeID string) (*ReflexiveDecisionTrace, bool)
}

// ============================================================================
// Internal Tier Interfaces (Orthogonal Evaluation Tiers)
// ============================================================================

// Tier1ConstitutionalGate defines the non-negotiable hard constitutional filter.
type Tier1ConstitutionalGate interface {
	// Filter evaluates CandidateSet against constitutional safety invariants,
	// returning surviving candidates and immediate rejection records.
	Filter(ctx context.Context, cs CandidateSet) ([]Candidate, []RejectedAlternative, error)
}

// Tier2ObjectiveScorer defines the objective scoring and trade-off engine.
type Tier2ObjectiveScorer interface {
	// ScoreReflexive computes linear dot-product utility scores U(c_i) = w^T * x_i.
	ScoreReflexive(candidates []Candidate, snapshot *DecisionStrategySnapshot) ([]CandidateScore, error)

	// ScoreDeliberative computes Multi-Criteria Decision Analysis (MCDA) trade-off matrix.
	ScoreDeliberative(ctx context.Context, candidates []Candidate, snapshot *DecisionStrategySnapshot) ([]CandidateScore, map[string]map[string]float64, error)
}

// CandidateScore represents the intermediate utility evaluation of a candidate.
type CandidateScore struct {
	CandidateID string
	Score       float64
	Confidence  float64
	Rationale   string
}

// StrategyProvider defines the contract for acquiring read-only strategy snapshots.
type StrategyProvider interface {
	// ActiveSnapshot returns the current immutable DecisionStrategySnapshot.
	ActiveSnapshot() (*DecisionStrategySnapshot, error)
}
