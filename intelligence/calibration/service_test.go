package calibration_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"idun/intelligence/calibration"
	"idun/intelligence/communication"
)

type mockStrategy struct {
	fixedWeight float64
}

func (m *mockStrategy) ComputeWeight(source string, topic communication.TopicID, records []calibration.AuditRecord) float64 {
	return m.fixedWeight
}

func TestLifecycle(t *testing.T) {
	svc := calibration.NewService()
	if svc.Name() != "Intelligence.Calibration" {
		t.Fatalf("unexpected Name: %s", svc.Name())
	}
	if err := svc.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if err := svc.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	err := svc.RecordAudit(calibration.AuditRecord{
		Source:             "Reasoning",
		Topic:              communication.TopicCandidatePlans,
		ReportedConfidence: 0.9,
		ActualAccuracy:     0.9,
	})
	if !errors.Is(err, calibration.ErrServiceClosed) {
		t.Fatalf("expected ErrServiceClosed after close, got: %v", err)
	}
}

func TestDefaultWeight(t *testing.T) {
	svc := calibration.NewService()
	_ = svc.Start()
	defer svc.Close()

	w := svc.GetWeight("UnknownAbility", communication.TopicPerception)
	if w != 1.0 {
		t.Fatalf("expected default weight 1.0, got %f", w)
	}
}

func TestRecordAuditAndDiscounting(t *testing.T) {
	svc := calibration.NewService()
	_ = svc.Start()
	defer svc.Close()

	source := "OverconfidentNeuralDriver"
	topic := communication.TopicCandidatePlans

	// Record 5 audits where driver claims 0.95 but empirical accuracy is only 0.20
	for i := 0; i < 5; i++ {
		err := svc.RecordAudit(calibration.AuditRecord{
			Source:             source,
			Topic:              topic,
			ReportedConfidence: 0.95,
			ActualAccuracy:     0.20,
			Timestamp:          time.Now().UTC(),
		})
		if err != nil {
			t.Fatalf("RecordAudit failed: %v", err)
		}
	}

	w := svc.GetWeight(source, topic)
	if w >= 0.8 {
		t.Fatalf("expected calibration weight to discount overconfident module (<0.8), got %f", w)
	}

	calibrated := svc.CalibrateConfidence(source, topic, 0.90)
	if calibrated >= 0.70 {
		t.Fatalf("expected calibrated confidence discounted significantly, got %f", calibrated)
	}
}

func TestCalibrateEnvelope(t *testing.T) {
	svc := calibration.NewService()
	_ = svc.Start()
	defer svc.Close()

	source := "CalibratedReasoning"
	topic := communication.TopicEvaluatedOptions

	_ = svc.RecordAudit(calibration.AuditRecord{
		Source:             source,
		Topic:              topic,
		ReportedConfidence: 0.80,
		ActualAccuracy:     0.40, // 0.5 ratio
	})

	env, _ := communication.NewEnvelopeBuilder().
		WithSource(source).
		WithTopic(topic).
		WithPayloadRef("storage://options/1").
		WithConfidence(0.80).
		WithUrgency(0).
		WithCostEstimate(0).
		Build()

	peff := svc.CalibrateEnvelope(env, 0.1, 0.1, 1000)
	// Weight should be around 0.5 -> P_eff ~ 0.80 * 0.5 = 0.40
	if peff >= 0.75 {
		t.Fatalf("expected discounted EffectivePriority (<0.75), got %f", peff)
	}
}

func TestSetWeightStrategy(t *testing.T) {
	svc := calibration.NewService()
	_ = svc.Start()
	defer svc.Close()

	source := "Planning"
	topic := communication.TopicCandidatePlans
	_ = svc.RecordAudit(calibration.AuditRecord{
		Source:             source,
		Topic:              topic,
		ReportedConfidence: 0.9,
		ActualAccuracy:     0.9,
	})

	if err := svc.SetWeightStrategy(&mockStrategy{fixedWeight: 0.42}); err != nil {
		t.Fatalf("SetWeightStrategy failed: %v", err)
	}

	w := svc.GetWeight(source, topic)
	if w != 0.42 {
		t.Fatalf("expected updated weight 0.42 from strategy, got %f", w)
	}
}

func TestConcurrentCalibrationLoad(t *testing.T) {
	svc := calibration.NewService()
	_ = svc.Start()
	defer svc.Close()

	var wg sync.WaitGroup
	topics := communication.AllTopics()

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			topic := topics[idx%len(topics)]
			source := "ConcurrentDriver"
			_ = svc.RecordAudit(calibration.AuditRecord{
				Source:             source,
				Topic:              topic,
				ReportedConfidence: 0.8,
				ActualAccuracy:     0.7,
			})
			_ = svc.GetWeight(source, topic)
			_ = svc.CalibrateConfidence(source, topic, 0.75)
			_ = svc.GetSnapshot(source, topic)
		}(i)
	}

	wg.Wait()
}
