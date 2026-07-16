// Package learning implements IDUN's Layer 1 Core Cognitive Engine Learning Ability.
//
// Architecture Version: 2.0.0-FROZEN
//
// Phase 1 establishes the immutable domain types, public interfaces, configuration,
// and validation firewall required by the Learning specification.
// Phase 1 strictly forbids implementing learning algorithms, statistical optimization,
// rollout execution, or policy adaptation.
package learning

import (
	"errors"
	"fmt"
	"math"
	"time"

	"idun/core/memory"
	"idun/intelligence/communication"
)

// SchemaVersion defines the canonical invariant schema version for Learning structures.
const SchemaVersion = "2.0.0-FROZEN"

// TopicLearningTraces defines the Global Workspace topic on which Learning publishes diagnostic traces.
const TopicLearningTraces communication.TopicID = "learning-traces"

// TopicCandidateSnapshots defines the Global Workspace topic on which validated candidate snapshots are broadcast.
const TopicCandidateSnapshots communication.TopicID = "candidate-snapshots"

// Cardinality and string length bounding limits to ensure bounded memory O(1) across decades.
const (
	MaxLearnerUsagesPerTrace       = 50
	MaxCandidatesPerResult         = 20
	MaxValidationRecordsPerSnapshot = 30
	MaxDomainWeights               = 100
	MaxStringLength                = 1024
	MaxPayloadBytes                = 10 * 1024 * 1024 // 10MB limit on snapshot payload
)

// LearningResultStatus records what state changes occurred during the learning cycle.
type LearningResultStatus string

const (
	StatusPublished      LearningResultStatus = "PUBLISHED"
	StatusValidationFail LearningResultStatus = "VALIDATION_FAILED"
	StatusAbstained      LearningResultStatus = "ABSTAINED"
	StatusNoChange       LearningResultStatus = "NO_CHANGE"
	StatusPartial        LearningResultStatus = "PARTIAL"
	StatusNoCandidates   LearningResultStatus = "NO_CANDIDATES"
	StatusRolledBack     LearningResultStatus = "ROLLED_BACK"
)

// LearningTerminationReason records exactly why computational execution halted.
type LearningTerminationReason string

const (
	ReasonSuccess               LearningTerminationReason = "SUCCESS"
	ReasonNoCandidates          LearningTerminationReason = "NO_CANDIDATES"
	ReasonSampleFloorNotMet     LearningTerminationReason = "SAMPLE_FLOOR_NOT_MET"
	ReasonCooldownActive        LearningTerminationReason = "COOLDOWN_ACTIVE"
	ReasonDriftAnomalyDetected  LearningTerminationReason = "DRIFT_ANOMALY_DETECTED"
	ReasonCapabilityUnavailable LearningTerminationReason = "CAPABILITY_UNAVAILABLE"
	ReasonTimeout               LearningTerminationReason = "TIMEOUT"
)

// CandidateLifecycle represents the immutable ownership state machine of a candidate snapshot.
type CandidateLifecycle string

const (
	// LifecycleDraft is owned by Learning (`idun/intelligence/learning`).
	LifecycleDraft CandidateLifecycle = "Draft"
	// LifecycleValidated is owned by Learning (`idun/intelligence/learning`).
	LifecycleValidated CandidateLifecycle = "Validated"
	// LifecycleShadow is owned by Rollout Executor (Infrastructure).
	LifecycleShadow CandidateLifecycle = "Shadow"
	// LifecycleCanary is owned by Rollout Executor (Infrastructure).
	LifecycleCanary CandidateLifecycle = "Canary"
	// LifecycleActive is owned by Rollout Executor (Infrastructure).
	LifecycleActive CandidateLifecycle = "Active"
	// LifecycleRetired is owned by Rollout Executor when superseded.
	LifecycleRetired CandidateLifecycle = "Retired"
	// LifecycleRolledBack is owned by Rollout Executor when regressed.
	LifecycleRolledBack CandidateLifecycle = "RolledBack"
)

// CampaignStatus represents the organizational tracking status of a multi-cycle LearningCampaign.
type CampaignStatus string

const (
	CampaignStatusScheduled CampaignStatus = "SCHEDULED"
	CampaignStatusActive    CampaignStatus = "ACTIVE"
	CampaignStatusCompleted CampaignStatus = "COMPLETED"
	CampaignStatusPaused    CampaignStatus = "PAUSED"
	CampaignStatusAborted   CampaignStatus = "ABORTED"
)

// Typed string identifiers and fingerprints.
type LearningFingerprint string
type PolicyFingerprint string
type SourceArtifactHash string

var (
	ErrInvalidSchemaVersion       = errors.New("learning: invalid schema version")
	ErrMissingID                  = errors.New("learning: missing required identifier")
	ErrInvalidConfidence          = errors.New("learning: value out of bounds [0.0, 1.0]")
	ErrInvalidStatus              = errors.New("learning: invalid result status")
	ErrInvalidTerminationReason   = errors.New("learning: invalid termination reason")
	ErrInvalidLifecycle           = errors.New("learning: invalid candidate lifecycle")
	ErrInvalidCampaignStatus      = errors.New("learning: invalid campaign status")
	ErrInvalidPolicyOwner         = errors.New("learning: LearningPolicyProfile author must be Executive")
	ErrStringLengthExceeded       = errors.New("learning: string length limit exceeded")
	ErrPayloadTooLarge            = errors.New("learning: payload exceeds maximum byte limit")
	ErrInvalidTimeWindow          = errors.New("learning: invalid time window")
	ErrCardinalityExceeded        = errors.New("learning: cardinality limit exceeded")
	ErrCapabilityUnavailable      = errors.New("learning: requested capability not supported by deployment")
	ErrServiceClosed              = errors.New("learning: service closed")
	ErrValidationFailed           = errors.New("learning: validation failed")
	ErrNotFound                   = errors.New("learning: resource not found")
)

