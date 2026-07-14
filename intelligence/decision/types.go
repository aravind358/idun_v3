// Package decision implements IDUN's Intelligence Pillar Decision Subsystem (Commitment Under Uncertainty).
//
// Architecture Version: 2.0.0-FROZEN
// Classification: Core Cognitive Ability Specification
//
// Every cognitive architecture reaches branching points where multiple courses of action
// are simultaneously valid. Decision is the sole cognitive ability authorized to perform
// Commitment Under Uncertainty—collapsing a space of live candidate alternatives into a
// single committed outcome or an explicit epistemic non-choice / escalation recommendation.
package decision

import (
	"errors"
	"fmt"
	"time"
)

// ============================================================================
// Core Errors
// ============================================================================

var (
	// ErrEmptyCandidateSet indicates that evaluation was requested with zero candidates.
	ErrEmptyCandidateSet = errors.New("decision: candidate set cannot be empty")

	// ErrCandidateSetOverflow indicates that the candidate set exceeds the maximum bound of 16.
	ErrCandidateSetOverflow = errors.New("decision: candidate set exceeds maximum capacity (16)")

	// ErrInvalidStrategySnapshot indicates that an invalid or nil strategy snapshot was provided.
	ErrInvalidStrategySnapshot = errors.New("decision: invalid strategy snapshot")

	// ErrInvalidDecisionRecord indicates that an invalid or nil decision record was provided.
	ErrInvalidDecisionRecord = errors.New("decision: invalid decision record")
	// ErrInvalidSchemaVersion indicates that a DecisionRecord lacks a valid schema version.
	ErrInvalidSchemaVersion = errors.New("decision: invalid schema version")

	// ErrInvalidConfidence indicates that a confidence value falls outside [0.0, 1.0].
	ErrInvalidConfidence = errors.New("decision: confidence must be between 0.0 and 1.0")
)

// ============================================================================
// Outcome & Deliberation Depth Enumerations
// ============================================================================

// OutcomeType defines the standardized result classification of a Decision evaluation.
type OutcomeType string

const (
	// OutcomeCommit indicates selection of exactly one candidate outcome c*.
	OutcomeCommit OutcomeType = "COMMIT"

	// OutcomeDefer indicates intentional postponement of decision commitment.
	OutcomeDefer OutcomeType = "DEFER"

	// OutcomeAbstain indicates principled refusal to select any candidate from the current set.
	OutcomeAbstain OutcomeType = "ABSTAIN"

	// OutcomeRequestCandidates indicates that additional candidate options are required from Reasoning.
	OutcomeRequestCandidates OutcomeType = "REQUEST_MORE_CANDIDATES"

	// OutcomeRequestAdditionalInfo indicates that critical missing attributes must be resolved.
	OutcomeRequestAdditionalInfo OutcomeType = "REQUEST_ADDITIONAL_INFO"

	// OutcomeEscalateToDeliberative indicates a recommendation to escalate from Reflexive to Deliberative depth.
	OutcomeEscalateToDeliberative OutcomeType = "ESCALATE_TO_DELIBERATIVE"
)

// DeliberationDepth defines the computational and evaluation depth of the decision pipeline.
type DeliberationDepth string

const (
	// DepthReflexive represents the fast-path micro-decision execution surface (< 2ms budget).
	DepthReflexive DeliberationDepth = "REFLEXIVE_MICRO"

	// DepthDeliberative represents the multi-criteria macro-decision execution surface (50-500ms budget).
	DepthDeliberative DeliberationDepth = "DELIBERATIVE_MACRO"
)

// ============================================================================
// Input Domain Types: Candidates & CandidateSet
// ============================================================================

// Candidate represents an individual option or alternative c_i in candidate set C.
type Candidate struct {
	ID             string             `json:"id"`
	Description    string             `json:"description"`
	SourceAbility  string             `json:"source_ability"` // e.g. "Reasoning", "Planning"
	Attributes     map[string]float64 `json:"attributes"`     // Feature vector x_i for scoring
	Metadata       map[string]string  `json:"metadata,omitempty"`
	FlaggedRisks   []string           `json:"flagged_risks,omitempty"`
	EstimatedCost  float64            `json:"estimated_cost"`
	EstimatedBenefit float64          `json:"estimated_benefit"`
}

// CandidateSet represents a bounded collection of alternatives C (1 <= |C| <= 16).
type CandidateSet struct {
	EpisodeID  string      `json:"episode_id"`
	Candidates []Candidate `json:"candidates"`
}

// Validate checks that CandidateSet conforms to architectural bounds (1 <= len <= 16).
func (cs CandidateSet) Validate() error {
	if len(cs.Candidates) == 0 {
		return ErrEmptyCandidateSet
	}
	if len(cs.Candidates) > 16 {
		return fmt.Errorf("%w: got %d candidates", ErrCandidateSetOverflow, len(cs.Candidates))
	}
	return nil
}

// ============================================================================
// Public Output Contract: DecisionRecord & Components
// ============================================================================

// InformationGap documents missing candidate attributes or epistemic blind spots.
type InformationGap struct {
	CandidateID      string `json:"candidate_id"`
	MissingAttribute string `json:"missing_attribute"`
	Reason           string `json:"reason"`
	ImpactOnChoice   string `json:"impact_on_choice"`
	TargetProvider   string `json:"target_provider"` // "UNDERSTANDING", "MEMORY", "HOST_INPUT"
}

// RejectedAlternative captures candidates eliminated during Tier 1 or Tier 2 evaluation.
type RejectedAlternative struct {
	CandidateID    string  `json:"candidate_id"`
	RejectionStage string  `json:"rejection_stage"` // "TIER_1_CONSTITUTION", "TIER_2_SCORING"
	PrimaryReason  string  `json:"primary_reason"`
	ScoreDelta     float64 `json:"score_delta"`     // Normalized distance from winning choice
}

