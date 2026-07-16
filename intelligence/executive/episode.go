package executive

import (
	"errors"
	"time"
)

// Sentinel validation errors for Episode domain artifacts.
var (
	ErrNilEpisodeDefinition = errors.New("executive_episode: definition artifact is nil")
	ErrNilEpisodeRuntime    = errors.New("executive_episode: runtime artifact is nil")
	ErrNilEpisodeCheckpoint = errors.New("executive_episode: checkpoint artifact is nil")
	ErrInvalidEpisodeID     = errors.New("executive_episode: missing or invalid episode_id")
	ErrInvalidLifecycle     = errors.New("executive_episode: invalid status or illegal transition")
	ErrInvalidOutcome       = errors.New("executive_episode: invalid or incompatible outcome")
	ErrInvalidFingerprint   = errors.New("executive_episode: fingerprint mismatch or tampering detected")
	ErrInvalidReference     = errors.New("executive_episode: mandatory core reference URI is missing or malformed")
)

// EpisodeID is a unique identifier for an episode (UUIDv4 or SHA-256).
type EpisodeID string

// EpisodeType categorizes the execution mechanism of the episode.
type EpisodeType string

const (
	EpisodeTypeCognitiveTurn    EpisodeType = "COGNITIVE_TURN"
	EpisodeTypeHomeostatic      EpisodeType = "HOMEOSTATIC"
	EpisodeTypeSkillAcquisition EpisodeType = "SKILL_ACQUISITION"
	EpisodeTypeMultiAgent       EpisodeType = "MULTI_AGENT"
	EpisodeTypeBackground       EpisodeType = "BACKGROUND"
)

// EpisodeIntent categorizes the semantic objective of the episode.
type EpisodeIntent string

const (
	EpisodeIntentConversation          EpisodeIntent = "CONVERSATION"
	EpisodeIntentPlanning              EpisodeIntent = "PLANNING"
	EpisodeIntentLearning              EpisodeIntent = "LEARNING"
	EpisodeIntentReflection            EpisodeIntent = "REFLECTION"
	EpisodeIntentBackgroundMaintenance EpisodeIntent = "BACKGROUND_MAINTENANCE"
	EpisodeIntentToolExecution         EpisodeIntent = "TOOL_EXECUTION"
	EpisodeIntentSkillTraining         EpisodeIntent = "SKILL_TRAINING"
	EpisodeIntentVisionProcessing      EpisodeIntent = "VISION_PROCESSING"
	EpisodeIntentStrategyActivation    EpisodeIntent = "STRATEGY_ACTIVATION"
)

// EpisodeOrigin records the factual initiator of the episode.
type EpisodeOrigin string

const (
	EpisodeOriginUser          EpisodeOrigin = "USER"
	EpisodeOriginExecutive     EpisodeOrigin = "EXECUTIVE"
	EpisodeOriginAttention     EpisodeOrigin = "ATTENTION"
	EpisodeOriginPlanning      EpisodeOrigin = "PLANNING"
	EpisodeOriginReflection    EpisodeOrigin = "REFLECTION"
	EpisodeOriginLearning      EpisodeOrigin = "LEARNING"
	EpisodeOriginScheduler     EpisodeOrigin = "SCHEDULER"
	EpisodeOriginAPI           EpisodeOrigin = "API"
	EpisodeOriginExternalEvent EpisodeOrigin = "EXTERNAL_EVENT"
)

// EpisodeStatus defines the bounded FSM lifecycle states of an episode.
// It describes what execution state the episode is currently in.
type EpisodeStatus string

const (
	EpisodeStatusCreated   EpisodeStatus = "CREATED"
	EpisodeStatusWaiting   EpisodeStatus = "WAITING"
	EpisodeStatusRunning   EpisodeStatus = "RUNNING"
	EpisodeStatusPaused    EpisodeStatus = "PAUSED"
	EpisodeStatusSuspended EpisodeStatus = "SUSPENDED"
	EpisodeStatusCancelled EpisodeStatus = "CANCELLED"
	EpisodeStatusCompleted EpisodeStatus = "COMPLETED"
	EpisodeStatusFailed    EpisodeStatus = "FAILED"
)

// EpisodeOutcome describes the final execution result of an episode.
// It is strictly orthogonal to EpisodeStatus.
type EpisodeOutcome string