func validateStringLength(fieldName string, value string) error {
	if len(value) > MaxStringLength {
		return fmt.Errorf("%w: field %s length %d exceeds max %d", ErrStringLengthExceeded, fieldName, len(value), MaxStringLength)
	}
	return nil
}

func validateRatio(fieldName string, value float64) error {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0.0 || value > 1.0 {
		return fmt.Errorf("%w: field %s has invalid ratio %f", ErrInvalidConfidence, fieldName, value)
	}
	return nil
}

// CandidateLineage records long-term evolutionary provenance across generations.
type CandidateLineage struct {
	ParentSnapshot   string `json:"parent_snapshot"`
	AncestorSnapshot string `json:"ancestor_snapshot"`
	GenerationDepth  uint32 `json:"generation_depth"`
}

// Validate verifies string length limits on CandidateLineage fields.
func (cl *CandidateLineage) Validate() error {
	if err := validateStringLength("parent_snapshot", cl.ParentSnapshot); err != nil {
		return err
	}
	if err := validateStringLength("ancestor_snapshot", cl.AncestorSnapshot); err != nil {
		return err
	}
	return nil
}

// ReplayMetadata records the exact cryptographic identity needed for bit-identical reproduction across decades.
type ReplayMetadata struct {
	LearningFingerprint LearningFingerprint `json:"learning_fingerprint"`
	PolicyFingerprint   PolicyFingerprint   `json:"policy_fingerprint"`
	LearnerFingerprint  string              `json:"learner_fingerprint,omitempty"`
	SourceArtifactHash  SourceArtifactHash  `json:"source_artifact_hash"`
	ReplaySeed          uint64              `json:"replay_seed"`
	ExperimentID        string              `json:"experiment_id,omitempty"`
	ParentSnapshot      string              `json:"parent_snapshot,omitempty"`
	AncestorSnapshot    string              `json:"ancestor_snapshot,omitempty"`
	GenerationDepth     uint32              `json:"generation_depth"`
}

// Validate verifies the structural integrity of ReplayMetadata.
func (rm *ReplayMetadata) Validate() error {
	if rm.LearningFingerprint == "" {
		return fmt.Errorf("%w: missing learning_fingerprint", ErrMissingID)
	}
	if rm.PolicyFingerprint == "" {
		return fmt.Errorf("%w: missing policy_fingerprint", ErrMissingID)
	}
	if rm.SourceArtifactHash == "" {
		return fmt.Errorf("%w: missing source_artifact_hash", ErrMissingID)
	}
	if err := validateStringLength("experiment_id", rm.ExperimentID); err != nil {
		return err
	}
	if err := validateStringLength("parent_snapshot", rm.ParentSnapshot); err != nil {
		return err
	}
	if err := validateStringLength("ancestor_snapshot", rm.AncestorSnapshot); err != nil {
		return err
	}
	if err := validateStringLength("learner_fingerprint", rm.LearnerFingerprint); err != nil {
		return err
	}
	return nil
}

// LearningCapabilities declares the structural and algorithmic features supported by this deployment.
type LearningCapabilities struct {
	SupportsCalibrationLearning   bool   `json:"supports_calibration_learning"`
	SupportsPreferenceLearning    bool   `json:"supports_preference_learning"`
	SupportsStatisticalLearning   bool   `json:"supports_statistical_learning"`
	SupportsReinforcementLearning bool   `json:"supports_reinforcement_learning"`
	SupportsOnlineLearning        bool   `json:"supports_online_learning"`
	SupportsOfflineLearning       bool   `json:"supports_offline_learning"`
	SupportsExperimentation       bool   `json:"supports_experimentation"`
	SupportsRollback              bool   `json:"supports_rollback"`
	SupportsShadowDeployment      bool   `json:"supports_shadow_deployment"`
	CapabilityFingerprint         string `json:"capability_fingerprint"`
}

// Validate verifies the capabilities declaration.
func (lc *LearningCapabilities) Validate() error {
	if lc.CapabilityFingerprint == "" {
		return fmt.Errorf("%w: missing capability_fingerprint", ErrMissingID)
	}
	return validateStringLength("capability_fingerprint", lc.CapabilityFingerprint)
}

// LearningPolicyProfile is Executive-owned and read-only to Learning.
// It dictates the rate limits, thresholds, and governance weights for learning candidates.
type LearningPolicyProfile struct {
	ProfileID                string             `json:"profile_id"`
	PolicyVersion            string             `json:"policy_version"`
	PolicyFingerprint        PolicyFingerprint  `json:"policy_fingerprint"`
	Author                   string             `json:"author"` // Must strictly equal "Executive"
	LearningRate             float64            `json:"learning_rate"`
	MinimumSampleSize        int                `json:"minimum_sample_size"`
	CooldownPeriod           time.Duration      `json:"cooldown_period"`
	ValidationThresholds     map[string]float64 `json:"validation_thresholds"`
	DriftDetectionThresholds map[string]float64 `json:"drift_detection_thresholds"`
	DomainWeights            map[string]float64 `json:"domain_weights"`
	LearnerWeights           map[string]float64 `json:"learner_weights"`
	ExperimentLimits         map[string]int     `json:"experiment_limits"`
}