// EscalationRecommendation details the dimensions triggering a recommendation to escalate.
type EscalationRecommendation struct {
	TriggeredDimensions []string `json:"triggered_dimensions"` // e.g. "CONFIDENCE_DROP", "AMBIGUITY_MARGIN", "TAIL_RISK"
	ConfidenceDelta     float64  `json:"confidence_delta"`
	UtilityScoreMargin  float64  `json:"utility_score_margin"`
	Reason              string   `json:"reason"`
}

// DecisionRecord is the standardized public output contract produced by every Decision evaluation.
// Per Phase 5 production invariants, a DecisionRecord becomes permanently immutable upon publication.
type DecisionRecord struct {
	DecisionID               string                        `json:"decision_id"`
	EpisodeID                string                        `json:"episode_id"`
	SchemaVersion            string                        `json:"schema_version"`
	Timestamp                time.Time                     `json:"timestamp"`
	StrategyVersion          string                        `json:"strategy_version"`
	DeliberationDepth        DeliberationDepth             `json:"deliberation_depth"`
	SelectedOutcome          OutcomeType                   `json:"selected_outcome"`
	SelectedCandidateID      string                        `json:"selected_candidate_id,omitempty"`
	Confidence               float64                       `json:"confidence"`
	Rationale                string                        `json:"rationale"`
	ReplaySeed               uint64                        `json:"replay_seed,omitempty"`
	RejectedCandidates       []RejectedAlternative         `json:"rejected_candidates"`
	ConstraintsApplied       []string                      `json:"constraints_applied"`
	InformationGaps          []InformationGap              `json:"information_gaps,omitempty"`
	FlaggedAssumptions       []string                      `json:"flagged_assumptions"`
	EscalationRecommendation *EscalationRecommendation     `json:"escalation_recommendation,omitempty"`
	TradeoffMatrix           map[string]map[string]float64 `json:"tradeoff_matrix,omitempty"`
}

// Validate executes the Validation Firewall, ensuring structural integrity before publication.
func (r *DecisionRecord) Validate() error {
	if r == nil || r.DecisionID == "" || r.EpisodeID == "" {
		return ErrInvalidDecisionRecord
	}
	if r.SchemaVersion == "" {
		return ErrInvalidSchemaVersion
	}
	if r.Confidence < 0.0 || r.Confidence > 1.0 {
		return ErrInvalidConfidence
	}
	if r.DeliberationDepth != DepthReflexive && r.DeliberationDepth != DepthDeliberative {
		return fmt.Errorf("%w: invalid deliberation depth '%s'", ErrInvalidDecisionRecord, r.DeliberationDepth)
	}
	switch r.SelectedOutcome {
	case OutcomeCommit:
		if r.SelectedCandidateID == "" {
			return fmt.Errorf("%w: COMMIT outcome requires selected candidate ID", ErrInvalidDecisionRecord)
		}
	case OutcomeDefer, OutcomeAbstain, OutcomeRequestCandidates, OutcomeRequestAdditionalInfo, OutcomeEscalateToDeliberative:
		// Valid non-commitment outcomes
	default:
		return fmt.Errorf("%w: unknown outcome type '%s'", ErrInvalidDecisionRecord, r.SelectedOutcome)
	}
	return nil
}

// ============================================================================
// Strategy Snapshot & Decision Policy Profiles
// ============================================================================

// DecisionPolicyProfile encapsulates versioned, cohesive decision parameters
// for a specific operating mode or risk regime (e.g., "BALANCED", "CONSERVATIVE_SAFETY").
type DecisionPolicyProfile struct {
	ProfileID                 string             `json:"profile_id"`
	PolicyVersion             string             `json:"policy_version"`
	PolicySource              string             `json:"policy_source"`
	PolicyFingerprint         string             `json:"policy_fingerprint"`
	Description               string             `json:"description"`
	FeatureWeights            map[string]float64 `json:"feature_weights"`
	RiskTolerance             float64            `json:"risk_tolerance"`              // 0.0 (zero tolerance) to 1.0 (high tolerance)
	EscalationConfidenceFloor float64            `json:"escalation_confidence_floor"` // e.g. 0.60
	EscalationAmbiguityMargin float64            `json:"escalation_ambiguity_margin"` // e.g. 0.05
	ObjectivePriorities       []string           `json:"objective_priorities"`        // e.g. ["safety", "utility", "reliability"]
	MaxReflexiveLatencyUs     uint32             `json:"max_reflexive_latency_us"`    // e.g. 2000
}

// DecisionStrategySnapshot represents an immutable read-only snapshot of scoring weights,
// calibration thresholds, and active decision policy profile published by Learning/Executive.
type DecisionStrategySnapshot struct {
	StrategyVersion           string                `json:"strategy_version"`
	ActiveProfileID           string                `json:"active_profile_id"`
	ActiveProfile             DecisionPolicyProfile `json:"active_profile"`

	// Direct convenience accessors mirroring ActiveProfile parameters
	FeatureWeights            map[string]float64    `json:"feature_weights"`
	EscalationConfidenceFloor float64               `json:"escalation_confidence_floor"`
	EscalationAmbiguityMargin float64               `json:"escalation_ambiguity_margin"`
	MaxReflexiveLatencyUs     uint32                `json:"max_reflexive_latency_us"`
}

// Validate verifies that the snapshot contains required strategy fields.
func (s *DecisionStrategySnapshot) Validate() error {
	if s == nil || s.StrategyVersion == "" {
		return ErrInvalidStrategySnapshot
	}
	return nil
}
