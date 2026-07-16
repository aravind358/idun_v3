package executive

import (
	"fmt"
	"time"
)

// ReplayMetadata records deterministic replay and provenance parameters so that any
// historical coordination decision can be reproduced or audited.
type ReplayMetadata struct {
	PolicyFingerprint        string    `json:"policy_fingerprint"`
	CapabilityFingerprint    string    `json:"capability_fingerprint"`
	ReplaySeed               uint64    `json:"replay_seed"`
	ExecutiveVersion         string    `json:"executive_version"`
	ConfigurationFingerprint string    `json:"configuration_fingerprint"`
	ReplayTimestamp          time.Time `json:"replay_timestamp"`
	ReplayFidelity           string    `json:"replay_fidelity"` // e.g., "EXACT", "BEST_EFFORT", "NOT_SUPPORTED"
}

// Validate checks whether required provenance fields exist and fidelity is valid.
func (r *ReplayMetadata) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: replay metadata is nil", ErrInvalidReplayMetadata)
	}
	if r.PolicyFingerprint == "" {
		return fmt.Errorf("%w: missing policy_fingerprint in replay metadata", ErrInvalidReplayMetadata)
	}
	if r.CapabilityFingerprint == "" {
		return fmt.Errorf("%w: missing capability_fingerprint in replay metadata", ErrInvalidReplayMetadata)
	}
	if r.ExecutiveVersion == "" {
		return fmt.Errorf("%w: missing executive_version in replay metadata", ErrInvalidReplayMetadata)
	}
	switch r.ReplayFidelity {
	case "EXACT", "BEST_EFFORT", "NOT_SUPPORTED", "":
		return nil
	default:
		return fmt.Errorf("%w: invalid replay_fidelity %q", ErrInvalidReplayMetadata, r.ReplayFidelity)
	}
}

// BudgetAllocationEvent records a factual budget allocation or escalation event.
type BudgetAllocationEvent struct {
	WorkflowID     string     `json:"workflow_id"`
	NodeID         string     `json:"node_id"`
	TierAssigned   BudgetTier `json:"tier_assigned"`
	UnitsAllocated int        `json:"units_allocated"`
	Timestamp      time.Time  `json:"timestamp"`
}

// PriorityChosenEvent records a factual prioritization band selection event.
type PriorityChosenEvent struct {
	WorkflowID  string       `json:"workflow_id"`
	StimulusID  string       `json:"stimulus_id"`
	BandChosen  PriorityBand `json:"band_chosen"`
	SalienceScore int        `json:"salience_score"`
	Timestamp   time.Time    `json:"timestamp"`
}

// CancellationEvent records a factual cooperative cancellation or preemption event.
type CancellationEvent struct {
	WorkflowID string    `json:"workflow_id"`
	Reason     string    `json:"reason"`
	Initiator  string    `json:"initiator"`
	Timestamp  time.Time `json:"timestamp"`
}

// WorkspaceArbitrationEvent records a factual Global Workspace competitive window decision.
type WorkspaceArbitrationEvent struct {
	Topic              string    `json:"topic"`
	WinningSource      string    `json:"winning_source"`
	WinningPriority    float64   `json:"winning_priority"`
	AdmissionThreshold float64   `json:"admission_threshold"`
	Admitted           bool      `json:"admitted"`
	ImpasseEmitted     bool      `json:"impasse_emitted"`
	Timestamp          time.Time `json:"timestamp"`
}

// ConstitutionInvocationEvent records a factual pre-broadcast action gate intercept.
type ConstitutionInvocationEvent struct {
	EnvelopeID   string    `json:"envelope_id"`
	Topic        string    `json:"topic"`
	Verdict      string    `json:"verdict"`
	RuleViolated string    `json:"rule_violated"`
	Timestamp    time.Time `json:"timestamp"`
}

// CalibrationInvocationEvent records a factual epistemic calibration computation.
type CalibrationInvocationEvent struct {
	Source               string    `json:"source"`
	Topic                string    `json:"topic"`
	RawConfidence        float64   `json:"raw_confidence"`
	CalibratedConfidence float64   `json:"calibrated_confidence"`
	Weight               float64   `json:"weight"`
	Timestamp            time.Time `json:"timestamp"`
}