// Validate verifies that the governing policy profile is valid and authored exclusively by Executive.
func (p *LearningPolicyProfile) Validate() error {
	if p.ProfileID == "" || p.PolicyVersion == "" || p.PolicyFingerprint == "" {
		return fmt.Errorf("%w: missing profile identification or fingerprint", ErrMissingID)
	}
	if p.Author != "Executive" {
		return fmt.Errorf("%w: author is %q, must be 'Executive'", ErrInvalidPolicyOwner, p.Author)
	}
	if err := validateRatio("learning_rate", p.LearningRate); err != nil {
		return err
	}
	if p.MinimumSampleSize < 1 {
		return fmt.Errorf("%w: minimum_sample_size must be >= 1", ErrValidationFailed)
	}
	if p.CooldownPeriod < 0 {
		return fmt.Errorf("%w: cooldown_period cannot be negative", ErrValidationFailed)
	}
	if len(p.DomainWeights) > MaxDomainWeights || len(p.LearnerWeights) > MaxDomainWeights {
		return fmt.Errorf("%w: domain/learner weights cardinality limit exceeded", ErrCardinalityExceeded)
	}
	for k, v := range p.ValidationThresholds {
		if err := validateRatio(fmt.Sprintf("validation_threshold[%s]", k), v); err != nil {
			return err
		}
	}
	for k, v := range p.DomainWeights {
		if err := validateRatio(fmt.Sprintf("domain_weight[%s]", k), v); err != nil {
			return err
		}
	}
	return nil
}

// SamplingMethod defines the strategy used by AggregationPolicy for sample selection.
type SamplingMethod string

const (
	SamplingMethodDeterministic SamplingMethod = "DETERMINISTIC"
	SamplingMethodUniform       SamplingMethod = "UNIFORM"
	SamplingMethodStratified    SamplingMethod = "STRATIFIED"
)

// OrderingStrategy defines how experiences are ordered within an aggregated summary.
type OrderingStrategy string

const (
	OrderingStrategyChronologicalAsc  OrderingStrategy = "CHRONOLOGICAL_ASC"
	OrderingStrategyChronologicalDesc OrderingStrategy = "CHRONOLOGICAL_DESC"
	OrderingStrategyDomainPriority    OrderingStrategy = "DOMAIN_PRIORITY"
)

// MergePolicy defines how duplicate experience records across windows are merged or retained.
type MergePolicy string

const (
	MergePolicyAppend           MergePolicy = "APPEND"
	MergePolicyDeduplicateHash  MergePolicy = "DEDUPLICATE_HASH"
	MergePolicyLatestWins       MergePolicy = "LATEST_WINS"
)

// AggregationPolicy defines the immutable rules governing how experiences are selected,
// filtered, and windowed. It is owned purely by Executive and consumed passively by Learning.
type AggregationPolicy struct {
	PolicyID           string             `json:"policy_id"`
	PolicyVersion      string             `json:"policy_version"`
	PolicyFingerprint  string             `json:"policy_fingerprint"`
	Strategy           string             `json:"strategy"` // e.g., "COGNITIVE_PERFORMANCE_DEFAULT"
	WindowDuration     time.Duration      `json:"window_duration"`
	SamplingMethod     SamplingMethod     `json:"sampling_method"`
	DomainPriorities   map[string]float64 `json:"domain_priorities"`
	MaximumArtifacts   uint32             `json:"maximum_artifacts"`
	MaximumMemoryBytes uint64             `json:"maximum_memory_bytes"`
	OrderingStrategy   OrderingStrategy   `json:"ordering_strategy"`
	MergePolicy        MergePolicy        `json:"merge_policy"`
}

// Validate checks the structural integrity and bounds of AggregationPolicy.
func (ap *AggregationPolicy) Validate() error {
	if ap.PolicyID == "" || ap.PolicyVersion == "" || ap.PolicyFingerprint == "" || ap.Strategy == "" {
		return fmt.Errorf("%w: missing required policy identification fields", ErrMissingID)
	}
	if ap.WindowDuration <= 0 {
		return fmt.Errorf("%w: window_duration must be positive", ErrInvalidTimeWindow)
	}
	switch ap.SamplingMethod {
	case SamplingMethodDeterministic, SamplingMethodUniform, SamplingMethodStratified:
	default:
		return fmt.Errorf("%w: invalid sampling method %q", ErrValidationFailed, ap.SamplingMethod)
	}
	switch ap.OrderingStrategy {
	case OrderingStrategyChronologicalAsc, OrderingStrategyChronologicalDesc, OrderingStrategyDomainPriority:
	default:
		return fmt.Errorf("%w: invalid ordering strategy %q", ErrValidationFailed, ap.OrderingStrategy)
	}
	switch ap.MergePolicy {
	case MergePolicyAppend, MergePolicyDeduplicateHash, MergePolicyLatestWins:
	default:
		return fmt.Errorf("%w: invalid merge policy %q", ErrValidationFailed, ap.MergePolicy)
	}
	if ap.MaximumArtifacts == 0 {
		return fmt.Errorf("%w: maximum_artifacts must be > 0", ErrValidationFailed)
	}
	if ap.MaximumMemoryBytes == 0 {
		return fmt.Errorf("%w: maximum_memory_bytes must be > 0", ErrValidationFailed)
	}
	if len(ap.DomainPriorities) > MaxDomainWeights {
		return fmt.Errorf("%w: domain_priorities cardinality exceeded", ErrCardinalityExceeded)
	}
	for domain, w := range ap.DomainPriorities {
		if err := validateRatio(fmt.Sprintf("domain_priorities[%s]", domain), w); err != nil {
			return err
		}
	}
	return nil
}

