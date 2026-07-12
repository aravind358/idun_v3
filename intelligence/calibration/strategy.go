package calibration

import (
	"math"

	"idun/intelligence/communication"
)

const (
	minCalibrationWeight = 0.1
	maxCalibrationWeight = 1.5
	defaultWeight        = 1.0
)

// DefaultWeightStrategy implements a robust Exponential Moving Ratio strategy
// that penalizes systematic over-confidence while rewarding conservative accuracy.
type DefaultWeightStrategy struct{}

// NewDefaultWeightStrategy constructs the default Epistemic Calibration weight strategy.
func NewDefaultWeightStrategy() *DefaultWeightStrategy {
	return &DefaultWeightStrategy{}
}

// ComputeWeight calculates the historical calibration trust multiplier W_calib in [0.1, 1.5].
func (s *DefaultWeightStrategy) ComputeWeight(source string, topic communication.TopicID, records []AuditRecord) float64 {
	if len(records) == 0 {
		return defaultWeight
	}

	// Calculate exponentially decayed calibration ratio over recent records.
	// Recent audits carry higher weight than historical audits.
	var weightedSumRatio float64
	var totalWeight float64
	decay := 0.85

	for i := len(records) - 1; i >= 0; i-- {
		rec := records[i]
		rep := rec.ReportedConfidence
		if rep <= 0.05 {
			rep = 0.05 // prevent division by near-zero
		}

		ratio := rec.ActualAccuracy / rep
		factor := math.Pow(decay, float64(len(records)-1-i))
		weightedSumRatio += ratio * factor
		totalWeight += factor
	}

	if totalWeight == 0 {
		return defaultWeight
	}

	weight := weightedSumRatio / totalWeight
	if weight < minCalibrationWeight {
		weight = minCalibrationWeight
	} else if weight > maxCalibrationWeight {
		weight = maxCalibrationWeight
	}

	return weight
}

// Ensure DefaultWeightStrategy implements WeightStrategy.
var _ WeightStrategy = (*DefaultWeightStrategy)(nil)
