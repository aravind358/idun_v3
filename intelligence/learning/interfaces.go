package learning

import (
	"context"

	"idun/intelligence/executive"
)

// HealthAction defines the crisp operational imperative recommended by Governance.
type HealthAction string

const (
	ActionPauseRollout     HealthAction = "RECOMMEND_PAUSE_ROLLOUT"
	ActionResumeRollout    HealthAction = "RECOMMEND_RESUME_ROLLOUT"
	ActionTriggerRollback  HealthAction = "RECOMMEND_TRIGGER_ROLLBACK"
	ActionDisableLearner   HealthAction = "RECOMMEND_DISABLE_LEARNER"
	ActionHumanReview      HealthAction = "RECOMMEND_HUMAN_REVIEW"
	ActionContinueRollout  HealthAction = "RECOMMEND_CONTINUE_ROLLOUT"
)

// HealthRecommendation encapsulates a high-level governance imperative emitted by GovernanceBridge.
type HealthRecommendation struct {
	RecommendationID string       `json:"recommendation_id"`
	TargetSnapshotID string       `json:"target_snapshot_id,omitempty"`
	TargetLearnerID  string       `json:"target_learner_id,omitempty"`
	Action           HealthAction `json:"action"`
	Confidence       float64      `json:"confidence"`
	Rationale        string       `json:"rationale"`
}

// RolloutRecommendation encapsulates an infrastructure rollout routing advice for RolloutExecutor.
type RolloutRecommendation struct {
	SnapshotID      string             `json:"snapshot_id"`
	TargetLifecycle CandidateLifecycle `json:"target_lifecycle"`
	Reason          string             `json:"reason"`
}

// LearningService defines the public API and coordinator for IDUN's Learning ability.
// It is strictly offline, asynchronous, and non-episodesynchronous.
type LearningService interface {
	executive.LearningAbility

	// RunCycle executes a windowed learning cycle against aggregated historical experiences.
	RunCycle(ctx context.Context, req *LearningRequest) (*LearningResult, error)

	// RegisterLearner registers a signature-based Learner in the open LearnerRegistry.
	RegisterLearner(learner Learner) error

	// GetActiveSnapshot returns the currently active CandidateSnapshot for a domain schema.
	GetActiveSnapshot(ctx context.Context, schemaID string) (*CandidateSnapshot, error)

	// Start boots the Learning Service coordinator.
	Start() error

	// Close cleanly shuts down the Learning Service coordinator.
	Close() error
}

// Learner defines the open signature-based interface implemented by specialized learning algorithms.
type Learner interface {
	// LearnerID returns the unique identifier of this learning specialist.
	LearnerID() string

	// LearnerVersion returns the semantic version of this learning specialist algorithm.
	LearnerVersion() string

	// LearnerFingerprint returns the exact cryptographic fingerprint of this learner implementation.
	LearnerFingerprint() string

	// Consumes returns the slice of artifact schema IDs this learner requires as input.
	Consumes() []string

	// Produces returns the slice of snapshot schema IDs this learner synthesizes.
	Produces() []string

	// Generate synthesizes candidate snapshots from the aggregated historical window.
	// Phase 1 requires only skeleton implementations.
	Generate(ctx context.Context, summary *AggregationSummary) ([]*CandidateSnapshot, error)
}

// StrategyProvider defines the contract for accessing the active LearningStrategySnapshot atomically.
type StrategyProvider interface {
	// ActiveSnapshot returns the currently active strategy package without locking.
	ActiveSnapshot() *LearningStrategySnapshot
}

// SnapshotRegistry defines the storage and pointer management contract for immutable candidate snapshots.
type SnapshotRegistry interface {
	// Publish stores and registers a validated candidate snapshot.
	Publish(ctx context.Context, candidate *CandidateSnapshot) error

	// GetActive retrieves the live active candidate snapshot for the specified domain schema.
	GetActive(ctx context.Context, schemaID string) (*CandidateSnapshot, error)

	// Rollback atomically flips the active pointer back to an earlier target version.
	Rollback(ctx context.Context, schemaID string, targetVersion string) error
}

// ValidationPipeline defines the multi-stage gate that verifies candidate snapshots prior to publishing.
type ValidationPipeline interface {
	// ValidateCandidate performs statistical, minimum-sample, constitutional, and structural verification.
	ValidateCandidate(
		ctx context.Context,
		candidate *CandidateSnapshot,
		summary *AggregationSummary,
		profile *LearningPolicyProfile,
	) ([]ValidationResult, *StructuralValidationResult, error)
}

// RolloutExecutor defines the interface for the non-cognitive infrastructure engine owning live deployments.
type RolloutExecutor interface {
	// PromoteCandidate transitions a snapshot from Validated -> Shadow -> Canary -> Active.
	PromoteCandidate(ctx context.Context, snapshotID string, targetLifecycle CandidateLifecycle) error

	// GetStatus checks the current lifecycle state of a candidate snapshot.
	GetStatus(ctx context.Context, snapshotID string) (CandidateLifecycle, error)
}

// GovernanceBridge defines the interface for translating raw LearningDiagnostics into HealthRecommendations.
type GovernanceBridge interface {
	// EvaluateDiagnostics inspects diagnostic telemetry and produces high-level governance advice.
	EvaluateDiagnostics(ctx context.Context, diagnosticsRef string) (*HealthRecommendation, error)
}

// ArtifactValidator defines the contract enforcing the Generalized Artifact Rule.
type ArtifactValidator interface {
	// IsSupportedSchema returns true if the schemaID is recognized by a frozen runtime interpreter.
	IsSupportedSchema(schemaID string) bool

	// ValidatePayload verifies that the payload deserializes and complies with the schemaID specification.
	ValidatePayload(schemaID string, payload []byte) error
}

// ExperimentManager defines the interface for scheduling bounded shadow and canary experiments.
type ExperimentManager interface {
	// StartExperiment initiates a bounded shadow or A/B evaluation.
	StartExperiment(ctx context.Context, profile *ExperimentProfile) error

	// StopExperiment halts an active experiment cleanly.
	StopExperiment(ctx context.Context, experimentID string) error
}