// LearnerUsage records telemetry counters and mechanically derived scores for a specific learner invocation.
type LearnerUsage struct {
	LearnerID          string        `json:"learner_id"`
	DomainSchemaID     string        `json:"domain_schema_id"`
	Invoked            bool          `json:"invoked"`
	Skipped            bool          `json:"skipped"`
	SkipReason         string        `json:"skip_reason,omitempty"`
	CandidatesProduced int           `json:"candidates_produced"`
	CandidatesAccepted int           `json:"candidates_accepted"`
	ExecutionTime      time.Duration `json:"execution_time"`
	ContributionScore  float64       `json:"contribution_score"` // Mechanically derived ratio: accepted / produced
}

// Validate verifies LearnerUsage consistency and mechanical derivation bounds.
func (lu *LearnerUsage) Validate() error {
	if lu.LearnerID == "" || lu.DomainSchemaID == "" {
		return fmt.Errorf("%w: missing learner_id or domain_schema_id", ErrMissingID)
	}
	if lu.CandidatesProduced < 0 || lu.CandidatesAccepted < 0 || lu.CandidatesAccepted > lu.CandidatesProduced {
		return fmt.Errorf("%w: invalid produced (%d) vs accepted (%d) counts", ErrValidationFailed, lu.CandidatesProduced, lu.CandidatesAccepted)
	}
	if lu.ExecutionTime < 0 {
		return fmt.Errorf("%w: negative execution time", ErrValidationFailed)
	}
	if err := validateRatio("contribution_score", lu.ContributionScore); err != nil {
		return err
	}
	return validateStringLength("skip_reason", lu.SkipReason)
}

// ExperimentProfile describes a bounded shadow or A/B experiment evaluation for a candidate snapshot.
type ExperimentProfile struct {
	ExperimentID     string        `json:"experiment_id"`
	DomainSchemaID   string        `json:"domain_schema_id"`
	TargetSnapshotID string        `json:"target_snapshot_id"`
	ShadowRatio      float64       `json:"shadow_ratio"`
	CanaryRatio      float64       `json:"canary_ratio"`
	MaxDuration      time.Duration `json:"max_duration"`
	ReplaySeed       uint64        `json:"replay_seed"`
	Priority         int           `json:"priority,omitempty"`
}

// Validate verifies ExperimentProfile parameters.
func (ep *ExperimentProfile) Validate() error {
	if ep.ExperimentID == "" || ep.DomainSchemaID == "" || ep.TargetSnapshotID == "" {
		return fmt.Errorf("%w: missing experiment identification fields", ErrMissingID)
	}
	if err := validateRatio("shadow_ratio", ep.ShadowRatio); err != nil {
		return err
	}
	if err := validateRatio("canary_ratio", ep.CanaryRatio); err != nil {
		return err
	}
	if ep.MaxDuration <= 0 {
		return fmt.Errorf("%w: max_duration must be positive", ErrValidationFailed)
	}
	if ep.Priority < 0 {
		return fmt.Errorf("%w: priority cannot be negative", ErrValidationFailed)
	}
	return nil
}

// ValidationEvidence records factual, purely observational counters and proof flags
// explaining why a candidate passed or failed without subjective interpretation.
type ValidationEvidence struct {
	SampleCount          uint64  `json:"sample_count"`
	DriftScore           float64 `json:"drift_score"`
	ReplayVerified       bool    `json:"replay_verified"`
	StructuralChecks     uint32  `json:"structural_checks"`
	ConstitutionalChecks uint32  `json:"constitutional_checks"`
	StatisticalChecks    uint32  `json:"statistical_checks"`
	ConstraintChecks     uint32  `json:"constraint_checks"`
	ValidationDurationUs uint64  `json:"validation_duration_us"`
}

// Validate verifies ValidationEvidence structural invariants.
func (ve *ValidationEvidence) Validate() error {
	if math.IsNaN(ve.DriftScore) || math.IsInf(ve.DriftScore, 0) || ve.DriftScore < 0.0 {
		return fmt.Errorf("%w: invalid drift_score %f", ErrValidationFailed, ve.DriftScore)
	}
	return nil
}

// ValidationResult records an individual statistical, minimum-sample, or constitutional check outcome.
type ValidationResult struct {
	Passed    bool                `json:"passed"`
	CheckID   string              `json:"check_id"`
	Score     float64             `json:"score"`
	Threshold float64             `json:"threshold"`
	Reason    string              `json:"reason"`
	Evidence  *ValidationEvidence `json:"evidence,omitempty"`
}

// Validate verifies ValidationResult metrics.
func (vr *ValidationResult) Validate() error {
	if vr.CheckID == "" {
		return fmt.Errorf("%w: missing check_id", ErrMissingID)
	}
	if err := validateStringLength("reason", vr.Reason); err != nil {
		return err
	}
	if vr.Evidence != nil {
		if err := vr.Evidence.Validate(); err != nil {
			return fmt.Errorf("evidence validation failed: %w", err)
		}
	}
	return nil
}

// StructuralValidationResult records rigorous static, complexity, memory, and cycle checks on a new strategy proposal.
type StructuralValidationResult struct {
	Passed             bool   `json:"passed"`
	StaticSyntaxPassed bool   `json:"static_syntax_passed"`
	ComplexityBounded  bool   `json:"complexity_bounded"`
	MemoryBounded      bool   `json:"memory_bounded"`
	CycleFree          bool   `json:"cycle_free"`
	APICompliant       bool   `json:"api_compliant"`
	MaxExecutionTimeMs int    `json:"max_execution_time_ms"`
	MaxMemoryBytes     int    `json:"max_memory_bytes"`
	Reason             string `json:"reason"`
}

// Validate verifies StructuralValidationResult invariants.
func (svr *StructuralValidationResult) Validate() error {
	if svr.MaxExecutionTimeMs < 0 || svr.MaxMemoryBytes < 0 {
		return fmt.Errorf("%w: negative complexity/memory bounds", ErrValidationFailed)
	}
	return validateStringLength("reason", svr.Reason)
}

