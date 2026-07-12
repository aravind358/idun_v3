package understanding

import (
	"sync/atomic"
)

// TelemetrySnapshot captures operational diagnostics and performance metrics for the
// Understanding subsystem. This snapshot is intended for Host/Kernel monitoring and
// must never be inspected by content-blind Executive Functions.
type TelemetrySnapshot struct {
	TotalInterpretations    int64   `json:"TotalInterpretations"`
	GrammarHits             int64   `json:"GrammarHits"`
	NeuralHits              int64   `json:"NeuralHits"`
	DeliberativeEscalations int64   `json:"DeliberativeEscalations"`
	AmbiguousBeamCount      int64   `json:"AmbiguousBeamCount"`
	TimeoutCount            int64   `json:"TimeoutCount"`
	CancellationCount       int64   `json:"CancellationCount"`
	MalformedInferenceCount int64   `json:"MalformedInferenceCount"`
	ValidationFailureCount  int64   `json:"ValidationFailureCount"`
	AvgProcessingDurationMs float64 `json:"AvgProcessingDurationMs"`
	AvgBeamWidth            float64 `json:"AvgBeamWidth"`
}

type telemetryCollector struct {
	totalInterpretations    atomic.Int64
	grammarHits             atomic.Int64
	neuralHits              atomic.Int64
	deliberativeEscalations atomic.Int64
	ambiguousBeamCount      atomic.Int64
	timeoutCount            atomic.Int64
	cancellationCount       atomic.Int64
	malformedInferenceCount atomic.Int64
	validationFailureCount  atomic.Int64
	totalDurationUs         atomic.Int64
	totalBeamWidthSum       atomic.Int64
}

func (t *telemetryCollector) recordInterpretation(durationUs int64, beamWidth int, isAmbiguous bool) {
	t.totalInterpretations.Add(1)
	t.totalDurationUs.Add(durationUs)
	t.totalBeamWidthSum.Add(int64(beamWidth))
	if isAmbiguous {
		t.ambiguousBeamCount.Add(1)
	}
}

func (t *telemetryCollector) recordGrammarHit() {
	t.grammarHits.Add(1)
}

func (t *telemetryCollector) recordNeuralHit() {
	t.neuralHits.Add(1)
}

func (t *telemetryCollector) recordDeliberativeEscalation() {
	t.deliberativeEscalations.Add(1)
}

func (t *telemetryCollector) recordTimeout() {
	t.timeoutCount.Add(1)
}

func (t *telemetryCollector) recordCancellation() {
	t.cancellationCount.Add(1)
}

func (t *telemetryCollector) recordMalformedInference() {
	t.malformedInferenceCount.Add(1)
}

func (t *telemetryCollector) recordValidationFailure() {
	t.validationFailureCount.Add(1)
}

func (t *telemetryCollector) snapshot() TelemetrySnapshot {
	total := t.totalInterpretations.Load()
	var avgDurationMs float64
	var avgBeam float64
	if total > 0 {
		avgDurationMs = float64(t.totalDurationUs.Load()) / float64(total) / 1000.0
		avgBeam = float64(t.totalBeamWidthSum.Load()) / float64(total)
	}
	return TelemetrySnapshot{
		TotalInterpretations:    total,
		GrammarHits:             t.grammarHits.Load(),
		NeuralHits:              t.neuralHits.Load(),
		DeliberativeEscalations: t.deliberativeEscalations.Load(),
		AmbiguousBeamCount:      t.ambiguousBeamCount.Load(),
		TimeoutCount:            t.timeoutCount.Load(),
		CancellationCount:       t.cancellationCount.Load(),
		MalformedInferenceCount: t.malformedInferenceCount.Load(),
		ValidationFailureCount:  t.validationFailureCount.Load(),
		AvgProcessingDurationMs: avgDurationMs,
		AvgBeamWidth:            avgBeam,
	}
}