const (
	EpisodeOutcomePending   EpisodeOutcome = "PENDING"
	EpisodeOutcomeSuccess   EpisodeOutcome = "SUCCESS"
	EpisodeOutcomeFailed    EpisodeOutcome = "FAILED"
	EpisodeOutcomeAbandoned EpisodeOutcome = "ABANDONED"
	EpisodeOutcomeEscalated EpisodeOutcome = "ESCALATED"
	EpisodeOutcomeCancelled EpisodeOutcome = "CANCELLED"
)

// ============================================================================
// Strongly Typed Transition Enums (Refinement 4)
// ============================================================================

// PriorityTransitionReason records the factual, typed reason for a priority change.
type PriorityTransitionReason string

const (
	PriorityReasonSalienceOverride PriorityTransitionReason = "SALIENCE_OVERRIDE"
	PriorityReasonEscalation       PriorityTransitionReason = "ESCALATION"
	PriorityReasonUserOverride     PriorityTransitionReason = "USER_OVERRIDE"
	PriorityReasonDefault          PriorityTransitionReason = "DEFAULT_ASSIGNMENT"
	PriorityReasonHomeostatic      PriorityTransitionReason = "HOMEOSTATIC_ADJUSTMENT"
)

// BudgetTransitionReason records the factual, typed reason for a budget tier change.
type BudgetTransitionReason string

const (
	BudgetReasonInitialAssignment   BudgetTransitionReason = "INITIAL_ASSIGNMENT"
	BudgetReasonEscalationGranted   BudgetTransitionReason = "ESCALATION_GRANTED"
	BudgetReasonHomeostaticDecrease BudgetTransitionReason = "HOMEOSTATIC_DECREASE"
	BudgetReasonHomeostaticIncrease BudgetTransitionReason = "HOMEOSTATIC_INCREASE"
)

// PauseReason records the factual, typed reason for pausing an episode.
type PauseReason string

const (
	PauseReasonAwaitingExternalInput PauseReason = "AWAITING_EXTERNAL_INPUT"
	PauseReasonAwaitingDependency    PauseReason = "AWAITING_DEPENDENCY"
	PauseReasonUserRequest           PauseReason = "USER_REQUEST"
	PauseReasonCheckpointing         PauseReason = "CHECKPOINTING"
	PauseReasonResourcePreemption    PauseReason = "RESOURCE_PREEMPTION"
)

// ResumeReason records the factual, typed reason for resuming a paused episode.
type ResumeReason string

const (
	ResumeReasonInputReceived      ResumeReason = "INPUT_RECEIVED"
	ResumeReasonDependencyResolved ResumeReason = "DEPENDENCY_RESOLVED"
	ResumeReasonUserRequest        ResumeReason = "USER_REQUEST"
	ResumeReasonSchedulerWake      ResumeReason = "SCHEDULER_WAKE"
)

// CheckpointReason records the factual, typed reason for generating an EpisodeCheckpoint.
type CheckpointReason string

const (
	CheckpointReasonPeriodic       CheckpointReason = "PERIODIC_SNAPSHOT"
	CheckpointReasonPreMigration   CheckpointReason = "PRE_MIGRATION"
	CheckpointReasonPause          CheckpointReason = "PAUSE_SNAPSHOT"
	CheckpointReasonSystemShutdown CheckpointReason = "SYSTEM_SHUTDOWN"
	CheckpointReasonManual         CheckpointReason = "MANUAL_REQUEST"
)

// ============================================================================
// Capabilities & Context (Refinements 2 & 3)
// ============================================================================

// EpisodeCapabilities advertises the orchestration features supported by this deployment.
type EpisodeCapabilities struct {
	SupportsPause           bool `json:"supports_pause"`
	SupportsResume          bool `json:"supports_resume"`
	SupportsCheckpoint      bool `json:"supports_checkpoint"`
	SupportsBackground      bool `json:"supports_background"`
	SupportsChildren        bool `json:"supports_children"`
	SupportsMigration       bool `json:"supports_migration"`
	SupportsRemoteExecution bool `json:"supports_remote_execution"`
}

// EpisodeContext implements the Hybrid model (Strongly Typed Core + Extensible Module Registry).
// Executive depends directly on WorkspaceReference, AttentionReference, and GoalReference.
// All optional cognitive/embodied modules are registered via ModuleReferences without changing public API.
type EpisodeContext struct {
	// Strongly Typed Core References (Mandatory for Executive coordination)
	WorkspaceReference string `json:"workspace_reference"`
	AttentionReference string `json:"attention_reference"`
	GoalReference      string `json:"goal_reference"`

	// Extensible Module Registry (e.g., "planning" -> "planning://...", "vision" -> "vision://...")
	ModuleReferences map[string]string `json:"module_references,omitempty"`
}