// AggregationSummary describes the bounded historical corpus of artifacts aggregated during a learning window.
type AggregationSummary struct {
	SummaryID              string             `json:"summary_id"`
	TimeWindowStart        time.Time          `json:"time_window_start"`
	TimeWindowEnd          time.Time          `json:"time_window_end"`
	TotalArtifactsIngested int                `json:"total_artifacts_ingested"`
	SourceArtifactHash     SourceArtifactHash `json:"source_artifact_hash"`
	DomainSchemaIDs        []string           `json:"domain_schema_ids"`
	AggregationPolicyID    string             `json:"aggregation_policy_id,omitempty"`
	Records                []memory.Record    `json:"records,omitempty"`
}

// Validate verifies AggregationSummary structure and time bounds.
func (as *AggregationSummary) Validate() error {
	if as.SummaryID == "" || as.SourceArtifactHash == "" {
		return fmt.Errorf("%w: missing summary_id or source_artifact_hash", ErrMissingID)
	}
	if as.TimeWindowEnd.Before(as.TimeWindowStart) {
		return fmt.Errorf("%w: end time before start time", ErrInvalidTimeWindow)
	}
	if as.TotalArtifactsIngested < 0 {
		return fmt.Errorf("%w: negative artifacts count", ErrValidationFailed)
	}
	if len(as.DomainSchemaIDs) > MaxDomainWeights {
		return fmt.Errorf("%w: domain_schema_ids cardinality exceeded", ErrCardinalityExceeded)
	}
	return nil
}

// CandidateSnapshot represents an immutable proposed strategy update or structural schema produced by a Learner.
type CandidateSnapshot struct {
	SnapshotID           string                      `json:"snapshot_id"`
	SemVer               string                      `json:"sem_ver"`
	SchemaID             string                      `json:"schema_id"`
	Lifecycle            CandidateLifecycle          `json:"lifecycle"`
	Lineage              ReplayMetadata              `json:"lineage"`
	Provenance           *CandidateLineage           `json:"provenance,omitempty"`
	Payload              []byte                      `json:"payload"`
	ValidationHash       string                      `json:"validation_hash,omitempty"`
	StructuralValidation *StructuralValidationResult `json:"structural_validation,omitempty"`
	ValidationRecords    []ValidationResult          `json:"validation_records,omitempty"`
}

// Validate verifies the candidate snapshot structure, lifecycle bounds, and payload limits.
func (cs *CandidateSnapshot) Validate() error {
	if cs.SnapshotID == "" || cs.SemVer == "" || cs.SchemaID == "" {
		return fmt.Errorf("%w: missing snapshot identification fields", ErrMissingID)
	}
	switch cs.Lifecycle {
	case LifecycleDraft, LifecycleValidated, LifecycleShadow, LifecycleCanary, LifecycleActive, LifecycleRetired, LifecycleRolledBack:
		// Valid
	default:
		return fmt.Errorf("%w: %q", ErrInvalidLifecycle, cs.Lifecycle)
	}
	if len(cs.Payload) > MaxPayloadBytes {
		return fmt.Errorf("%w: payload length %d exceeds max %d", ErrPayloadTooLarge, len(cs.Payload), MaxPayloadBytes)
	}
	if len(cs.ValidationRecords) > MaxValidationRecordsPerSnapshot {
		return fmt.Errorf("%w: validation records cardinality exceeded", ErrCardinalityExceeded)
	}
	if cs.Provenance != nil {
		if err := cs.Provenance.Validate(); err != nil {
			return fmt.Errorf("provenance validation failed: %w", err)
		}
	}
	if err := cs.Lineage.Validate(); err != nil {
		return fmt.Errorf("lineage validation failed: %w", err)
	}
	if cs.StructuralValidation != nil {
		if err := cs.StructuralValidation.Validate(); err != nil {
			return fmt.Errorf("structural validation failed: %w", err)
		}
	}
	for i, vr := range cs.ValidationRecords {
		if err := vr.Validate(); err != nil {
			return fmt.Errorf("validation record [%d] failed: %w", i, err)
		}
	}
	return nil
}

// ValidationSummary provides bounded O(1) telemetry on validation firewall activity across candidates.
type ValidationSummary struct {
	TotalValidated       int `json:"total_validated"`
	Passed               int `json:"passed"`
	Failed               int `json:"failed"`
	StructuralRejections int `json:"structural_rejections"`
}

// CandidateSummary provides bounded O(1) telemetry on candidate snapshot distribution by lifecycle.
type CandidateSummary struct {
	TotalProduced  int `json:"total_produced"`
	DraftCount     int `json:"draft_count"`
	ValidatedCount int `json:"validated_count"`
	ActiveCount    int `json:"active_count"`
}

// LearningCampaign groups related learning cycles under one objective scheduled by Executive.
// Learning records campaign metadata and increments mechanical counters; Learning must never evaluate campaign success.
type LearningCampaign struct {
	CampaignID          string         `json:"campaign_id"`
	Objective           string         `json:"objective"`
	StartTime           time.Time      `json:"start_time"`
	EndTime             time.Time      `json:"end_time"`
	CampaignFingerprint string         `json:"campaign_fingerprint"`
	PolicyFingerprint   PolicyFingerprint `json:"policy_fingerprint"`
	CampaignStatus      CampaignStatus `json:"campaign_status"`
	LearningCycles      uint64         `json:"learning_cycles"`
	CandidatesGenerated uint64         `json:"candidates_generated"`
	CandidatesValidated uint64         `json:"candidates_validated"`
	ExperimentsCreated  uint64         `json:"experiments_created"`
	SnapshotsPublished  uint64         `json:"snapshots_published"`
}

