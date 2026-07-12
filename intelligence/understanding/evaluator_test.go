package understanding_test

import (
	"context"
	"sync"
	"testing"

	"idun/intelligence/calibration"
	"idun/intelligence/communication"
	"idun/intelligence/understanding"
)

type mockCalibrator struct {
	mu     sync.Mutex
	weight float64
}

func (m *mockCalibrator) GetWeight(source string, topic communication.TopicID) float64 {
	return m.weight
}

func (m *mockCalibrator) CalibrateConfidence(source string, topic communication.TopicID, raw float64) float64 {
	return raw * m.weight
}

func (m *mockCalibrator) CalibrateEnvelope(env communication.Envelope, alpha, beta float64, totalBudget int) float64 {
	return env.RawConfidence * m.weight
}

func (m *mockCalibrator) RecordAudit(record calibration.AuditRecord) error            { return nil }
func (m *mockCalibrator) SetWeightStrategy(strategy calibration.WeightStrategy) error { return nil }
func (m *mockCalibrator) GetSnapshot(source string, topic communication.TopicID) calibration.CalibrationSnapshot {
	return calibration.CalibrationSnapshot{}
}
func (m *mockCalibrator) Name() string { return "MockCalibrator" }
func (m *mockCalibrator) Start() error { return nil }
func (m *mockCalibrator) Close() error { return nil }

type mockMultiNeuralSpecialist struct{}

func (m *mockMultiNeuralSpecialist) Evaluate(norm understanding.NormalizedText, boundSlots []understanding.Slot) ([]understanding.Hypothesis, error) {
	return []understanding.Hypothesis{
		{Intent: "intent_a", CalibratedConfidence: 0.85, SourceLayer: understanding.LayerNeuralClassifier},
		{Intent: "intent_b", CalibratedConfidence: 0.82, SourceLayer: understanding.LayerNeuralClassifier},
		{Intent: "intent_c", CalibratedConfidence: 0.80, SourceLayer: understanding.LayerNeuralClassifier},
		{Intent: "intent_d", CalibratedConfidence: 0.78, SourceLayer: understanding.LayerNeuralClassifier},
		{Intent: "intent_e", CalibratedConfidence: 0.40, SourceLayer: understanding.LayerNeuralClassifier},
	}, nil
}

func TestSlotAwareMergeHypothesesByIntent(t *testing.T) {
	candidates := []understanding.Hypothesis{
		{
			Intent:               "book_flight",
			CalibratedConfidence: 0.90,
			SourceLayer:          understanding.LayerReflexiveGrammar,
			Slots: []understanding.Slot{
				{Name: "destination", Value: "SEA", Confidence: 0.95, GroundingID: "geo-sea"},
				{Name: "date", Value: "tomorrow", Confidence: 0.80, GroundingID: "time-tom"},
			},
		},
		{
			Intent:               "book_flight",
			CalibratedConfidence: 0.88,
			SourceLayer:          understanding.LayerNeuralClassifier,
			Slots: []understanding.Slot{
				{Name: "airline", Value: "Alaska", Confidence: 0.92, GroundingID: "air-alaska"}, // Complementary slot
				{Name: "date", Value: "2026-07-13", Confidence: 0.96, GroundingID: "time-iso"},  // Conflicting slot with higher confidence
			},
		},
	}

	merged := understanding.MergeHypothesesByIntent(candidates)
	if len(merged) != 1 {
		t.Fatalf("expected 1 merged hypothesis, got %d", len(merged))
	}

	win := merged[0]
	if win.CalibratedConfidence != 0.90 {
		t.Fatalf("expected winning base confidence 0.90, got %f", win.CalibratedConfidence)
	}
	if len(win.Slots) != 3 {
		t.Fatalf("expected 3 merged slots (airline, date, destination), got %d: %+v", len(win.Slots), win.Slots)
	}

	slotMap := make(map[string]understanding.Slot)
	for _, s := range win.Slots {
		slotMap[s.Name] = s
	}

	// 1. Verify complementary slot was preserved
	if slotMap["airline"].Value != "Alaska" {
		t.Fatalf("expected complementary airline slot preserved, got %+v", slotMap["airline"])
	}
	// 2. Verify conflicting slot resolved to higher confidence slot
	if slotMap["date"].Value != "2026-07-13" || slotMap["date"].Confidence != 0.96 {
		t.Fatalf("expected date slot resolved to higher confidence 2026-07-13 (0.96), got %+v", slotMap["date"])
	}
}

