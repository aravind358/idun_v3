// Package calibration implements the Epistemic Calibration System for IDUN V3
// Cognitive Communication & Executive Architecture Version 2.0.
//
// Architecture Version: 2.0.0-FROZEN
//
// The calibration package tracks historical epistemic accuracy per module source
// and workspace topic, discounting uncalibrated over-confidence without requiring
// Executive Functions to inspect semantic payloads.
package calibration

import (
	"errors"
	"time"

	"idun/intelligence/communication"
)

// Sentinel errors returned by CalibrationService methods.
var (
	ErrServiceClosed  = errors.New("calibration: service is closed")
	ErrInvalidSource  = errors.New("calibration: source identifier is required")
	ErrInvalidTopic   = errors.New("calibration: topic is invalid")
	ErrInvalidRecord  = errors.New("calibration: audit record accuracy/confidence out of range [0.0, 1.0]")
	ErrNilStrategy    = errors.New("calibration: weight strategy cannot be nil")
)

// AuditRecord records the historical comparison between a module's self-reported
// confidence and its actual empirical accuracy as audited by Reflection/Learning.
type AuditRecord struct {
	// Source identifies the cognitive ability or model driver.
	Source string

	// Topic specifies the leveled workspace channel audited.
	Topic communication.TopicID

	// ReportedConfidence records what the source claimed [0.0, 1.0].
	ReportedConfidence float64

	// ActualAccuracy records the audited empirical accuracy [0.0, 1.0].
	ActualAccuracy float64

	// Timestamp records when the audit occurred.
	Timestamp time.Time
}

// Validate verifies that an AuditRecord is structurally valid.
func (r AuditRecord) Validate() error {
	if r.Source == "" {
		return ErrInvalidSource
	}
	if !r.Topic.IsValid() {
		return ErrInvalidTopic
	}
	if r.ReportedConfidence < 0.0 || r.ReportedConfidence > 1.0 || r.ActualAccuracy < 0.0 || r.ActualAccuracy > 1.0 {
		return ErrInvalidRecord
	}
	return nil
}

// CalibrationSnapshot summarizes the operational calibration state for a source/topic pair.
type CalibrationSnapshot struct {
	Source      string
	Topic       communication.TopicID
	Weight      float64
	TotalAudits int64
	LastAudited time.Time
}

// WeightStrategy defines the pluggable algorithm for computing historical trust weights
// from accumulated audit records.
type WeightStrategy interface {
	// ComputeWeight calculates the calibration multiplier W_calib in [0.1, 1.5].
	ComputeWeight(source string, topic communication.TopicID, records []AuditRecord) float64
}

// CalibrationService defines the public capability contract injected into Executive
// Functions, Reflection, and Learning.
type CalibrationService interface {
	// GetWeight returns the current calibration multiplier W_calib for a source and topic.
	GetWeight(source string, topic communication.TopicID) float64

	// CalibrateConfidence returns calibrated confidence: rawConfidence * W_calib clamped to [0.0, 1.0].
	CalibrateConfidence(source string, topic communication.TopicID, rawConfidence float64) float64

	// CalibrateEnvelope computes the Calibrated Effective Priority (P_eff) for an Envelope.
	CalibrateEnvelope(env communication.Envelope, alpha, beta float64, totalBudget int) float64

	// RecordAudit records an empirical accuracy audit from Reflection or Learning.
	RecordAudit(record AuditRecord) error

	// SetWeightStrategy dynamically upgrades the weight calculation algorithm.
	SetWeightStrategy(strategy WeightStrategy) error

	// GetSnapshot retrieves the current calibration summary for a source and topic.
	GetSnapshot(source string, topic communication.TopicID) CalibrationSnapshot

	// Name returns the canonical Kernel component name ("Intelligence.Calibration").
	Name() string

	// Start boots the Calibration Service.
	Start() error

	// Close gracefully shuts down the Calibration Service.
	Close() error
}
