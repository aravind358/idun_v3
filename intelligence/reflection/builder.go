package reflection

import (
	"fmt"
)

// ReflectionReportBuilder constructs a validated immutable ReflectionReport.
type ReflectionReportBuilder struct {
	report ReflectionReport
}

// NewReflectionReportBuilder initializes a new builder for ReflectionReport.
func NewReflectionReportBuilder(reportID string, mode ReflectionMode, timestamp int64) *ReflectionReportBuilder {
	return &ReflectionReportBuilder{
		report: ReflectionReport{
			SchemaVersion:              SchemaVersion,
			ReportID:                   reportID,
			Timestamp:                  timestamp,
			Mode:                       mode,
			SpecialistReports:          make([]SpecialistReport, 0),
			CrossCognitiveFindings:     make([]CrossCognitiveFinding, 0),
			GrowthPotentialEstimates:   make([]GrowthPotentialEstimate, 0),
			TrendFindings:              make([]TrendFinding, 0),
			RecommendedLearningSignals: make([]RecommendedLearningSignal, 0),
			SessionNotes:               make([]string, 0),
		},
	}
}

// WithEpisodeID sets the episode ID (required when Mode == ModeEpisode).
func (b *ReflectionReportBuilder) WithEpisodeID(episodeID string) *ReflectionReportBuilder {
	b.report.EpisodeID = episodeID
	return b
}

// WithSummaryID sets the historical summary ID (required when Mode == ModePeriodic).
func (b *ReflectionReportBuilder) WithSummaryID(summaryID string) *ReflectionReportBuilder {
	b.report.SummaryID = summaryID
	return b
}

// WithTimeWindow sets the evaluated time window specification.
func (b *ReflectionReportBuilder) WithTimeWindow(start, end int64) *ReflectionReportBuilder {
	b.report.TimeWindow = &TimeWindowSpec{StartTime: start, EndTime: end}
	return b
}

// AddSpecialistReport adds a specialist report finding.
func (b *ReflectionReportBuilder) AddSpecialistReport(sr SpecialistReport) *ReflectionReportBuilder {
	b.report.SpecialistReports = append(b.report.SpecialistReports, sr)
	return b
}

// AddCrossCognitiveFinding adds a cross-cognitive interaction finding.
func (b *ReflectionReportBuilder) AddCrossCognitiveFinding(cf CrossCognitiveFinding) *ReflectionReportBuilder {
	b.report.CrossCognitiveFindings = append(b.report.CrossCognitiveFindings, cf)
	return b
}

// AddGrowthPotentialEstimate adds a growth potential estimate.
func (b *ReflectionReportBuilder) AddGrowthPotentialEstimate(gp GrowthPotentialEstimate) *ReflectionReportBuilder {
	b.report.GrowthPotentialEstimates = append(b.report.GrowthPotentialEstimates, gp)
	return b
}

// AddTrendFinding adds a longitudinal trend finding.
func (b *ReflectionReportBuilder) AddTrendFinding(tf TrendFinding) *ReflectionReportBuilder {
	b.report.TrendFindings = append(b.report.TrendFindings, tf)
	return b
}

// AddRecommendedLearningSignal adds a recommended learning signal.
func (b *ReflectionReportBuilder) AddRecommendedLearningSignal(rls RecommendedLearningSignal) *ReflectionReportBuilder {
	b.report.RecommendedLearningSignals = append(b.report.RecommendedLearningSignals, rls)
	return b
}

// AddSessionNote appends a session note.
func (b *ReflectionReportBuilder) AddSessionNote(note string) *ReflectionReportBuilder {
	b.report.SessionNotes = append(b.report.SessionNotes, note)
	return b
}

// WithSelfCalibration sets the self calibration report.
func (b *ReflectionReportBuilder) WithSelfCalibration(sc *SelfCalibrationReport) *ReflectionReportBuilder {
	b.report.SelfCalibration = sc
	return b
}

// Build validates and returns the immutable ReflectionReport.
func (b *ReflectionReportBuilder) Build() (ReflectionReport, error) {
	if err := b.report.Validate(); err != nil {
		return ReflectionReport{}, fmt.Errorf("reflection report builder validation failed: %w", err)
	}
	return b.report.Clone(), nil
}

// HistoricalSummaryBuilder constructs a validated HistoricalSummary.
type HistoricalSummaryBuilder struct {
	summary HistoricalSummary
}