// Validate checks structural integrity and boundaries of LearningCampaign.
func (lc *LearningCampaign) Validate() error {
	if lc.CampaignID == "" || lc.CampaignFingerprint == "" || lc.PolicyFingerprint == "" {
		return fmt.Errorf("%w: missing campaign identification fields", ErrMissingID)
	}
	if err := validateStringLength("objective", lc.Objective); err != nil {
		return err
	}
	switch lc.CampaignStatus {
	case CampaignStatusScheduled, CampaignStatusActive, CampaignStatusCompleted, CampaignStatusPaused, CampaignStatusAborted:
		// Valid
	default:
		return fmt.Errorf("%w: %q", ErrInvalidCampaignStatus, lc.CampaignStatus)
	}
	if !lc.EndTime.IsZero() && !lc.StartTime.IsZero() && lc.EndTime.Before(lc.StartTime) {
		return fmt.Errorf("%w: campaign end time before start time", ErrInvalidTimeWindow)
	}
	return nil
}

// LearningCampaignSummary provides strictly bounded aggregate telemetry for Reflection and long-term analysis.
type LearningCampaignSummary struct {
	CampaignID                  string        `json:"campaign_id"`
	TotalCycles                 uint64        `json:"total_cycles"`
	TotalCandidates             uint64        `json:"total_candidates"`
	TotalValidated              uint64        `json:"total_validated"`
	TotalPublished              uint64        `json:"total_published"`
	TotalAbstained              uint64        `json:"total_abstained"`
	AverageValidationConfidence float64       `json:"average_validation_confidence"`
	AverageExecutionTimeUs      uint64        `json:"average_execution_time_us"`
	AverageReplayFidelity       float64       `json:"average_replay_fidelity"`
	CampaignDuration            time.Duration `json:"campaign_duration"`
}

// Validate checks structural integrity and boundaries of LearningCampaignSummary.
func (lcs *LearningCampaignSummary) Validate() error {
	if lcs.CampaignID == "" {
		return fmt.Errorf("%w: missing campaign_id", ErrMissingID)
	}
	if lcs.AverageValidationConfidence < 0.0 || lcs.AverageValidationConfidence > 1.0 {
		return fmt.Errorf("%w: average_validation_confidence out of bounds", ErrInvalidConfidence)
	}
	if lcs.AverageReplayFidelity < 0.0 || lcs.AverageReplayFidelity > 1.0 {
		return fmt.Errorf("%w: average_replay_fidelity out of bounds", ErrInvalidConfidence)
	}
	return nil
}

// LearningRequest represents an incoming request or time-window trigger to run a learning cycle.
type LearningRequest struct {
	RequestID         string            `json:"request_id"`
	DomainSchemaID    string            `json:"domain_schema_id"`
	CampaignID        string            `json:"campaign_id,omitempty"`
	TimeWindowStart   time.Time         `json:"time_window_start"`
	TimeWindowEnd     time.Time         `json:"time_window_end"`
	TargetSnapshotID  string            `json:"target_snapshot_id,omitempty"`
	PolicyFingerprint PolicyFingerprint `json:"policy_fingerprint"`
}

// Validate checks the request fields.
func (lr *LearningRequest) Validate() error {
	if lr.RequestID == "" || lr.DomainSchemaID == "" || lr.PolicyFingerprint == "" {
		return fmt.Errorf("%w: missing request_id, domain_schema_id, or policy_fingerprint", ErrMissingID)
	}
	if !lr.TimeWindowEnd.IsZero() && !lr.TimeWindowStart.IsZero() && lr.TimeWindowEnd.Before(lr.TimeWindowStart) {
		return fmt.Errorf("%w: end time before start time", ErrInvalidTimeWindow)
	}
	return nil
}

// TraceStatisticalSummary captures compact O(1) telemetry aggregates across the analyzed corpus
// for a single learning cycle, enabling bounded trend evaluation by Reflection without memory bloat.
type TraceStatisticalSummary struct {
	TotalArtifactsAnalyzed int                        `json:"total_artifacts_analyzed"`
	MeanValidationScore    float64                    `json:"mean_validation_score"`
	MinValidationScore     float64                    `json:"min_validation_score"`
	MaxValidationScore     float64                    `json:"max_validation_score"`
	EstimatedDriftScore    float64                    `json:"estimated_drift_score"`
	ReplayFidelityRatio    float64                    `json:"replay_fidelity_ratio"`
	DomainCoverageRatio    float64                    `json:"domain_coverage_ratio"`
	RejectionSummary       *CandidateRejectionSummary `json:"rejection_summary,omitempty"`
}

// Validate verifies statistical summary boundaries and non-negative counts.
func (ts *TraceStatisticalSummary) Validate() error {
	if ts.TotalArtifactsAnalyzed < 0 {
		return fmt.Errorf("%w: total_artifacts_analyzed cannot be negative", ErrValidationFailed)
	}
	if math.IsNaN(ts.MeanValidationScore) || math.IsNaN(ts.EstimatedDriftScore) || math.IsNaN(ts.ReplayFidelityRatio) {
		return fmt.Errorf("%w: statistical summary contains NaN scores", ErrValidationFailed)
	}
	if ts.RejectionSummary != nil {
		if err := ts.RejectionSummary.Validate(); err != nil {
			return fmt.Errorf("rejection summary validation failed: %w", err)
		}
	}
	return nil
}

