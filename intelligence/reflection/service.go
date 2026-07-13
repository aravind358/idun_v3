package reflection

import (
	"context"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"idun/intelligence/communication"
	"idun/intelligence/executive"
	"idun/intelligence/workspace"
)

// Service implements ReflectionService and executive.AbilityDriver.
// It orchestrates read-only trace evaluation across registered specialists and publishes
// validated ReflectionReport envelopes to the Global Workspace.
type Service struct {
	mu        sync.RWMutex
	cfg       Config
	ws        workspace.Workspace
	registry  *SpecialistRegistry
	telemetry *telemetryCollector

	started atomic.Bool
	closed  atomic.Bool
}

// ServiceOption configures a Service instance at initialization time.
type ServiceOption func(*Service)

// WithWorkspace injects the Global Workspace instance for event subscription and report publication.
func WithWorkspace(ws workspace.Workspace) ServiceOption {
	return func(s *Service) {
		s.ws = ws
	}
}

// WithConfig sets custom operational configuration for the service.
func WithConfig(cfg Config) ServiceOption {
	return func(s *Service) {
		s.cfg = cfg
	}
}

// NewService constructs a new Reflection Service coordinator.
func NewService(opts ...ServiceOption) *Service {
	srv := &Service{
		cfg:       DefaultConfig(),
		registry:  NewSpecialistRegistry(),
		telemetry: &telemetryCollector{},
	}
	for _, opt := range opts {
		opt(srv)
	}
	return srv
}

// Telemetry returns an immutable snapshot of all content-blind operational metrics.
func (s *Service) Telemetry() TelemetrySnapshot {
	return s.telemetry.snapshot()
}

// Name returns the canonical ability driver name.
func (s *Service) Name() string {
	return "Reflection"
}

// Ability returns the executive CognitiveAbility enum value.
func (s *Service) Ability() executive.CognitiveAbility {
	return executive.AbilityReflection
}

// Start boots the Reflection Service coordinator.
func (s *Service) Start() error {
	if s.closed.Load() {
		return ErrServiceClosed
	}
	s.started.Store(true)
	return nil
}

// Close gracefully shuts down the Reflection Service coordinator.
func (s *Service) Close() error {
	if s.closed.Swap(true) {
		return ErrServiceClosed
	}
	s.started.Store(false)
	return nil
}

// RegisterSpecialist registers a specialist evaluator with the service's registry.
func (s *Service) RegisterSpecialist(eval SpecialistEvaluator) error {
	if s.closed.Load() {
		return ErrServiceClosed
	}
	return s.registry.Register(eval)
}

// GetSpecialists returns a snapshot of all registered specialist evaluators.
func (s *Service) GetSpecialists() []SpecialistEvaluator {
	return s.registry.All()
}