// NewHistoricalSummaryBuilder initializes a new builder for HistoricalSummary.
func NewHistoricalSummaryBuilder(summaryID string, timestamp int64) *HistoricalSummaryBuilder {
	return &HistoricalSummaryBuilder{
		summary: HistoricalSummary{
			SchemaVersion:      SchemaVersion,
			SummaryID:          summaryID,
			GeneratedTimestamp: timestamp,
			AverageScores:      make(map[string]float64),
			TrendMetrics:       make(map[string]float64),
			FailureRates:       make(map[string]float64),
			ImprovementRates:   make(map[string]float64),
			SummaryConfidence:  1.0,
		},
	}
}

// WithTimeWindow sets the time window specification.
func (b *HistoricalSummaryBuilder) WithTimeWindow(start, end int64) *HistoricalSummaryBuilder {
	b.summary.TimeWindow = TimeWindowSpec{StartTime: start, EndTime: end}
	return b
}

// WithEpisodeCount sets the episode count.
func (b *HistoricalSummaryBuilder) WithEpisodeCount(count int64) *HistoricalSummaryBuilder {
	b.summary.EpisodeCount = count
	return b
}

// WithSummaryConfidence sets the summary confidence.
func (b *HistoricalSummaryBuilder) WithSummaryConfidence(conf float64) *HistoricalSummaryBuilder {
	b.summary.SummaryConfidence = conf
	return b
}

// AddAverageScore sets an average score for a metric.
func (b *HistoricalSummaryBuilder) AddAverageScore(metric string, score float64) *HistoricalSummaryBuilder {
	b.summary.AverageScores[metric] = score
	return b
}

// AddFailureRate sets a failure rate for a metric.
func (b *HistoricalSummaryBuilder) AddFailureRate(metric string, rate float64) *HistoricalSummaryBuilder {
	b.summary.FailureRates[metric] = rate
	return b
}

// AddTrendMetric sets a trend metric trajectory.
func (b *HistoricalSummaryBuilder) AddTrendMetric(metric string, val float64) *HistoricalSummaryBuilder {
	b.summary.TrendMetrics[metric] = val
	return b
}

// AddImprovementRate sets an improvement rate velocity.
func (b *HistoricalSummaryBuilder) AddImprovementRate(metric string, rate float64) *HistoricalSummaryBuilder {
	b.summary.ImprovementRates[metric] = rate
	return b
}

// Build validates and returns the HistoricalSummary.
func (b *HistoricalSummaryBuilder) Build() (HistoricalSummary, error) {
	if err := b.summary.Validate(); err != nil {
		return HistoricalSummary{}, fmt.Errorf("historical summary builder validation failed: %w", err)
	}
	return b.summary.Clone(), nil
}

// HistoricalSummaryRequestBuilder constructs a validated HistoricalSummaryRequest.
type HistoricalSummaryRequestBuilder struct {
	req HistoricalSummaryRequest
}

// NewHistoricalSummaryRequestBuilder initializes a new builder for HistoricalSummaryRequest.
func NewHistoricalSummaryRequestBuilder(requestID string) *HistoricalSummaryRequestBuilder {
	return &HistoricalSummaryRequestBuilder{
		req: HistoricalSummaryRequest{
			SchemaVersion:    SchemaVersion,
			RequestID:        requestID,
			TargetAbilities:  make([]string, 0),
			RequestedMetrics: make([]string, 0),
		},
	}
}

// WithTimeWindow sets the requested time window.
func (b *HistoricalSummaryRequestBuilder) WithTimeWindow(start, end int64) *HistoricalSummaryRequestBuilder {
	b.req.TimeWindow = TimeWindowSpec{StartTime: start, EndTime: end}
	return b
}

// AddTargetAbility appends a target ability to the request.
func (b *HistoricalSummaryRequestBuilder) AddTargetAbility(ability string) *HistoricalSummaryRequestBuilder {
	b.req.TargetAbilities = append(b.req.TargetAbilities, ability)
	return b
}

// WithAggregationLevel sets the aggregation granularity level.
func (b *HistoricalSummaryRequestBuilder) WithAggregationLevel(level string) *HistoricalSummaryRequestBuilder {
	b.req.AggregationLevel = level
	return b
}

// AddRequestedMetric appends a requested metric name.
func (b *HistoricalSummaryRequestBuilder) AddRequestedMetric(metric string) *HistoricalSummaryRequestBuilder {
	b.req.RequestedMetrics = append(b.req.RequestedMetrics, metric)
	return b
}

// Build validates and returns the HistoricalSummaryRequest.
func (b *HistoricalSummaryRequestBuilder) Build() (HistoricalSummaryRequest, error) {
	if err := b.req.Validate(); err != nil {
		return HistoricalSummaryRequest{}, fmt.Errorf("historical summary request builder validation failed: %w", err)
	}
	return b.req.Clone(), nil
}