// LearningTrace records diagnostic telemetry for a completed learning cycle.
// LearningTrace is write-only from Learning's perspective and consumed strictly by Reflection.
type LearningTrace struct {
	TraceID             string                    `json:"trace_id"`
	RequestID           string                    `json:"request_id"`
	DomainSchemaID      string                    `json:"domain_schema_id"`
	CampaignID          string                    `json:"campaign_id,omitempty"`
	CampaignSummary     *LearningCampaignSummary  `json:"campaign_summary,omitempty"`
	StatisticalSummary  *TraceStatisticalSummary  `json:"statistical_summary,omitempty"`
	LearningFingerprint LearningFingerprint       `json:"learning_fingerprint,omitempty"`
	PolicyFingerprint   PolicyFingerprint         `json:"policy_fingerprint"`
	LearnerFingerprint  string                    `json:"learner_fingerprint,omitempty"`
	ReplaySeed          uint64                    `json:"replay_seed,omitempty"`
	SourceArtifactHash  SourceArtifactHash        `json:"source_artifact_hash,omitempty"`
	ParentSnapshot      string                    `json:"parent_snapshot,omitempty"`
	AncestorSnapshot    string                    `json:"ancestor_snapshot,omitempty"`
	GenerationDepth     uint32                    `json:"generation_depth,omitempty"`
	Lineage             ReplayMetadata               `json:"lineage"`
	Aggregation         AggregationSummary           `json:"aggregation"`
	LearnerUsages       []LearnerUsage               `json:"learner_usages"`
	LearnerPerformance  []*LearnerPerformanceSummary `json:"learner_performance,omitempty"`
	CandidateCount      int                          `json:"candidate_count"`
	Status              LearningResultStatus         `json:"status"`
	TerminationReason   LearningTerminationReason    `json:"termination_reason"`
	TotalDuration       time.Duration                `json:"total_duration"`
	TraceTimestamp      time.Time                    `json:"trace_timestamp"`
}

// Validate checks the trace fields and cardinality limits.
func (lt *LearningTrace) Validate() error {
	if lt.TraceID == "" || lt.RequestID == "" || lt.DomainSchemaID == "" {
		return fmt.Errorf("%w: missing trace identification fields", ErrMissingID)
	}
	switch lt.Status {
	case StatusPublished, StatusValidationFail, StatusAbstained, StatusNoChange, StatusPartial, StatusNoCandidates, StatusRolledBack:
		// Valid
	default:
		return fmt.Errorf("%w: %q", ErrInvalidStatus, lt.Status)
	}
	switch lt.TerminationReason {
	case ReasonSuccess, ReasonNoCandidates, ReasonSampleFloorNotMet, ReasonCooldownActive, ReasonDriftAnomalyDetected, ReasonCapabilityUnavailable, ReasonTimeout:
		// Valid
	default:
		return fmt.Errorf("%w: %q", ErrInvalidTerminationReason, lt.TerminationReason)
	}
	if len(lt.LearnerUsages) > MaxLearnerUsagesPerTrace {
		return fmt.Errorf("%w: learner usages cardinality exceeded (%d > %d)", ErrCardinalityExceeded, len(lt.LearnerUsages), MaxLearnerUsagesPerTrace)
	}
	if lt.TotalDuration < 0 {
		return fmt.Errorf("%w: negative total duration", ErrValidationFailed)
	}
	if err := lt.Aggregation.Validate(); err != nil {
		return fmt.Errorf("trace aggregation validation failed: %w", err)
	}
	if lt.CampaignSummary != nil {
		if err := lt.CampaignSummary.Validate(); err != nil {
			return fmt.Errorf("trace campaign summary validation failed: %w", err)
		}
	}
	if lt.StatisticalSummary != nil {
		if err := lt.StatisticalSummary.Validate(); err != nil {
			return fmt.Errorf("trace statistical summary validation failed: %w", err)
		}
	}
	for i, lu := range lt.LearnerUsages {
		if err := lu.Validate(); err != nil {
			return fmt.Errorf("trace learner usage [%d] failed: %w", i, err)
		}
	}
	for i, lp := range lt.LearnerPerformance {
		if lp == nil {
			return fmt.Errorf("%w: trace learner performance [%d] is nil", ErrValidationFailed, i)
		}
		if err := lp.Validate(); err != nil {
			return fmt.Errorf("trace learner performance [%d] failed: %w", i, err)
		}
	}
	return nil
}

// LearningResult encapsulates the complete output of a learning cycle invocation.
type LearningResult struct {
	ResultID          string                    `json:"result_id"`
	RequestID         string                    `json:"request_id"`
	CampaignID        string                    `json:"campaign_id,omitempty"`
	CampaignSummary   *LearningCampaignSummary  `json:"campaign_summary,omitempty"`
	Status            LearningResultStatus      `json:"status"`
	TerminationReason LearningTerminationReason `json:"termination_reason"`
	Candidates        []*CandidateSnapshot      `json:"candidates"`
	Traces            []*LearningTrace          `json:"traces"`
	TotalDuration     time.Duration             `json:"total_duration"`
}

