package reflection

import (
	"errors"
	"fmt"
	"math"

	"idun/intelligence/communication"
)

// SchemaVersion defines the canonical invariant schema version for Reflection structures.
const SchemaVersion = "2.0.0-FROZEN"

// TopicReflectionReports defines the Global Workspace topic on which Reflection publishes ReflectionReports.
const TopicReflectionReports communication.TopicID = communication.TopicReflections

// Cardinality and string length bounding limits to prevent report explosion over decades.
const (
	MaxSpecialistFindingsPerReport = 20
	MaxCrossCognitiveFindings      = 10
	MaxTrendFindings               = 10
	MaxSessionNotes                = 15
	MaxRecommendations             = 10
	MaxStringLength                = 1024
)

// EvaluationVerdict represents the explicit evaluation outcome of a specialist.
type EvaluationVerdict string

const (
	VerdictEvaluated        EvaluationVerdict = "EVALUATED"
	VerdictInsufficientData EvaluationVerdict = "INSUFFICIENT_DATA"
	VerdictAbstain          EvaluationVerdict = "ABSTAIN"
)

// ReflectionMode indicates whether the reflection episode evaluates a single episode or periodic trends.
type ReflectionMode string

const (
	ModeEpisode  ReflectionMode = "MODE_EPISODE"
	ModePeriodic ReflectionMode = "MODE_PERIODIC"
)

var (
	ErrInvalidSchemaVersion             = errors.New("reflection: invalid schema version")
	ErrMissingID                        = errors.New("reflection: missing required identifier")
	ErrInvalidConfidence                = errors.New("reflection: confidence out of bounds [0.0, 1.0]")
	ErrInvalidVerdict                   = errors.New("reflection: invalid evaluation verdict")
	ErrInvalidMode                      = errors.New("reflection: invalid reflection mode")
	ErrCardinalityExceeded              = errors.New("reflection: report cardinality limit exceeded")
	ErrStringLengthExceeded             = errors.New("reflection: string length limit exceeded")
	ErrInvalidTimeWindow                = errors.New("reflection: invalid time window")
	ErrInvalidTraceReference            = errors.New("reflection: invalid trace reference")
	ErrServiceClosed                    = errors.New("reflection: service closed")
	ErrPeriodicReflectionNotImplemented = errors.New("reflection: periodic reflection not implemented in Phase 2")
	ErrSpecialistAlreadyRegistered      = errors.New("reflection: specialist already registered")
	ErrNoSpecialistsRegistered          = errors.New("reflection: no specialists registered")
	ErrInvalidSpecialist                = errors.New("reflection: invalid specialist evaluator")
	ErrValidationFailed                 = errors.New("reflection: report validation failed")
)

func validateStringLength(fieldName string, value string) error {
	if len(value) > MaxStringLength {
		return fmt.Errorf("%w: field %s length %d exceeds max %d", ErrStringLengthExceeded, fieldName, len(value), MaxStringLength)
	}
	return nil
}

func validateConfidence(fieldName string, value float64) error {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0.0 || value > 1.0 {
		return fmt.Errorf("%w: field %s has value %f", ErrInvalidConfidence, fieldName, value)
	}
	return nil
}

// TraceReference references an immutable Global Workspace trace envelope that contributed to an evaluation.
type TraceReference struct {
	EnvelopeID     string `json:"envelope_id"`
	SourceAbility  string `json:"source_ability"`
	TraceTimestamp int64  `json:"trace_timestamp"`
	PayloadHashRef string `json:"payload_hash_ref"`
}

// Validate verifies that TraceReference is well-formed.
func (tr TraceReference) Validate() error {
	if tr.EnvelopeID == "" {
		return fmt.Errorf("%w: trace reference missing envelope ID", ErrInvalidTraceReference)
	}
	if err := validateStringLength("trace_envelope_id", tr.EnvelopeID); err != nil {
		return err
	}
	if err := validateStringLength("trace_source_ability", tr.SourceAbility); err != nil {
		return err
	}
	return validateStringLength("trace_payload_hash_ref", tr.PayloadHashRef)
}

