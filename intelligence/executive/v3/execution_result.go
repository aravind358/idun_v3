package v3

import (
	"idun/core/foundation"
	"time"
)

// ExecutionStatus represents the overall completion state of a DAG execution.
type ExecutionStatus string

const (
	StatusCompleted  ExecutionStatus = "COMPLETED"
	StatusFailed     ExecutionStatus = "FAILED"
	StatusTimedOut   ExecutionStatus = "TIMED_OUT"
	StatusCancelled  ExecutionStatus = "CANCELLED"
	StatusImpasse    ExecutionStatus = "IMPASSE"
)

// NodeStatus represents the execution state of an individual node.
type NodeStatus string

const (
	NodeCompleted NodeStatus = "COMPLETED"
	NodeFailed    NodeStatus = "FAILED"
	NodeSkipped   NodeStatus = "SKIPPED"
)

// NodeResult captures the physical execution outcome for a specific capability node.
type NodeResult struct {
	NodeID    string        `json:"node_id"`
	Status    NodeStatus    `json:"status"`
	Duration  time.Duration `json:"duration"`
	OutputRef string        `json:"output_ref,omitempty"`
	Error     string        `json:"error,omitempty"`
	Metadata  foundation.InteractionMetadata `json:"metadata"`
}

// ExecutionResult represents the physical runtime outcome of executing a committed DecisionRecord.
// It is strictly immutable and content-blind.
type ExecutionResult struct {
	artifactID       foundation.ArtifactID
	parentArtifactID foundation.ArtifactID
	envelopeID       string
	timestamp        time.Time
	version          string

	status        ExecutionStatus
	nodeResults   map[string]NodeResult
	totalDuration time.Duration
	overallError  string
}

// ArtifactID returns the unique UUID for this execution result.
func (e *ExecutionResult) ArtifactID() foundation.ArtifactID { return e.artifactID }

// ParentArtifactID returns the DecisionRecord ID that authorized this execution.
func (e *ExecutionResult) ParentArtifactID() foundation.ArtifactID { return e.parentArtifactID }

// EnvelopeID returns the original threaded Envelope ID.
func (e *ExecutionResult) EnvelopeID() string { return e.envelopeID }

// Timestamp returns the time this result was finalized.
func (e *ExecutionResult) Timestamp() time.Time { return e.timestamp }

// Version returns the schema version.
func (e *ExecutionResult) Version() string { return e.version }

// Status returns the overall execution status.
func (e *ExecutionResult) Status() ExecutionStatus { return e.status }

// NodeResults returns a copy of the individual node outcomes to maintain immutability.
func (e *ExecutionResult) NodeResults() map[string]NodeResult {
	cp := make(map[string]NodeResult, len(e.nodeResults))
	for k, v := range e.nodeResults {
		cp[k] = v
	}
	return cp
}

// TotalDuration returns the wall-clock time spent executing the DAG.
func (e *ExecutionResult) TotalDuration() time.Duration { return e.totalDuration }

// OverallError returns the high-level error if the DAG failed to execute completely.
func (e *ExecutionResult) OverallError() string { return e.overallError }