// Validate checks the result structure and cardinality limits.
func (lr *LearningResult) Validate() error {
	if lr.ResultID == "" || lr.RequestID == "" {
		return fmt.Errorf("%w: missing result_id or request_id", ErrMissingID)
	}
	switch lr.Status {
	case StatusPublished, StatusValidationFail, StatusAbstained, StatusNoChange, StatusPartial, StatusNoCandidates, StatusRolledBack:
		// Valid
	default:
		return fmt.Errorf("%w: %q", ErrInvalidStatus, lr.Status)
	}
	switch lr.TerminationReason {
	case ReasonSuccess, ReasonNoCandidates, ReasonSampleFloorNotMet, ReasonCooldownActive, ReasonDriftAnomalyDetected, ReasonCapabilityUnavailable, ReasonTimeout:
		// Valid
	default:
		return fmt.Errorf("%w: %q", ErrInvalidTerminationReason, lr.TerminationReason)
	}
	if len(lr.Candidates) > MaxCandidatesPerResult {
		return fmt.Errorf("%w: candidates cardinality limit exceeded", ErrCardinalityExceeded)
	}
	if lr.TotalDuration < 0 {
		return fmt.Errorf("%w: negative total duration", ErrValidationFailed)
	}
	if lr.CampaignSummary != nil {
		if err := lr.CampaignSummary.Validate(); err != nil {
			return fmt.Errorf("result campaign summary validation failed: %w", err)
		}
	}
	for i, c := range lr.Candidates {
		if c == nil {
			return fmt.Errorf("%w: candidate [%d] is nil", ErrValidationFailed, i)
		}
		if err := c.Validate(); err != nil {
			return fmt.Errorf("result candidate [%d] failed: %w", i, err)
		}
	}
	for i, t := range lr.Traces {
		if t == nil {
			return fmt.Errorf("%w: trace [%d] is nil", ErrValidationFailed, i)
		}
		if err := t.Validate(); err != nil {
			return fmt.Errorf("result trace [%d] failed: %w", i, err)
		}
	}
	return nil
}

// LearningStrategySnapshot represents the immutable strategy package read atomically by cognitive drivers.
type LearningStrategySnapshot struct {
	SnapshotID        string                 `json:"snapshot_id"`
	SchemaVersion     string                 `json:"schema_version"`
	ActiveProfile     *LearningPolicyProfile `json:"active_profile"`
	Capabilities      *LearningCapabilities  `json:"capabilities"`
	AggregationPolicy *AggregationPolicy     `json:"aggregation_policy,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
}

// Validate checks the strategy snapshot integrity.
func (s *LearningStrategySnapshot) Validate() error {
	if s.SnapshotID == "" {
		return fmt.Errorf("%w: missing snapshot_id", ErrMissingID)
	}
	if s.SchemaVersion != SchemaVersion {
		return fmt.Errorf("%w: schema version mismatch (%s != %s)", ErrInvalidSchemaVersion, s.SchemaVersion, SchemaVersion)
	}
	if s.ActiveProfile == nil || s.Capabilities == nil {
		return fmt.Errorf("%w: active_profile and capabilities must not be nil", ErrValidationFailed)
	}
	if err := s.ActiveProfile.Validate(); err != nil {
		return fmt.Errorf("snapshot profile validation failed: %w", err)
	}
	if err := s.Capabilities.Validate(); err != nil {
		return fmt.Errorf("snapshot capabilities validation failed: %w", err)
	}
	if s.AggregationPolicy != nil {
		if err := s.AggregationPolicy.Validate(); err != nil {
			return fmt.Errorf("snapshot aggregation policy validation failed: %w", err)
		}
	}
	return nil
}

// LearnerPerformanceSummary captures bounded O(1) statistical counters summarizing
// historical performance for a single registered learner across multiple learning cycles.
type LearnerPerformanceSummary struct {
	LearnerID              string  `json:"learner_id"`
	LearnerVersion         string  `json:"learner_version"`
	LearnerFingerprint     string  `json:"learner_fingerprint"`
	Executions             uint64  `json:"executions"`
	AcceptedCandidates     uint64  `json:"accepted_candidates"`
	RejectedCandidates     uint64  `json:"rejected_candidates"`
	AverageValidationScore float64 `json:"average_validation_score"`
	AverageExecutionTimeUs uint64  `json:"average_execution_time_us"`
	SuccessRatio           float64 `json:"success_ratio"`
}

// Validate checks the learner performance summary boundaries and non-negative counts.
func (lps *LearnerPerformanceSummary) Validate() error {
	if lps.LearnerID == "" {
		return fmt.Errorf("%w: missing learner_id in performance summary", ErrMissingID)
	}
	if math.IsNaN(lps.AverageValidationScore) || math.IsNaN(lps.SuccessRatio) {
		return fmt.Errorf("%w: performance summary contains NaN score or ratio", ErrValidationFailed)
	}
	if lps.SuccessRatio < 0.0 || lps.SuccessRatio > 1.0 {
		return fmt.Errorf("%w: success_ratio out of bounds [0.0, 1.0] (%f)", ErrValidationFailed, lps.SuccessRatio)
	}
	return nil
}

// CandidateRejectionSummary aggregates rejection counts by category across candidate evaluation.
type CandidateRejectionSummary struct {
	TotalRejected                    uint64 `json:"total_rejected"`
	StatisticalValidationFailures    uint64 `json:"statistical_validation_failures"`
	StructuralValidationFailures     uint64 `json:"structural_validation_failures"`
	ReplayValidationFailures         uint64 `json:"replay_validation_failures"`
	CapabilityValidationFailures     uint64 `json:"capability_validation_failures"`
	ConstitutionalValidationFailures uint64 `json:"constitutional_validation_failures"`
	DuplicateCandidateFailures       uint64 `json:"duplicate_candidate_failures"`
	OtherFailures                    uint64 `json:"other_failures"`
}

// Validate verifies non-negative aggregate counts and category consistency.
func (crs *CandidateRejectionSummary) Validate() error {
	sum := crs.StatisticalValidationFailures + crs.StructuralValidationFailures +
		crs.ReplayValidationFailures + crs.CapabilityValidationFailures +
		crs.ConstitutionalValidationFailures + crs.DuplicateCandidateFailures + crs.OtherFailures
	if crs.TotalRejected < sum {
		return fmt.Errorf("%w: total_rejected (%d) is less than sum of specific category failures (%d)", ErrValidationFailed, crs.TotalRejected, sum)
	}
	return nil
}
