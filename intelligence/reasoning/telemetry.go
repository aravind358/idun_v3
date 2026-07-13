package reasoning

import (
	"sync"
	"sync/atomic"
)

// TelemetrySnapshot exposes operational statistics strictly to Host/Kernel monitors.
// Executive Functions and Cognitive Abilities MUST NEVER import or use TelemetrySnapshot for semantic reasoning.
type TelemetrySnapshot struct {
	TotalReasoningEpisodes  int64   `json:"total_reasoning_episodes"`
	SymbolicHits            int64   `json:"symbolic_hits"`
	GraphHits               int64   `json:"graph_hits"`
	BayesianHits            int64   `json:"bayesian_hits"`
	AnalogyHits             int64   `json:"analogy_hits"`
	DeliberativeEscalations int64   `json:"deliberative_escalations"`
	BeamSelections          int64   `json:"beam_selections"`
	CalibrationCalls        int64   `json:"calibration_calls"`
	ConstitutionEvaluations int64   `json:"constitution_evaluations"`
	AmbiguousBeamCount      int64   `json:"ambiguous_beam_count"`
	TimeoutCount            int64   `json:"timeout_count"`
	CancellationCount       int64   `json:"cancellation_count"`
	ValidationFailureCount  int64   `json:"validation_failure_count"`
	AvgReasoningDurationMs  float64 `json:"avg_reasoning_duration_ms"`
	AvgBeamWidth            float64 `json:"avg_beam_width"`
}

// telemetryCollector manages thread-safe collection of operational telemetry.
type telemetryCollector struct {
	totalEpisodes           int64
	symbolicHits            int64
	graphHits               int64
	bayesianHits            int64
	analogyHits             int64
	deliberativeEscalations int64
	beamSelections          int64
	calibrationCalls        int64
	constitutionEvaluations int64
	ambiguousBeamCount      int64
	timeoutCount            int64
	cancellationCount       int64
	validationFailureCount  int64

	mu                  sync.Mutex
	totalDurationMsSum  float64
	totalBeamWidthSum   float64
	episodesWithAvgCalc int64
}

func (tc *telemetryCollector) recordEpisode(
	durationMs float64,
	beamWidth int,
	symbolic bool,
	graph bool,
	bayesian bool,
	analogy bool,
	escalated bool,
	ambiguous bool,
	calibrated bool,
	constitutionEval bool,
) {
	atomic.AddInt64(&tc.totalEpisodes, 1)
	if symbolic {
		atomic.AddInt64(&tc.symbolicHits, 1)
	}
	if graph {
		atomic.AddInt64(&tc.graphHits, 1)
	}
	if bayesian {
		atomic.AddInt64(&tc.bayesianHits, 1)
	}
	if analogy {
		atomic.AddInt64(&tc.analogyHits, 1)
	}
	if escalated {
		atomic.AddInt64(&tc.deliberativeEscalations, 1)
	}
	atomic.AddInt64(&tc.beamSelections, 1)
	if ambiguous {
		atomic.AddInt64(&tc.ambiguousBeamCount, 1)
	}
	if calibrated {
		atomic.AddInt64(&tc.calibrationCalls, 1)
	}
	if constitutionEval {
		atomic.AddInt64(&tc.constitutionEvaluations, 1)
	}

	tc.mu.Lock()
	tc.totalDurationMsSum += durationMs
	tc.totalBeamWidthSum += float64(beamWidth)
	tc.episodesWithAvgCalc++
	tc.mu.Unlock()
}

func (tc *telemetryCollector) recordTimeout() {
	atomic.AddInt64(&tc.timeoutCount, 1)
}

func (tc *telemetryCollector) recordCancellation() {
	atomic.AddInt64(&tc.cancellationCount, 1)
}

func (tc *telemetryCollector) recordValidationFailure() {
	atomic.AddInt64(&tc.validationFailureCount, 1)
}

func (tc *telemetryCollector) snapshot() TelemetrySnapshot {
	snap := TelemetrySnapshot{
		TotalReasoningEpisodes:  atomic.LoadInt64(&tc.totalEpisodes),
		SymbolicHits:            atomic.LoadInt64(&tc.symbolicHits),
		GraphHits:               atomic.LoadInt64(&tc.graphHits),
		BayesianHits:            atomic.LoadInt64(&tc.bayesianHits),
		AnalogyHits:             atomic.LoadInt64(&tc.analogyHits),
		DeliberativeEscalations: atomic.LoadInt64(&tc.deliberativeEscalations),
		BeamSelections:          atomic.LoadInt64(&tc.beamSelections),
		CalibrationCalls:        atomic.LoadInt64(&tc.calibrationCalls),
		ConstitutionEvaluations: atomic.LoadInt64(&tc.constitutionEvaluations),
		AmbiguousBeamCount:      atomic.LoadInt64(&tc.ambiguousBeamCount),
		TimeoutCount:            atomic.LoadInt64(&tc.timeoutCount),
		CancellationCount:       atomic.LoadInt64(&tc.cancellationCount),
		ValidationFailureCount:  atomic.LoadInt64(&tc.validationFailureCount),
	}

	tc.mu.Lock()
	if tc.episodesWithAvgCalc > 0 {
		snap.AvgReasoningDurationMs = tc.totalDurationMsSum / float64(tc.episodesWithAvgCalc)
		snap.AvgBeamWidth = tc.totalBeamWidthSum / float64(tc.episodesWithAvgCalc)
	}
	tc.mu.Unlock()

	return snap
}