// ReflectEpisode evaluates a single completed cognitive episode from read-only Workspace traces,
// validates the generated ReflectionReport against strict invariants, and publishes it to the Workspace.
func (s *Service) ReflectEpisode(ctx context.Context, episodeID string, traces []communication.Envelope) (ReflectionReport, error) {
	if s.closed.Load() {
		return ReflectionReport{}, ErrServiceClosed
	}
	if err := ctx.Err(); err != nil {
		s.telemetry.cancellationCount.Add(1)
		return ReflectionReport{}, err
	}
	start := time.Now()

	specialistReports, err := s.registry.EvaluateAll(ctx, traces)
	if err != nil {
		return ReflectionReport{}, fmt.Errorf("reflection episode evaluation failed: %w", err)
	}

	now := time.Now().UnixNano()
	reportID := fmt.Sprintf("refl-%s-%d", episodeID, now)
	builder := NewReflectionReportBuilder(reportID, ModeEpisode, now).
		WithEpisodeID(episodeID)

	for _, sr := range specialistReports {
		builder.AddSpecialistReport(sr)
		if sr.Verdict == VerdictAbstain {
			s.telemetry.abstainCount.Add(1)
		}
	}

	// Phase 3: Integrate Cross-Cognitive Analysis and Growth Potential estimation
	ccFindings := AnalyzeCrossCognitiveEpisode(specialistReports)
	for _, cf := range ccFindings {
		builder.AddCrossCognitiveFinding(cf)
	}
	gpEstimates, rlsSignals := EstimateGrowthPotentialEpisode(specialistReports, episodeID)
	for _, gp := range gpEstimates {
		builder.AddGrowthPotentialEstimate(gp)
	}
	for _, sig := range rlsSignals {
		builder.AddRecommendedLearningSignal(sig)
	}

	report, err := builder.Build()
	if err != nil {
		s.telemetry.validationFailures.Add(1)
		return ReflectionReport{}, fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	// Publish validated report to Global Workspace if configured
	if s.ws != nil {
		env := communication.Envelope{
			ID:              fmt.Sprintf("env-%s", report.ReportID),
			Source:          s.Name(),
			Topic:           TopicReflectionReports,
			ParentRef:       episodeID,
			PayloadRef:      report.ReportID,
			PayloadModality: "reflection-report",
			RawConfidence:   1.0,
			CreatedAt:       time.Now(),
		}
		if pubErr := s.ws.Publish(ctx, env); pubErr != nil {
			return report, fmt.Errorf("reflection report publication failed: %w", pubErr)
		}
	}

	s.telemetry.totalEpisodeReflections.Add(1)
	s.telemetry.specialistEvaluations.Add(uint64(len(specialistReports)))
	s.telemetry.crossCognitiveAnalyses.Add(uint64(len(ccFindings)))
	s.telemetry.growthPotentialEstimations.Add(uint64(len(gpEstimates)))
	s.telemetry.totalEpisodeDurationNs.Add(time.Since(start).Nanoseconds())

	return report, nil
}

// ReflectPeriodic evaluates multi-episode longitudinal trends from read-only HistoricalSummary contracts,
// integrates trend analysis, cross-cognitive analysis, and growth potential recommendations,
// validates the generated ReflectionReport, and publishes it to the Workspace.
func (s *Service) ReflectPeriodic(ctx context.Context, summary HistoricalSummary) (ReflectionReport, error) {
	if s.closed.Load() {
		return ReflectionReport{}, ErrServiceClosed
	}
	if err := ctx.Err(); err != nil {
		s.telemetry.cancellationCount.Add(1)
		return ReflectionReport{}, err
	}
	if err := summary.Validate(); err != nil {
		s.telemetry.validationFailures.Add(1)
		return ReflectionReport{}, fmt.Errorf("%w: invalid historical summary: %v", ErrValidationFailed, err)
	}
	start := time.Now()

	now := time.Now().UnixNano()
	reportID := fmt.Sprintf("refl-periodic-%s-%d", summary.SummaryID, now)
	builder := NewReflectionReportBuilder(reportID, ModePeriodic, now).
		WithSummaryID(summary.SummaryID).
		WithTimeWindow(summary.TimeWindow.StartTime, summary.TimeWindow.EndTime)

	// Perform Phase 3 pure evaluations
	trends, err := AnalyzeTrends(summary)
	if err != nil {
		return ReflectionReport{}, fmt.Errorf("trend analysis failed: %w", err)
	}
	for _, tf := range trends {
		builder.AddTrendFinding(tf)
	}

	ccFindings := AnalyzeCrossCognitivePeriodic(summary)
	for _, cf := range ccFindings {
		builder.AddCrossCognitiveFinding(cf)
	}

	gpEstimates, rlsSignals := EstimateGrowthPotentialPeriodic(summary)
	for _, gp := range gpEstimates {
		builder.AddGrowthPotentialEstimate(gp)
	}
	for _, sig := range rlsSignals {
		builder.AddRecommendedLearningSignal(sig)
	}

	report, err := builder.Build()
	if err != nil {
		s.telemetry.validationFailures.Add(1)
		return ReflectionReport{}, fmt.Errorf("%w: periodic reflection validation failed: %v", ErrValidationFailed, err)
	}

	// Publish validated periodic report to Global Workspace if configured
	if s.ws != nil {
		env := communication.Envelope{
			ID:              fmt.Sprintf("env-%s", report.ReportID),
			Source:          s.Name(),
			Topic:           TopicReflectionReports,
			ParentRef:       summary.SummaryID,
			PayloadRef:      report.ReportID,
			PayloadModality: "reflection-report",
			RawConfidence:   1.0,
			CreatedAt:       time.Now(),
		}
		if pubErr := s.ws.Publish(ctx, env); pubErr != nil {
			return report, fmt.Errorf("periodic reflection report publication failed: %w", pubErr)
		}
	}

	s.telemetry.totalPeriodicReflections.Add(1)
	s.telemetry.trendAnalyses.Add(uint64(len(trends)))
	s.telemetry.crossCognitiveAnalyses.Add(uint64(len(ccFindings)))
	s.telemetry.growthPotentialEstimations.Add(uint64(len(gpEstimates)))
	s.telemetry.totalPeriodicDurationNs.Add(time.Since(start).Nanoseconds())

	return report, nil
}

// ReflectOnReflection executes the Phase 4 metacognitive Reflection-on-Reflection pipeline.
// It evaluates the quality and accuracy of previous ReflectionReports against actual outcomes
// in currentSummary, produces a diagnostic SelfCalibrationReport, and publishes a validated report.
// It never modifies Reflection itself or internal EvaluationStrategy weights.
func (s *Service) ReflectOnReflection(ctx context.Context, priorReports []ReflectionReport, currentSummary HistoricalSummary) (ReflectionReport, error) {
	if s.closed.Load() {
		return ReflectionReport{}, ErrServiceClosed
	}
	if err := ctx.Err(); err != nil {
		s.telemetry.cancellationCount.Add(1)
		return ReflectionReport{}, err
	}
	start := time.Now()

	selfCal, err := CompareHistoricalOutcomes(priorReports, currentSummary)
	if err != nil {
		s.telemetry.validationFailures.Add(1)
		return ReflectionReport{}, fmt.Errorf("reflection-on-reflection evaluation failed: %w", err)
	}

	now := time.Now().UnixNano()
	reportID := fmt.Sprintf("refl-meta-%s-%d", currentSummary.SummaryID, now)
	builder := NewReflectionReportBuilder(reportID, ModePeriodic, now).
		WithSummaryID(currentSummary.SummaryID).
		WithTimeWindow(currentSummary.TimeWindow.StartTime, currentSummary.TimeWindow.EndTime).
		WithSelfCalibration(selfCal).
		AddSessionNote(fmt.Sprintf("Metacognitive audit completed: reliability=%.2f trend=%s", selfCal.ReflectionReliability, selfCal.CalibrationTrend))

	if len(selfCal.BiasIndicators) > 0 {
		builder.AddSessionNote(fmt.Sprintf("Systematic bias indicators surfaced: %v", selfCal.BiasIndicators))
	}

	if selfCal.PriorPredictionError > 0.15 {
		builder.AddRecommendedLearningSignal(RecommendedLearningSignal{
			SignalID:      fmt.Sprintf("sig-cal-%s", currentSummary.SummaryID),
			TargetAbility: "Reflection",
			SignalType:    "SELF_CALIBRATION_REFINEMENT",
			Description:   fmt.Sprintf("Prior prediction error %.2f exceeds threshold; recommend calibration refinement", selfCal.PriorPredictionError),
			ExpectedROI:   math.Min(1.0, selfCal.PriorPredictionError+0.2),
		})
	}

	report, err := builder.Build()
	if err != nil {
		s.telemetry.validationFailures.Add(1)
		return ReflectionReport{}, fmt.Errorf("%w: reflection-on-reflection validation failed: %v", ErrValidationFailed, err)
	}

	if s.ws != nil {
		env := communication.Envelope{
			ID:              fmt.Sprintf("env-%s", report.ReportID),
			Source:          s.Name(),
			Topic:           TopicReflectionReports,
			ParentRef:       currentSummary.SummaryID,
			PayloadRef:      report.ReportID,
			PayloadModality: "reflection-on-reflection-report",
			RawConfidence:   1.0,
			CreatedAt:       time.Now(),
		}
		if pubErr := s.ws.Publish(ctx, env); pubErr != nil {
			return report, fmt.Errorf("reflection-on-reflection report publication failed: %w", pubErr)
		}
	}

	s.telemetry.totalMetaReflections.Add(1)
	s.telemetry.selfCalibrationRuns.Add(1)
	s.telemetry.totalMetaDurationNs.Add(time.Since(start).Nanoseconds())

	return report, nil
}
