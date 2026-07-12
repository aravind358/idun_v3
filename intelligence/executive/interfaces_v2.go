// Package executive implements IDUN V3 Executive Functions Architecture Version 2.0.
//
// Architecture Version: 2.0.0-FROZEN
//
// Executive Version 2.0 extends Executive V1 with content-blind Global Workspace
// arbitration, Epistemic Calibration discounting, Pre-Broadcast Constitutional
// integration, Multi-Horizon scheduling, and automatic Impasse emission.
package executive

import (
	"context"
	"errors"

	"idun/intelligence/calibration"
	"idun/intelligence/communication"
	"idun/intelligence/constitution"
	"idun/intelligence/workspace"
)

// Sentinel errors returned by Executive Version 2.0 operations.
var (
	ErrExecutiveV2Closed = errors.New("executive_v2: service is closed")
	ErrNilWorkspace      = errors.New("executive_v2: workspace cannot be nil")
	ErrNilCalibration    = errors.New("executive_v2: calibration service cannot be nil")
	ErrNilConstitution   = errors.New("executive_v2: constitutional gate cannot be nil")
	ErrInvalidBid        = errors.New("executive_v2: candidate bid envelope is invalid")
)

// Horizon identifies the scheduling horizon for a candidate workflow or bid.
type Horizon int

const (
	// HorizonReflexive (<15ms) handles physical safety and subsumption reflexes.
	HorizonReflexive Horizon = iota

	// HorizonDeliberative (100ms-500ms) handles continuous competitive workspace arbitration.
	HorizonDeliberative

	// HorizonBackground (Minutes/Hours) handles long-running asynchronous consolidation.
	HorizonBackground
)

// String returns the human-readable name of the Horizon.
func (h Horizon) String() string {
	switch h {
	case HorizonReflexive:
		return "REFLEXIVE"
	case HorizonDeliberative:
		return "DELIBERATIVE"
	case HorizonBackground:
		return "BACKGROUND"
	default:
		return "UNKNOWN"
	}
}

// ArbiterDecision reports the result of a competitive Global Workspace window arbitration.
type ArbiterDecision struct {
	// Admitted reports whether a candidate bid won and was published.
	Admitted bool

	// Winner contains the winning Envelope (if Admitted is true).
	Winner communication.Envelope

	// EffectivePriority reports the P_eff score of the winning bid.
	EffectivePriority float64

	// ImpasseEmitted reports whether an impasse event was emitted because no bid met admission threshold.
	ImpasseEmitted bool

	// Reason explains the decision or impasse rationale.
	Reason string
}

// ExecutiveV2 defines the composite capability interface of Executive Functions Version 2.0.
type ExecutiveV2 interface {
	Executive // Inherits all Version 1 capabilities

	// Workspace returns the integrated Global Workspace & Leveled Blackboard engine.
	Workspace() workspace.Workspace

	// Calibration returns the integrated Epistemic Calibration service.
	Calibration() calibration.CalibrationService

	// Constitution returns the integrated Pre-Broadcast Constitutional Action Gate.
	Constitution() constitution.ActionGate

	// SubmitBid submits a candidate Envelope bid into the specified Horizon queue for arbitration.
	SubmitBid(ctx context.Context, env communication.Envelope, horizon Horizon) error

	// ArbitrateCompetition resolves competition among pending bids on a leveled topic channel.
	// Uses Calibrated Effective Priority (P_eff). If no bid crosses admissionThreshold, emits TopicImpasses.
	ArbitrateCompetition(ctx context.Context, topic communication.TopicID, admissionThreshold float64) (ArbiterDecision, error)

	// RemainingBudgetUnits returns the currently available computational cost units.
	RemainingBudgetUnits() int

	// ConsumeBudget deducts computational cost units from the current cycle budget.
	ConsumeBudget(units int) error
}
