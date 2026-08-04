package v3

import (
	"errors"
	"fmt"
	"idun/core/foundation"
)

var ErrValidation = errors.New("decision validation error")

const SpecVersion = "3.0"

// ResolutionStatus represents the final decision outcome.
type ResolutionStatus string

const (
	StatusApproved ResolutionStatus = "APPROVED"
	StatusRejected ResolutionStatus = "REJECTED"
	StatusDeferred ResolutionStatus = "DEFERRED"
)

// DecisionFinding represents a machine-readable diagnostic for a policy/safety/auth check.
type DecisionFinding struct {
	validatorType string // e.g. "Safety", "Policy", "Auth", "Budget"
	nodeID        string // The execution node that triggered the finding
	passed        bool
	code          string // A machine-readable violation code, e.g. "ERR_NO_AUTH"
	message       string // Human-readable rationale
}

func NewDecisionFinding(vType, nodeID string, passed bool, code, message string) DecisionFinding {
	return DecisionFinding{
		validatorType: vType,
		nodeID:        nodeID,
		passed:        passed,
		code:          code,
		message:       message,
	}
}

func (f DecisionFinding) ValidatorType() string { return f.validatorType }
func (f DecisionFinding) NodeID() string        { return f.nodeID }
func (f DecisionFinding) Passed() bool          { return f.passed }
func (f DecisionFinding) Code() string          { return f.code }
func (f DecisionFinding) Message() string       { return f.message }

// DecisionRecord represents the immutable output of the Decision layer.
type DecisionRecord struct {
	// Lineage
	specVersion      string
	artifactID       foundation.ArtifactID
	parentArtifactID foundation.ParentArtifactID
	envelopeID       foundation.EnvelopeID
	timestamp        foundation.Timestamp

	// Evaluation Outcome
	resolution        ResolutionStatus
	reason            string // Overall human-readable rationale
	safetyPassed      bool
	policyPassed      bool
	budgetPassed      bool
	permissionsPassed bool
	effectivePermissions []string
	findings          []DecisionFinding
}

// CognitiveArtifact implementation
func (d *DecisionRecord) IsImmutable() bool                             { return true }
func (d *DecisionRecord) ArtifactID() foundation.ArtifactID             { return d.artifactID }
func (d *DecisionRecord) ParentArtifactID() foundation.ParentArtifactID { return d.parentArtifactID }
func (d *DecisionRecord) EnvelopeID() foundation.EnvelopeID             { return d.envelopeID }
func (d *DecisionRecord) Timestamp() foundation.Timestamp               { return d.timestamp }
func (d *DecisionRecord) Version() foundation.Version                   { return foundation.Version(d.specVersion) }

// Field Getters
func (d *DecisionRecord) Resolution() ResolutionStatus { return d.resolution }
func (d *DecisionRecord) Reason() string               { return d.reason }
func (d *DecisionRecord) SafetyPassed() bool           { return d.safetyPassed }
func (d *DecisionRecord) PolicyPassed() bool           { return d.policyPassed }
func (d *DecisionRecord) BudgetPassed() bool           { return d.budgetPassed }
func (d *DecisionRecord) PermissionsPassed() bool      { return d.permissionsPassed }
func (d *DecisionRecord) EffectivePermissions() []string {
	cp := make([]string, len(d.effectivePermissions))
	copy(cp, d.effectivePermissions)
	return cp
}
func (d *DecisionRecord) Findings() []DecisionFinding {
	cp := make([]DecisionFinding, len(d.findings))
	copy(cp, d.findings)
	return cp
}

// Validate enforces core foundation lineage rules.
func (d *DecisionRecord) Validate() error {
	if d.specVersion != SpecVersion {
		return fmt.Errorf("%w: expected %q, got %q", ErrValidation, SpecVersion, d.specVersion)
	}
	if d.artifactID == "" {
		return fmt.Errorf("%w: ArtifactID cannot be empty", ErrValidation)
	}
	if d.parentArtifactID == "" {
		return fmt.Errorf("%w: ParentArtifactID cannot be empty", ErrValidation)
	}
	if d.envelopeID == "" {
		return fmt.Errorf("%w: EnvelopeID cannot be empty", ErrValidation)
	}
	if d.resolution != StatusApproved && d.resolution != StatusRejected && d.resolution != StatusDeferred {
		return fmt.Errorf("%w: invalid resolution status %q", ErrValidation, d.resolution)
	}
	return nil
}