// CoordinationTerminationSummary holds bounded statistical counts describing why coordination episodes terminated.
// It records pure observational facts without interpretation, ranking, or self-evaluation.
type CoordinationTerminationSummary struct {
	SuccessCount             uint64 `json:"success_count"`
	UserCancelledCount       uint64 `json:"user_cancelled_count"`
	InterruptedCount         uint64 `json:"interrupted_count"`
	TimeBudgetExceededCount  uint64 `json:"time_budget_exceeded_count"`
	DependencyFailureCount   uint64 `json:"dependency_failure_count"`
	ResourceExhaustedCount   uint64 `json:"resource_exhausted_count"`
	ConstitutionBlockedCount uint64 `json:"constitution_blocked_count"`
	ExecutiveAbortCount      uint64 `json:"executive_abort_count"`
}

// Validate implements the Validation Firewall for CoordinationTerminationSummary.
func (s *CoordinationTerminationSummary) Validate() error {
	if s == nil {
		return fmt.Errorf("%w: coordination termination summary is nil", ErrInvalidSummary)
	}
	return nil
}

// ExecutiveCoordinationSummary holds bounded numerical facts and counters about coordination performance.
// Like TraceStatisticalSummary in Learning or SearchStatistics in Planning, this struct records pure facts
// without self-evaluation or subjective interpretation, providing Reflection with raw data for metacognition.
type ExecutiveCoordinationSummary struct {
	EpisodesCoordinated         uint64                         `json:"episodes_coordinated"`
	AverageCoordinationDuration time.Duration                  `json:"average_coordination_duration"`
	TotalCoordinationDuration   time.Duration                  `json:"total_coordination_duration"`
	InterruptCount              uint64                         `json:"interrupt_count"`
	CancellationCount           uint64                         `json:"cancellation_count"`
	RetryCount                  uint64                         `json:"retry_count"`
	BudgetExhaustions           uint64                         `json:"budget_exhaustions"`
	SuccessfulCoordinations     uint64                         `json:"successful_coordinations"`
	FailedCoordinations         uint64                         `json:"failed_coordinations"`
	ImpasseEmittedCount         uint64                         `json:"impasse_emitted_count"`
	ConstitutionalBlocks        uint64                         `json:"constitutional_blocks"`
	TerminationSummary          CoordinationTerminationSummary `json:"termination_summary"`
}

// Validate checks non-negative duration bounds on ExecutiveCoordinationSummary.
func (s *ExecutiveCoordinationSummary) Validate() error {
	if s == nil {
		return fmt.Errorf("%w: coordination summary is nil", ErrInvalidSummary)
	}
	if s.AverageCoordinationDuration < 0 || s.TotalCoordinationDuration < 0 {
		return fmt.Errorf("%w: coordination durations cannot be negative", ErrInvalidSummary)
	}
	return s.TerminationSummary.Validate()
}

// ExecutiveTrace records factual coordination events during an episode.
// Executive strictly records coordination facts (budgets, priorities, cancellations, workspace/constitution/calibration calls)
// but never interprets or evaluates them; Reflection owns post-hoc analysis.
type ExecutiveTrace struct {
	TraceID               string                        `json:"trace_id"`
	EpisodeID             string                        `json:"episode_id"`
	SchemaVersion         string                        `json:"schema_version"`
	Timestamp             time.Time                     `json:"timestamp"`
	BudgetsAllocated      []BudgetAllocationEvent       `json:"budgets_allocated"`
	PriorityChosen        []PriorityChosenEvent         `json:"priority_chosen"`
	CancellationEvents    []CancellationEvent           `json:"cancellation_events"`
	WorkspaceArbitration  []WorkspaceArbitrationEvent   `json:"workspace_arbitration"`
	ConstitutionInvocation []ConstitutionInvocationEvent `json:"constitution_invocation"`
	CalibrationInvocation []CalibrationInvocationEvent  `json:"calibration_invocation"`
	ExecutionDuration     time.Duration                 `json:"execution_duration"`
	PolicyFingerprint     string                        `json:"policy_fingerprint"`
	CapabilityFingerprint string                        `json:"capability_fingerprint"`
	ExecutiveVersion      string                        `json:"executive_version"`
	ReplayMetadata        ReplayMetadata                `json:"replay_metadata"`
}