func TestDefaultNeuralSpecialist(t *testing.T) {
	neural := understanding.NewDefaultNeuralSpecialist()
	normalizer := understanding.NewDefaultNormalizer()

	norm := normalizer.Normalize("Can we reschedule the meeting?")
	hyps, err := neural.Evaluate(norm, nil)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}

	if len(hyps) == 0 {
		t.Fatalf("expected neural hypothesis for reschedule meeting")
	}
	found := false
	for _, h := range hyps {
		if h.Intent == "reschedule_meeting" && h.SourceLayer == understanding.LayerNeuralClassifier {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected reschedule_meeting intent, got: %+v", hyps)
	}
}

func TestBoundedBeamSelectionAndPruning(t *testing.T) {
	evaluator := understanding.NewSpeculativeEvaluator(0.15)
	multiNeural := &mockMultiNeuralSpecialist{}

	ctx := context.Background()
	primary, ambiguitySet, err := evaluator.EvaluateParallel(ctx, understanding.NormalizedText{}, nil, nil, multiNeural, nil)
	if err != nil {
		t.Fatalf("EvaluateParallel failed: %v", err)
	}

	if primary.Intent != "intent_a" {
		t.Fatalf("expected primary intent_a, got %s", primary.Intent)
	}

	if len(ambiguitySet) != 2 {
		t.Fatalf("expected exactly 2 runner-up hypotheses in AmbiguitySet (beam width 3), got %d: %+v", len(ambiguitySet), ambiguitySet)
	}
	if ambiguitySet[0].Intent != "intent_b" || ambiguitySet[1].Intent != "intent_c" {
		t.Fatalf("unexpected ambiguity set: %+v", ambiguitySet)
	}
}

func TestCalibrationIntegration(t *testing.T) {
	evaluator := understanding.NewSpeculativeEvaluator(0.15)
	neural := understanding.NewDefaultNeuralSpecialist()
	calibrator := &mockCalibrator{weight: 0.5}

	normalizer := understanding.NewDefaultNormalizer()
	norm := normalizer.Normalize("Can we reschedule the meeting?")

	primary, _, err := evaluator.EvaluateParallel(context.Background(), norm, nil, nil, neural, calibrator)
	if err != nil {
		t.Fatalf("EvaluateParallel failed: %v", err)
	}

	if primary.CalibratedConfidence > 0.50 {
		t.Fatalf("expected calibrated confidence down-weighted below 0.50, got %f", primary.CalibratedConfidence)
	}
}

func TestServicePhase3SpeculativeInterpretation(t *testing.T) {
	svc := understanding.NewService(
		understanding.WithConfigOptions(),
		nil,
		understanding.WithNeuralSpecialist(&mockMultiNeuralSpecialist{}),
	)

	frame, err := svc.InterpretEnvelope(context.Background(), communication.Envelope{
		ID:         "spec-env",
		PayloadRef: "test multi neural utterance",
	})
	if err != nil {
		t.Fatalf("InterpretEnvelope failed: %v", err)
	}

	if frame.Status != understanding.StatusAmbiguousBeam {
		t.Fatalf("expected AMBIGUOUS_BEAM status when runner-up hypotheses exist, got %s", frame.Status)
	}
	if frame.PrimaryHypothesis.Intent != "intent_a" {
		t.Fatalf("expected primary hypothesis intent_a, got %s", frame.PrimaryHypothesis.Intent)
	}
	if len(frame.AmbiguitySet) != 2 {
		t.Fatalf("expected 2 ambiguity hypotheses, got %d", len(frame.AmbiguitySet))
	}
}
