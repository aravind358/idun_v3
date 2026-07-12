// Package constitution implements the Pre-Broadcast Constitutional Action Gate
// for IDUN V3 Cognitive Communication & Executive Architecture Version 2.0.
//
// Architecture Version: 2.0.0-FROZEN
//
// The constitution package enforces invariant safety, ethical boundaries, and
// constitutional compliance by intercepting external world-modifying actions
// before physical/actuator broadcast.
package constitution

import (
	"context"
	"errors"
	"time"

	"idun/intelligence/communication"
	"idun/intelligence/workspace"
)

// Sentinel errors returned by Constitutional Action Gate methods.
var (
	ErrGateClosed       = errors.New("constitution: action gate is closed")
	ErrInvalidEnvelope  = errors.New("constitution: invalid candidate envelope")
	ErrNilRule          = errors.New("constitution: rule cannot be nil")
	ErrRuleIDMissing    = errors.New("constitution: rule ID is required")
	ErrDuplicateRule    = errors.New("constitution: rule ID is already registered")
	ErrNilWorkspace     = errors.New("constitution: workspace cannot be nil")
	ErrActionVetoed     = errors.New("constitution: action vetoed by constitutional rule")
	ErrActionEscalation = errors.New("constitution: action escalated for user confirmation")
)

// Verdict defines the discrete decision returned by a Constitutional Rule or ActionGate.
type Verdict int

const (
	// VerdictApproved indicates the action satisfies all constitutional invariants.
	VerdictApproved Verdict = iota

	// VerdictVetoed indicates the action violates invariant safety or ethical boundaries.
	VerdictVetoed

	// VerdictEscalateToUser indicates ambiguity or high risk requiring explicit user confirmation.
	VerdictEscalateToUser
)

// String returns the canonical human-readable representation of Verdict.
func (v Verdict) String() string {
	switch v {
	case VerdictApproved:
		return "APPROVED"
	case VerdictVetoed:
		return "VETOED"
	case VerdictEscalateToUser:
		return "ESCALATE_TO_USER"
	default:
		return "UNKNOWN"
	}
}

// EvaluationResult summarizes the constitutional decision for a candidate envelope.
type EvaluationResult struct {
	// EnvelopeID identifies the evaluated envelope.
	EnvelopeID string

	// Verdict reports the decision (Approved, Vetoed, EscalateToUser).
	Verdict Verdict

	// Reason explains the rationale for Veto or Escalation decisions.
	Reason string

	// RuleViolated records the Rule.ID() that triggered a Veto or Escalation (empty if Approved).
	RuleViolated string

	// Signature is the cryptographic or HMAC approval token (empty unless VerdictApproved).
	Signature string

	// EvaluatedAt records UTC evaluation timestamp.
	EvaluatedAt time.Time
}

// Rule defines an independent constitutional invariant check.
type Rule interface {
	// ID returns the unique canonical rule identifier.
	ID() string

	// Description returns human-readable documentation of what this invariant enforces.
	Description() string

	// Evaluate inspects the candidate Envelope and returns a Verdict and rationale.
	Evaluate(ctx context.Context, env communication.Envelope) (Verdict, string, error)
}

// ActionGate defines the capability contract for the Pre-Broadcast Constitutional Action Gate.
type ActionGate interface {
	// EvaluateAction evaluates all registered rules against a candidate Envelope.
	EvaluateAction(ctx context.Context, env communication.Envelope) (EvaluationResult, error)

	// InterceptAndPublish evaluates a candidate action Envelope:
	// - If Approved: signs the envelope and publishes it to TopicActionExecution.
	// - If Vetoed: blocks the action and publishes a high-urgency alert to TopicValueFlags.
	// - If EscalateToUser: blocks the action and publishes an inquiry to TopicUserIntent.
	InterceptAndPublish(ctx context.Context, env communication.Envelope, ws workspace.Workspace) (EvaluationResult, error)

	// RegisterRule registers a new invariant constitutional rule.
	RegisterRule(rule Rule) error

	// ListRules returns all currently registered rule IDs.
	ListRules() []string

	// Name returns the canonical Kernel component name ("Intelligence.Constitution").
	Name() string

	// Start boots the Constitutional Action Gate.
	Start() error

	// Close gracefully shuts down the Action Gate.
	Close() error
}
