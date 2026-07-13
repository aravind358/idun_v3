package reflection

import (
	"sync/atomic"
)

// TelemetrySnapshot captures a content-blind operational metrics snapshot for the Reflection subsystem.
type TelemetrySnapshot struct {
	TotalEpisodeReflections    uint64  `json:"total_episode_reflections"`
	TotalPeriodicReflections   uint64  `json:"total_periodic_reflections"`
	TotalMetaReflections       uint64  `json:"total_meta_reflections"`
	SpecialistEvaluations      uint64  `json:"specialist_evaluations"`
	TrendAnalyses              uint64  `json:"trend_analyses"`
	CrossCognitiveAnalyses     uint64  `json:"cross_cognitive_analyses"`
	GrowthPotentialEstimations uint64  `json:"growth_potential_estimations"`
	SelfCalibrationRuns        uint64  `json:"self_calibration_runs"`
	ValidationFailures         uint64  `json:"validation_failures"`
	AbstainCount               uint64  `json:"abstain_count"`
	TimeoutCount               uint64  `json:"timeout_count"`
	CancellationCount          uint64  `json:"cancellation_count"`
	AvgEpisodeDurationMs       float64 `json:"avg_episode_duration_ms"`
	AvgPeriodicDurationMs      float64 `json:"avg_periodic_duration_ms"`
	AvgMetaDurationMs          float64 `json:"avg_meta_duration_ms"`
}

// telemetryCollector holds lock-free atomic counters for operational monitoring.
type telemetryCollector struct {
	totalEpisodeReflections    atomic.Uint64
	totalPeriodicReflections   atomic.Uint64
	totalMetaReflections       atomic.Uint64
	specialistEvaluations      atomic.Uint64
	trendAnalyses              atomic.Uint64
	crossCognitiveAnalyses     atomic.Uint64
	growthPotentialEstimations atomic.Uint64
	selfCalibrationRuns        atomic.Uint64
	validationFailures         atomic.Uint64
	abstainCount               atomic.Uint64
	timeoutCount               atomic.Uint64
	cancellationCount          atomic.Uint64

	totalEpisodeDurationNs  atomic.Int64
	totalPeriodicDurationNs atomic.Int64
	totalMetaDurationNs     atomic.Int64
}

// snapshot returns an immutable snapshot of all operational telemetry counters.
func (tc *telemetryCollector) snapshot() TelemetrySnapshot {
	epCount := tc.totalEpisodeReflections.Load()
	perCount := tc.totalPeriodicReflections.Load()
	metaCount := tc.totalMetaReflections.Load()

	var avgEpMs, avgPerMs, avgMetaMs float64
	if epCount > 0 {
		avgEpMs = float64(tc.totalEpisodeDurationNs.Load()) / (float64(epCount) * 1e6)
	}
	if perCount > 0 {
		avgPerMs = float64(tc.totalPeriodicDurationNs.Load()) / (float64(perCount) * 1e6)
	}
	if metaCount > 0 {
		avgMetaMs = float64(tc.totalMetaDurationNs.Load()) / (float64(metaCount) * 1e6)
	}

	return TelemetrySnapshot{
		TotalEpisodeReflections:    epCount,
		TotalPeriodicReflections:   perCount,
		TotalMetaReflections:       metaCount,
		SpecialistEvaluations:      tc.specialistEvaluations.Load(),
		TrendAnalyses:              tc.trendAnalyses.Load(),
		CrossCognitiveAnalyses:     tc.crossCognitiveAnalyses.Load(),
		GrowthPotentialEstimations: tc.growthPotentialEstimations.Load(),
		SelfCalibrationRuns:        tc.selfCalibrationRuns.Load(),
		ValidationFailures:         tc.validationFailures.Load(),
		AbstainCount:               tc.abstainCount.Load(),
		TimeoutCount:               tc.timeoutCount.Load(),
		CancellationCount:          tc.cancellationCount.Load(),
		AvgEpisodeDurationMs:       avgEpMs,
		AvgPeriodicDurationMs:      avgPerMs,
		AvgMetaDurationMs:          avgMetaMs,
	}
}
