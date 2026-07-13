package reasoning

import (
	"context"
	"math"
	"testing"

	"idun/intelligence/calibration"
	"idun/intelligence/communication"
)

type mockCalibService struct {
	multiplier float64
}

func (m *mockCalibService) GetWeight(source string, topic communication.TopicID) float64 {
	return m.multiplier
}

func (m *mockCalibService) CalibrateConfidence(source string, topic communication.TopicID, rawConfidence float64) float64 {
	out := rawConfidence * m.multiplier
	if out > 1.0 {
		return 1.0
	}
	return out
}

func (m *mockCalibService) CalibrateEnvelope(env communication.Envelope, alpha, beta float64, totalBudget int) float64 {
	return 1.0
}

func (m *mockCalibService) RecordAudit(record calibration.AuditRecord) error {
	return nil
}

func (m *mockCalibService) SetWeightStrategy(strategy calibration.WeightStrategy) error {
	return nil
}

func (m *mockCalibService) GetSnapshot(source string, topic communication.TopicID) calibration.CalibrationSnapshot {
	return calibration.CalibrationSnapshot{}
}

func (m *mockCalibService) Name() string { return "mock.calib" }
func (m *mockCalibService) Start() error { return nil }
func (m *mockCalibService) Close() error { return nil }

func TestCalibrationSpecialist_CalibrateHypotheses(t *testing.T) {
	mock := &mockCalibService{multiplier: 0.90}
	specialist := NewCalibrationSpecialist(mock)

	if specialist.ID() != StageS7Calibration {
		t.Errorf("expected stage ID S7, got %s", specialist.ID())
	}

	primary := ReasoningHypothesis{
		ID:                  "primary-1",
		ReasoningConfidence: 0.80,
	}
	beam := []ReasoningHypothesis{
		{
			ID:                  "beam-1",
			ReasoningConfidence: 0.70,
		},
	}

	calPrimary, calBeam, err := specialist.CalibrateHypotheses(context.Background(), "intelligence.reasoning", communication.TopicActiveGoals, primary, beam)
	if err != nil {
		t.Fatalf("expected CalibrateHypotheses to succeed, got %v", err)
	}

	expectedPrimary := 0.80 * 0.90
	if math.Abs(calPrimary.CalibratedConfidence-expectedPrimary) > 1e-9 {
		t.Errorf("expected calibrated primary confidence %f, got %f", expectedPrimary, calPrimary.CalibratedConfidence)
	}

	expectedBeam := 0.70 * 0.90
	if len(calBeam) != 1 || math.Abs(calBeam[0].CalibratedConfidence-expectedBeam) > 1e-9 {
		t.Errorf("expected calibrated beam confidence %f, got %v", expectedBeam, calBeam)
	}
}
