package reasoning

import (
	"context"

	"idun/intelligence/calibration"
	"idun/intelligence/communication"
)

// CalibrationSpecialist implements Stage S7 Calibration Integration.
// It delegates longitudinal confidence adjustment to idun/intelligence/calibration.
//
// MANDATORY INVARIANTS:
// 1. Enforces the Single-Owner Confidence Principle: Stage S7 is the ONLY writer of CalibratedConfidence.
// 2. NEVER duplicates historical calibration or trust estimation inside Reasoning.
type CalibrationSpecialist struct {
	calib calibration.CalibrationService
}

// NewCalibrationSpecialist returns an initialized CalibrationSpecialist.
func NewCalibrationSpecialist(calib calibration.CalibrationService) *CalibrationSpecialist {
	return &CalibrationSpecialist{calib: calib}
}

// ID returns StageS7Calibration.
func (s *CalibrationSpecialist) ID() StageIdentifier {
	return StageS7Calibration
}

// CalibrateHypotheses transforms ReasoningConfidence -> CalibratedConfidence via CalibrationService.
func (s *CalibrationSpecialist) CalibrateHypotheses(
	ctx context.Context,
	source string,
	topic communication.TopicID,
	primary ReasoningHypothesis,
	beam []ReasoningHypothesis,
) (ReasoningHypothesis, []ReasoningHypothesis, error) {
	if err := ctx.Err(); err != nil {
		return ReasoningHypothesis{}, nil, err
	}

	calPrimary := primary.Clone()
	if s.calib != nil {
		calPrimary.CalibratedConfidence = s.calib.CalibrateConfidence(source, topic, calPrimary.ReasoningConfidence)
	} else {
		calPrimary.CalibratedConfidence = calPrimary.ReasoningConfidence
	}
	calPrimary.ContributingStages = appendUniqueStage(calPrimary.ContributingStages, StageS7Calibration)

	calBeam := make([]ReasoningHypothesis, len(beam))
	for i, h := range beam {
		ch := h.Clone()
		if s.calib != nil {
			ch.CalibratedConfidence = s.calib.CalibrateConfidence(source, topic, ch.ReasoningConfidence)
		} else {
			ch.CalibratedConfidence = ch.ReasoningConfidence
		}
		ch.ContributingStages = appendUniqueStage(ch.ContributingStages, StageS7Calibration)
		calBeam[i] = ch
	}

	return calPrimary, calBeam, nil
}