// Validate implements the Validation Firewall for ExecutiveTrace.
func (t *ExecutiveTrace) Validate() error {
	if t == nil {
		return fmt.Errorf("%w: trace object is nil", ErrInvalidTrace)
	}
	if t.TraceID == "" || t.EpisodeID == "" {
		return fmt.Errorf("%w: missing required trace_id or episode_id", ErrInvalidTrace)
	}
	if t.SchemaVersion != SchemaVersion {
		return fmt.Errorf("%w: schema_version %q != required %q", ErrInvalidTrace, t.SchemaVersion, SchemaVersion)
	}
	if t.PolicyFingerprint == "" || t.CapabilityFingerprint == "" || t.ExecutiveVersion == "" {
		return fmt.Errorf("%w: missing mandatory provenance fingerprints or version", ErrInvalidTrace)
	}
	if t.ExecutionDuration < 0 {
		return fmt.Errorf("%w: execution_duration cannot be negative", ErrInvalidTrace)
	}
	if err := t.ReplayMetadata.Validate(); err != nil {
		return err
	}
	return nil
}

// ExecutiveRequest encapsulates a formal request to coordinate a cognitive workflow or episode.
type ExecutiveRequest struct {
	RequestID         string            `json:"request_id"`
	EpisodeID         string            `json:"episode_id"`
	Workflow          *WorkflowGraph    `json:"workflow,omitempty"`
	Horizon           Horizon           `json:"horizon"`
	Priority          PriorityBand      `json:"priority"`
	Budget            BudgetTier        `json:"budget"`
	ActiveGoal        ActiveGoalContext `json:"active_goal"`
	PolicyFingerprint string            `json:"policy_fingerprint,omitempty"`
}

// Validate implements the Validation Firewall for ExecutiveRequest.
func (q *ExecutiveRequest) Validate() error {
	if q == nil {
		return fmt.Errorf("%w: request object is nil", ErrInvalidResult)
	}
	if q.RequestID == "" || q.EpisodeID == "" {
		return fmt.Errorf("%w: missing required request_id or episode_id in request", ErrInvalidResult)
	}
	if q.Priority < PriorityBand0CriticalSafety || q.Priority > PriorityBand4Idle {
		return fmt.Errorf("%w: invalid priority band %d in request", ErrInvalidResult, q.Priority)
	}
	if q.Budget < BudgetReflexive || q.Budget > BudgetDeliberative {
		return fmt.Errorf("%w: invalid budget tier %d in request", ErrInvalidResult, q.Budget)
	}
	return nil
}

// ExecutiveResult represents the complete factual output of an episode coordination attempt,
// separating status ("What happened?") from termination reason ("Why execution terminated?").
type ExecutiveResult struct {
	EpisodeID             string                         `json:"episode_id"`
	WorkflowID            string                         `json:"workflow_id"`
	Status                ExecutiveResultStatus          `json:"status"`
	TerminationReason     ExecutiveTerminationReason     `json:"termination_reason"`
	OutputRef             string                         `json:"output_ref"`
	Error                 error                          `json:"-"`
	ErrorString           string                         `json:"error_string,omitempty"`
	CoordinationSummary   ExecutiveCoordinationSummary   `json:"coordination_summary"`
	TerminationSummary    CoordinationTerminationSummary `json:"termination_summary"`
	TraceID               string                         `json:"trace_id"`
	PolicyFingerprint     string                         `json:"policy_fingerprint"`
	CapabilityFingerprint string                         `json:"capability_fingerprint"`
	ExecutiveVersion      string                         `json:"executive_version"`
	ReplayMetadata        ReplayMetadata                 `json:"replay_metadata"`
}

// Validate implements the Validation Firewall for ExecutiveResult.
func (r *ExecutiveResult) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: result object is nil", ErrInvalidResult)
	}
	if r.EpisodeID == "" || r.WorkflowID == "" {
		return fmt.Errorf("%w: missing required episode_id or workflow_id", ErrInvalidResult)
	}
	if err := r.Status.Validate(); err != nil {
		return err
	}
	if err := r.TerminationReason.Validate(); err != nil {
		return err
	}
	if err := r.CoordinationSummary.Validate(); err != nil {
		return err
	}
	if err := r.TerminationSummary.Validate(); err != nil {
		return err
	}
	if r.PolicyFingerprint == "" || r.CapabilityFingerprint == "" || r.ExecutiveVersion == "" {
		return fmt.Errorf("%w: missing mandatory provenance fingerprints or version in result", ErrInvalidResult)
	}
	if err := r.ReplayMetadata.Validate(); err != nil {
		return err
	}
	return nil
}