// SpecialistReport contains the structured evaluation produced by an individual Reflection specialist.
type SpecialistReport struct {
	SpecialistID         string            `json:"specialist_id"`
	TargetAbility        string            `json:"target_ability"`
	Verdict              EvaluationVerdict `json:"verdict"`
	WentWell             []string          `json:"went_well"`
	WentPoorly           []string          `json:"went_poorly"`
	CouldImprove         []string          `json:"could_improve"`
	ReflectionConfidence float64           `json:"reflection_confidence"`
	SourceTraceRefs      []TraceReference  `json:"source_trace_refs"`
}

// Validate verifies schema invariants and cardinality bounds for a SpecialistReport.
func (sr SpecialistReport) Validate() error {
	if sr.SpecialistID == "" {
		return fmt.Errorf("%w: specialist report missing specialist ID", ErrMissingID)
	}
	if err := validateStringLength("specialist_id", sr.SpecialistID); err != nil {
		return err
	}
	if err := validateStringLength("target_ability", sr.TargetAbility); err != nil {
		return err
	}

	switch sr.Verdict {
	case VerdictEvaluated, VerdictInsufficientData, VerdictAbstain:
	default:
		return fmt.Errorf("%w: %s", ErrInvalidVerdict, sr.Verdict)
	}

	if err := validateConfidence("specialist_reflection_confidence", sr.ReflectionConfidence); err != nil {
		return err
	}

	totalFindings := len(sr.WentWell) + len(sr.WentPoorly) + len(sr.CouldImprove)
	if totalFindings > MaxSpecialistFindingsPerReport {
		return fmt.Errorf("%w: specialist report findings %d exceeds max %d", ErrCardinalityExceeded, totalFindings, MaxSpecialistFindingsPerReport)
	}

	for i, s := range sr.WentWell {
		if err := validateStringLength(fmt.Sprintf("went_well[%d]", i), s); err != nil {
			return err
		}
	}
	for i, s := range sr.WentPoorly {
		if err := validateStringLength(fmt.Sprintf("went_poorly[%d]", i), s); err != nil {
			return err
		}
	}
	for i, s := range sr.CouldImprove {
		if err := validateStringLength(fmt.Sprintf("could_improve[%d]", i), s); err != nil {
			return err
		}
	}
	for _, tr := range sr.SourceTraceRefs {
		if err := tr.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// CrossCognitiveFinding captures inter-ability interaction patterns and handoff issues.
type CrossCognitiveFinding struct {
	FindingID        string           `json:"finding_id"`
	SourceAbilities  []string         `json:"source_abilities"`
	InteractionIssue string           `json:"interaction_issue"`
	Recommendation   string           `json:"recommendation"`
	SeverityScore    float64          `json:"severity_score"`
	SourceTraceRefs  []TraceReference `json:"source_trace_refs"`
}

// Validate verifies CrossCognitiveFinding invariants.
func (cf CrossCognitiveFinding) Validate() error {
	if cf.FindingID == "" {
		return fmt.Errorf("%w: cross cognitive finding missing finding ID", ErrMissingID)
	}
	if err := validateStringLength("finding_id", cf.FindingID); err != nil {
		return err
	}
	if err := validateStringLength("interaction_issue", cf.InteractionIssue); err != nil {
		return err
	}
	if err := validateStringLength("recommendation", cf.Recommendation); err != nil {
		return err
	}
	if err := validateConfidence("severity_score", cf.SeverityScore); err != nil {
		return err
	}
	for _, tr := range cf.SourceTraceRefs {
		if err := tr.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// GrowthPotentialEstimate evaluates where future Learning investment would yield optimal ROI.
type GrowthPotentialEstimate struct {
	AbilityName           string  `json:"ability_name"`
	CurrentQualitySignal  float64 `json:"current_quality_signal"`
	GrowthPotentialRating float64 `json:"growth_potential_rating"`
	Rationale             string  `json:"rationale"`
}

// Validate checks GrowthPotentialEstimate bounds.
func (gp GrowthPotentialEstimate) Validate() error {
	if gp.AbilityName == "" {
		return fmt.Errorf("%w: growth potential estimate missing ability name", ErrMissingID)
	}
	if err := validateStringLength("ability_name", gp.AbilityName); err != nil {
		return err
	}
	if err := validateStringLength("rationale", gp.Rationale); err != nil {
		return err
	}
	if err := validateConfidence("current_quality_signal", gp.CurrentQualitySignal); err != nil {
		return err
	}
	return validateConfidence("growth_potential_rating", gp.GrowthPotentialRating)
}

// TrendFinding surfaces multi-episode longitudinal behavioral shifts identified by Periodic Reflection.
type TrendFinding struct {
	FindingID     string  `json:"finding_id"`
	TargetAbility string  `json:"target_ability"`
	PatternType   string  `json:"pattern_type"`
	Description   string  `json:"description"`
	TrendVelocity float64 `json:"trend_velocity"`
	Confidence    float64 `json:"confidence"`
}

// Validate verifies TrendFinding correctness.
func (tf TrendFinding) Validate() error {
	if tf.FindingID == "" {
		return fmt.Errorf("%w: trend finding missing finding ID", ErrMissingID)
	}
	if err := validateStringLength("finding_id", tf.FindingID); err != nil {
		return err
	}
	if err := validateStringLength("target_ability", tf.TargetAbility); err != nil {
		return err
	}
	if err := validateStringLength("description", tf.Description); err != nil {
		return err
	}
	if math.IsNaN(tf.TrendVelocity) || math.IsInf(tf.TrendVelocity, 0) {
		return errors.New("reflection: trend velocity must be finite")
	}
	return validateConfidence("trend_confidence", tf.Confidence)
}

// RecommendedLearningSignal proposes structured targets for permanent improvement by Learning.
type RecommendedLearningSignal struct {
	SignalID      string  `json:"signal_id"`
	TargetAbility string  `json:"target_ability"`
	SignalType    string  `json:"signal_type"`
	Description   string  `json:"description"`
	ExpectedROI   float64 `json:"expected_roi"`
}

// Validate verifies RecommendedLearningSignal correctness.
func (rls RecommendedLearningSignal) Validate() error {
	if rls.SignalID == "" {
		return fmt.Errorf("%w: recommended learning signal missing signal ID", ErrMissingID)
	}
	if err := validateStringLength("signal_id", rls.SignalID); err != nil {
		return err
	}
	if err := validateStringLength("target_ability", rls.TargetAbility); err != nil {
		return err
	}
	if err := validateStringLength("description", rls.Description); err != nil {
		return err
	}
	if math.IsNaN(rls.ExpectedROI) || math.IsInf(rls.ExpectedROI, 0) {
		return errors.New("reflection: expected ROI must be finite")
	}
	return nil
}

// SelfCalibrationReport measures how accurately earlier Reflection evaluations predicted real-world outcomes.
type SelfCalibrationReport struct {
	ReportID              string   `json:"report_id"`
	EvaluatedPeriodStart  int64    `json:"evaluated_period_start"`
	EvaluatedPeriodEnd    int64    `json:"evaluated_period_end"`
	PriorPredictionError  float64  `json:"prior_prediction_error"`
	ReflectionReliability float64  `json:"reflection_reliability"`
	CalibrationTrend      string   `json:"calibration_trend,omitempty"`
	BiasIndicators        []string `json:"bias_indicators,omitempty"`
	Summary               string   `json:"summary"`
}

// Validate checks SelfCalibrationReport bounds.
func (sc SelfCalibrationReport) Validate() error {
	if sc.ReportID == "" {
		return fmt.Errorf("%w: self calibration report missing report ID", ErrMissingID)
	}
	if err := validateStringLength("report_id", sc.ReportID); err != nil {
		return err
	}
	if err := validateStringLength("summary", sc.Summary); err != nil {
		return err
	}
	if err := validateStringLength("calibration_trend", sc.CalibrationTrend); err != nil {
		return err
	}
	if len(sc.BiasIndicators) > MaxSpecialistFindingsPerReport {
		return fmt.Errorf("%w: bias indicators %d exceeds max %d", ErrCardinalityExceeded, len(sc.BiasIndicators), MaxSpecialistFindingsPerReport)
	}
	for i, bias := range sc.BiasIndicators {
		if err := validateStringLength(fmt.Sprintf("bias_indicators[%d]", i), bias); err != nil {
			return err
		}
	}
	if math.IsNaN(sc.PriorPredictionError) || math.IsInf(sc.PriorPredictionError, 0) {
		return errors.New("reflection: prior prediction error must be finite")
	}
	return validateConfidence("reflection_reliability", sc.ReflectionReliability)
}

// ReportCategory classifies the cognitive target of a ReflectionReport.
type ReportCategory string

const (
	ReportCategoryCognitivePerformance ReportCategory = "COGNITIVE_PERFORMANCE"
	ReportCategoryLearningDiagnostics  ReportCategory = "LEARNING_DIAGNOSTICS"
)

// ReflectionReport is the canonical immutable output envelope published by Reflection to the Global Workspace.
type ReflectionReport struct {
	SchemaVersion              string                      `json:"schema_version"`
	ReportID                   string                      `json:"report_id"`
	Category                   ReportCategory              `json:"category,omitempty"`
	EpisodeID                  string                      `json:"episode_id,omitempty"`
	SummaryID                  string                      `json:"summary_id,omitempty"`
	TimeWindow                 *TimeWindowSpec             `json:"time_window,omitempty"`
	Timestamp                  int64                       `json:"timestamp"`
	Mode                       ReflectionMode              `json:"mode"`
	SpecialistReports          []SpecialistReport          `json:"specialist_reports"`
	CrossCognitiveFindings     []CrossCognitiveFinding     `json:"cross_cognitive_findings"`
	GrowthPotentialEstimates   []GrowthPotentialEstimate   `json:"growth_potential_estimates"`
	TrendFindings              []TrendFinding              `json:"trend_findings,omitempty"`
	RecommendedLearningSignals []RecommendedLearningSignal `json:"recommended_learning_signals"`
	SessionNotes               []string                    `json:"session_notes"`
	SelfCalibration            *SelfCalibrationReport      `json:"self_calibration,omitempty"`
}

// Validate strictly enforces schema versions, mode requirements, cardinality limits, string lengths, and nested validity.
func (r ReflectionReport) Validate() error {
	if r.SchemaVersion != SchemaVersion {
		return fmt.Errorf("%w: got %s, want %s", ErrInvalidSchemaVersion, r.SchemaVersion, SchemaVersion)
	}
	if r.ReportID == "" {
		return fmt.Errorf("%w: reflection report missing report ID", ErrMissingID)
	}
	if err := validateStringLength("report_id", r.ReportID); err != nil {
		return err
	}

	switch r.Mode {
	case ModeEpisode:
		if r.EpisodeID == "" {
			return fmt.Errorf("%w: episode mode report missing episode ID", ErrMissingID)
		}
	case ModePeriodic:
		if r.SummaryID == "" {
			return fmt.Errorf("%w: periodic mode report missing summary ID", ErrMissingID)
		}
	default:
		return fmt.Errorf("%w: %s", ErrInvalidMode, r.Mode)
	}

	if err := validateStringLength("episode_id", r.EpisodeID); err != nil {
		return err
	}
	if err := validateStringLength("summary_id", r.SummaryID); err != nil {
		return err
	}

	if len(r.CrossCognitiveFindings) > MaxCrossCognitiveFindings {
		return fmt.Errorf("%w: cross cognitive findings %d exceeds max %d", ErrCardinalityExceeded, len(r.CrossCognitiveFindings), MaxCrossCognitiveFindings)
	}
	if len(r.TrendFindings) > MaxTrendFindings {
		return fmt.Errorf("%w: trend findings %d exceeds max %d", ErrCardinalityExceeded, len(r.TrendFindings), MaxTrendFindings)
	}
	if len(r.RecommendedLearningSignals) > MaxRecommendations {
		return fmt.Errorf("%w: recommended learning signals %d exceeds max %d", ErrCardinalityExceeded, len(r.RecommendedLearningSignals), MaxRecommendations)
	}
	if len(r.SessionNotes) > MaxSessionNotes {
		return fmt.Errorf("%w: session notes %d exceeds max %d", ErrCardinalityExceeded, len(r.SessionNotes), MaxSessionNotes)
	}

	for i, note := range r.SessionNotes {
		if err := validateStringLength(fmt.Sprintf("session_notes[%d]", i), note); err != nil {
			return err
		}
	}
	for _, sr := range r.SpecialistReports {
		if err := sr.Validate(); err != nil {
			return err
		}
	}
	for _, cf := range r.CrossCognitiveFindings {
		if err := cf.Validate(); err != nil {
			return err
		}
	}
	for _, gp := range r.GrowthPotentialEstimates {
		if err := gp.Validate(); err != nil {
			return err
		}
	}
	for _, tf := range r.TrendFindings {
		if err := tf.Validate(); err != nil {
			return err
		}
	}
	for _, sig := range r.RecommendedLearningSignals {
		if err := sig.Validate(); err != nil {
			return err
		}
	}
	if r.SelfCalibration != nil {
		if err := r.SelfCalibration.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Clone creates a deep copy of a ReflectionReport to guarantee immutability.
func (r ReflectionReport) Clone() ReflectionReport {
	clone := r
	clone.SpecialistReports = make([]SpecialistReport, len(r.SpecialistReports))
	for i, sr := range r.SpecialistReports {
		clone.SpecialistReports[i] = sr
		clone.SpecialistReports[i].WentWell = append([]string(nil), sr.WentWell...)
		clone.SpecialistReports[i].WentPoorly = append([]string(nil), sr.WentPoorly...)
		clone.SpecialistReports[i].CouldImprove = append([]string(nil), sr.CouldImprove...)
		clone.SpecialistReports[i].SourceTraceRefs = append([]TraceReference(nil), sr.SourceTraceRefs...)
	}

	clone.CrossCognitiveFindings = make([]CrossCognitiveFinding, len(r.CrossCognitiveFindings))
	for i, cf := range r.CrossCognitiveFindings {
		clone.CrossCognitiveFindings[i] = cf
		clone.CrossCognitiveFindings[i].SourceAbilities = append([]string(nil), cf.SourceAbilities...)
		clone.CrossCognitiveFindings[i].SourceTraceRefs = append([]TraceReference(nil), cf.SourceTraceRefs...)
	}

	clone.GrowthPotentialEstimates = append([]GrowthPotentialEstimate(nil), r.GrowthPotentialEstimates...)
	clone.TrendFindings = append([]TrendFinding(nil), r.TrendFindings...)
	clone.RecommendedLearningSignals = append([]RecommendedLearningSignal(nil), r.RecommendedLearningSignals...)
	clone.SessionNotes = append([]string(nil), r.SessionNotes...)

	if r.SelfCalibration != nil {
		sc := *r.SelfCalibration
		sc.BiasIndicators = append([]string(nil), r.SelfCalibration.BiasIndicators...)
		clone.SelfCalibration = &sc
	}
	if r.TimeWindow != nil {
		tw := *r.TimeWindow
		clone.TimeWindow = &tw
	}
	return clone
}

// TimeWindowSpec specifies a bounded time window interval for longitudinal requests and summaries.
type TimeWindowSpec struct {
	StartTime int64 `json:"start_time"`
	EndTime   int64 `json:"end_time"`
}

// Validate checks that TimeWindowSpec boundaries are valid.
func (tw TimeWindowSpec) Validate() error {
	if tw.EndTime < tw.StartTime {
		return fmt.Errorf("%w: end time %d < start time %d", ErrInvalidTimeWindow, tw.EndTime, tw.StartTime)
	}
	return nil
}

// HistoricalSummaryRequest represents a versioned query emitted by Periodic Reflection to retrieve longitudinal summaries.
type HistoricalSummaryRequest struct {
	SchemaVersion    string         `json:"schema_version"`
	RequestID        string         `json:"request_id"`
	TimeWindow       TimeWindowSpec `json:"time_window"`
	TargetAbilities  []string       `json:"target_abilities"`
	AggregationLevel string         `json:"aggregation_level"`
	RequestedMetrics []string       `json:"requested_metrics"`
}

// Validate checks HistoricalSummaryRequest invariants.
func (hr HistoricalSummaryRequest) Validate() error {
	if hr.SchemaVersion != SchemaVersion {
		return fmt.Errorf("%w: got %s, want %s", ErrInvalidSchemaVersion, hr.SchemaVersion, SchemaVersion)
	}
	if hr.RequestID == "" {
		return fmt.Errorf("%w: historical summary request missing request ID", ErrMissingID)
	}
	if err := validateStringLength("request_id", hr.RequestID); err != nil {
		return err
	}
	if err := hr.TimeWindow.Validate(); err != nil {
		return err
	}
	for i, a := range hr.TargetAbilities {
		if err := validateStringLength(fmt.Sprintf("target_abilities[%d]", i), a); err != nil {
			return err
		}
	}
	return nil
}

// Clone creates a deep copy of HistoricalSummaryRequest.
func (hr HistoricalSummaryRequest) Clone() HistoricalSummaryRequest {
	clone := hr
	clone.TargetAbilities = append([]string(nil), hr.TargetAbilities...)
	clone.RequestedMetrics = append([]string(nil), hr.RequestedMetrics...)
	return clone
}

// HistoricalSummary is the versioned contract returned read-only by Memory to Trend Reflection.
type HistoricalSummary struct {
	SchemaVersion      string             `json:"schema_version"`
	SummaryID          string             `json:"summary_id"`
	GeneratedTimestamp int64              `json:"generated_timestamp"`
	TimeWindow         TimeWindowSpec     `json:"time_window"`
	EpisodeCount       int64              `json:"episode_count"`
	AverageScores      map[string]float64 `json:"average_scores"`
	TrendMetrics       map[string]float64 `json:"trend_metrics"`
	FailureRates       map[string]float64 `json:"failure_rates"`
	ImprovementRates   map[string]float64 `json:"improvement_rates"`
	SummaryConfidence  float64            `json:"summary_confidence"`
}

// Validate verifies HistoricalSummary invariants.
func (hs HistoricalSummary) Validate() error {
	if hs.SchemaVersion != SchemaVersion {
		return fmt.Errorf("%w: got %s, want %s", ErrInvalidSchemaVersion, hs.SchemaVersion, SchemaVersion)
	}
	if hs.SummaryID == "" {
		return fmt.Errorf("%w: historical summary missing summary ID", ErrMissingID)
	}
	if err := validateStringLength("summary_id", hs.SummaryID); err != nil {
		return err
	}
	if hs.EpisodeCount < 0 {
		return errors.New("reflection: episode count cannot be negative")
	}
	if err := hs.TimeWindow.Validate(); err != nil {
		return err
	}
	if err := validateConfidence("summary_confidence", hs.SummaryConfidence); err != nil {
		return err
	}
	for k, v := range hs.AverageScores {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return fmt.Errorf("reflection: average score for %s must be finite", k)
		}
	}
	for k, v := range hs.FailureRates {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return fmt.Errorf("reflection: failure rate for %s must be finite", k)
		}
	}
	return nil
}

// Clone creates a deep copy of HistoricalSummary.
func (hs HistoricalSummary) Clone() HistoricalSummary {
	clone := hs
	clone.AverageScores = make(map[string]float64, len(hs.AverageScores))
	for k, v := range hs.AverageScores {
		clone.AverageScores[k] = v
	}
	clone.TrendMetrics = make(map[string]float64, len(hs.TrendMetrics))
	for k, v := range hs.TrendMetrics {
		clone.TrendMetrics[k] = v
	}
	clone.FailureRates = make(map[string]float64, len(hs.FailureRates))
	for k, v := range hs.FailureRates {
		clone.FailureRates[k] = v
	}
	clone.ImprovementRates = make(map[string]float64, len(hs.ImprovementRates))
	for k, v := range hs.ImprovementRates {
		clone.ImprovementRates[k] = v
	}
	return clone
}