// Validate ensures all mandatory core references are present and well-formed.
func (ec *EpisodeContext) Validate() error {
	if ec.WorkspaceReference == "" || ec.AttentionReference == "" || ec.GoalReference == "" {
		return ErrInvalidReference
	}
	return nil
}

// ============================================================================
// Episode Definition & Runtime Artifacts
// ============================================================================

// ExecutiveEpisodeDefinition defines the immutable, content-blind identity and configuration of an episode.
type ExecutiveEpisodeDefinition struct {
	EpisodeID           EpisodeID           `json:"episode_id"`
	EpisodeType         EpisodeType         `json:"episode_type"`
	EpisodeIntent       EpisodeIntent       `json:"episode_intent"`
	EpisodeOrigin       EpisodeOrigin       `json:"episode_origin"`
	ParentEpisodeID     EpisodeID           `json:"parent_episode_id,omitempty"`
	RootEpisodeID       EpisodeID           `json:"root_episode_id"`
	SchemaVersion       string              `json:"schema_version"`       // "2.0.0"
	HierarchyReference  string              `json:"hierarchy_reference"`  // Reference to Workspace graph
	DependencyReference string              `json:"dependency_reference"` // Reference to Workspace DAG
	ContextReference    EpisodeContext      `json:"context_reference"`
	Capabilities        EpisodeCapabilities `json:"capabilities"`
	EpisodeFingerprint  string              `json:"episode_fingerprint"` // SHA-256 over immutable fields
	ReplayMetadata      ReplayMetadata      `json:"replay_metadata"`
}

// Validate verifies the immutable definition firewall rules.
func (d *ExecutiveEpisodeDefinition) Validate() error {
	if d == nil {
		return ErrNilEpisodeDefinition
	}
	if d.EpisodeID == "" {
		return ErrInvalidEpisodeID
	}
	if d.EpisodeFingerprint == "" {
		return ErrInvalidFingerprint
	}
	return d.ContextReference.Validate()
}

// PriorityTransition records a factual priority change with typed enum reason.
type PriorityTransition struct {
	FromPriority PriorityBand             `json:"from_priority"`
	ToPriority   PriorityBand             `json:"to_priority"`
	Reason       PriorityTransitionReason `json:"reason"`
	Timestamp    time.Time                `json:"timestamp"`
}

// BudgetTransition records a factual budget change with typed enum reason.
type BudgetTransition struct {
	FromBudget BudgetTier             `json:"from_budget"`
	ToBudget   BudgetTier             `json:"to_budget"`
	Reason     BudgetTransitionReason `json:"reason"`
	Timestamp  time.Time              `json:"timestamp"`
}

// SubsystemExecutionMetric records bounded statistical invocation facts for a single subsystem.
type SubsystemExecutionMetric struct {
	Invoked       bool          `json:"invoked"`
	ExecutionTime time.Duration `json:"execution_time"`
	Success       bool          `json:"success"`
	Skipped       bool          `json:"skipped,omitempty"`
	SkipReason    string        `json:"skip_reason,omitempty"`
}

// EpisodeSubsystemUsage records purely observational, bounded telemetry about subsystem invocations.
type EpisodeSubsystemUsage struct {
	Planning   SubsystemExecutionMetric `json:"planning"`
	Reasoning  SubsystemExecutionMetric `json:"reasoning"`
	Workspace  SubsystemExecutionMetric `json:"workspace"`
	Decision   SubsystemExecutionMetric `json:"decision"`
	Reflection SubsystemExecutionMetric `json:"reflection"`
}

// EpisodeTerminationSummary records self-contained, purely observational termination facts.
type EpisodeTerminationSummary struct {
	Completed           bool `json:"completed"`
	Cancelled           bool `json:"cancelled"`
	Interrupted         bool `json:"interrupted"`
	ConstitutionBlocked bool `json:"constitution_blocked"`
	DependencyFailure   bool `json:"dependency_failure"`
	Timeout             bool `json:"timeout"`
	BudgetExhausted     bool `json:"budget_exhausted"`
}

// ExecutiveEpisodeRuntime records the mutable execution lifecycle, rolling histories, and telemetry.
type ExecutiveEpisodeRuntime struct {
	EpisodeID          EpisodeID                 `json:"episode_id"`
	Status             EpisodeStatus             `json:"status"`
	Outcome            EpisodeOutcome            `json:"outcome"`
	CurrentPriority    PriorityBand              `json:"current_priority"`
	CurrentHorizon     int                       `json:"current_horizon"`
	CurrentBudget      BudgetTier                `json:"current_budget"`
	RemainingCostUnits int                       `json:"remaining_cost_units"`
	PriorityHistory    []PriorityTransition      `json:"priority_history"` // Bounded rolling history (max 16)
	BudgetHistory      []BudgetTransition        `json:"budget_history"`   // Bounded rolling history (max 16)
	PauseReason        PauseReason               `json:"pause_reason,omitempty"`
	ResumeReason       ResumeReason              `json:"resume_reason,omitempty"`
	ExecutorID         string                    `json:"executor_id"` // Node/device executing the episode
	MigrationCount     int                       `json:"migration_count,omitempty"`
	LastMigratedAt     *time.Time                `json:"last_migrated_at,omitempty"`
	CreatedAt          time.Time                 `json:"created_at"`
	UpdatedAt          time.Time                 `json:"updated_at"`
	CompletedAt        *time.Time                `json:"completed_at,omitempty"`
	SubsystemUsage     EpisodeSubsystemUsage     `json:"subsystem_usage"`
	TerminationSummary EpisodeTerminationSummary `json:"termination_summary"`
}

// Validate verifies runtime FSM and outcome compatibility firewall rules.
func (r *ExecutiveEpisodeRuntime) Validate() error {
	if r == nil {
		return ErrNilEpisodeRuntime
	}
	if r.EpisodeID == "" {
		return ErrInvalidEpisodeID
	}
	switch r.Status {
	case EpisodeStatusCreated, EpisodeStatusWaiting, EpisodeStatusRunning, EpisodeStatusPaused, EpisodeStatusSuspended:
		if r.Outcome != EpisodeOutcomePending {
			return ErrInvalidOutcome
		}
	case EpisodeStatusCompleted:
		if r.Outcome != EpisodeOutcomeSuccess && r.Outcome != EpisodeOutcomeAbandoned && r.Outcome != EpisodeOutcomeEscalated {
			return ErrInvalidOutcome
		}
	case EpisodeStatusFailed:
		if r.Outcome != EpisodeOutcomeFailed {
			return ErrInvalidOutcome
		}
	case EpisodeStatusCancelled:
		if r.Outcome != EpisodeOutcomeCancelled && r.Outcome != EpisodeOutcomeAbandoned {
			return ErrInvalidOutcome
		}
	default:
		return ErrInvalidLifecycle
	}
	return nil
}

// ExecutiveEpisode groups an immutable definition with its mutable runtime execution state.
type ExecutiveEpisode struct {
	Definition *ExecutiveEpisodeDefinition `json:"definition"`
	Runtime    *ExecutiveEpisodeRuntime    `json:"runtime"`
}

// Validate checks both definition and runtime invariants.
func (e *ExecutiveEpisode) Validate() error {
	if e == nil {
		return errors.New("executive_episode: artifact is nil")
	}
	if err := e.Definition.Validate(); err != nil {
		return err
	}
	if err := e.Runtime.Validate(); err != nil {
		return err
	}
	if e.Definition.EpisodeID != e.Runtime.EpisodeID {
		return errors.New("executive_episode: episode_id mismatch between definition and runtime")
	}
	return nil
}

// ============================================================================
// Minimal Immutable Checkpoint Artifact (Refinement 5)
// ============================================================================

// EpisodeCheckpoint defines a minimal, immutable recovery and migration snapshot of an episode.
// It is NOT the episode itself nor runtime state; it is a frozen recovery artifact.
type EpisodeCheckpoint struct {
	CheckpointID       string         `json:"checkpoint_id"`
	EpisodeID          EpisodeID      `json:"episode_id"`
	RuntimeFingerprint string         `json:"runtime_fingerprint"`
	WorkspaceReference string         `json:"workspace_reference"`
	AttentionReference string         `json:"attention_reference"`
	Timestamp          time.Time      `json:"timestamp"`
	ReplayMetadata     ReplayMetadata `json:"replay_metadata"`
}

// Validate enforces immutability and completeness firewall checks for the checkpoint.
func (cp *EpisodeCheckpoint) Validate() error {
	if cp == nil {
		return ErrNilEpisodeCheckpoint
	}
	if cp.CheckpointID == "" || cp.EpisodeID == "" {
		return ErrInvalidEpisodeID
	}
	if cp.RuntimeFingerprint == "" {
		return ErrInvalidFingerprint
	}
	if cp.WorkspaceReference == "" || cp.AttentionReference == "" {
		return ErrInvalidReference
	}
	return nil
}
