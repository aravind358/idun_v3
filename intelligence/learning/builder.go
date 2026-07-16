package learning

import (
	"fmt"
	"time"
)

// ============================================================================
// LearningRequestBuilder
// ============================================================================

type LearningRequestBuilder struct {
	req *LearningRequest
}

func NewLearningRequestBuilder() *LearningRequestBuilder {
	return &LearningRequestBuilder{
		req: &LearningRequest{},
	}
}

func (b *LearningRequestBuilder) WithRequestID(id string) *LearningRequestBuilder {
	b.req.RequestID = id
	return b
}

func (b *LearningRequestBuilder) WithDomainSchemaID(id string) *LearningRequestBuilder {
	b.req.DomainSchemaID = id
	return b
}

func (b *LearningRequestBuilder) WithCampaignID(id string) *LearningRequestBuilder {
	b.req.CampaignID = id
	return b
}

func (b *LearningRequestBuilder) WithTimeWindow(start, end time.Time) *LearningRequestBuilder {
	b.req.TimeWindowStart = start
	b.req.TimeWindowEnd = end
	return b
}

func (b *LearningRequestBuilder) WithTargetSnapshotID(id string) *LearningRequestBuilder {
	b.req.TargetSnapshotID = id
	return b
}

func (b *LearningRequestBuilder) WithPolicyFingerprint(fp PolicyFingerprint) *LearningRequestBuilder {
	b.req.PolicyFingerprint = fp
	return b
}

func (b *LearningRequestBuilder) Build() (*LearningRequest, error) {
	if err := b.req.Validate(); err != nil {
		return nil, fmt.Errorf("learning request builder failed validation: %w", err)
	}
	return b.req, nil
}

// ============================================================================
// LearningTraceBuilder
// ============================================================================

type LearningTraceBuilder struct {
	trace *LearningTrace
}

func NewLearningTraceBuilder(traceID, requestID, domainSchemaID string, policyFP PolicyFingerprint) *LearningTraceBuilder {
	return &LearningTraceBuilder{
		trace: &LearningTrace{
			TraceID:           traceID,
			RequestID:         requestID,
			DomainSchemaID:    domainSchemaID,
			PolicyFingerprint: policyFP,
			Status:            StatusPublished,
			TerminationReason: ReasonSuccess,
			TraceTimestamp:    time.Now(),
		},
	}
}

func (b *LearningTraceBuilder) WithAggregation(agg AggregationSummary) *LearningTraceBuilder {
	b.trace.Aggregation = agg
	return b
}

func (b *LearningTraceBuilder) WithLineage(meta ReplayMetadata) *LearningTraceBuilder {
	b.trace.Lineage = meta
	if meta.LearningFingerprint != "" {
		b.trace.LearningFingerprint = meta.LearningFingerprint
	}
	if meta.PolicyFingerprint != "" {
		b.trace.PolicyFingerprint = meta.PolicyFingerprint
	}
	if meta.LearnerFingerprint != "" {
		b.trace.LearnerFingerprint = meta.LearnerFingerprint
	}
	b.trace.ReplaySeed = meta.ReplaySeed
	if meta.SourceArtifactHash != "" {
		b.trace.SourceArtifactHash = meta.SourceArtifactHash
	}
	b.trace.ParentSnapshot = meta.ParentSnapshot
	b.trace.AncestorSnapshot = meta.AncestorSnapshot
	b.trace.GenerationDepth = meta.GenerationDepth
	return b
}

func (b *LearningTraceBuilder) AddLearnerUsage(lu LearnerUsage) *LearningTraceBuilder {
	b.trace.LearnerUsages = append(b.trace.LearnerUsages, lu)
	return b
}

func (b *LearningTraceBuilder) WithCandidateCount(count int) *LearningTraceBuilder {
	b.trace.CandidateCount = count
	return b
}

func (b *LearningTraceBuilder) WithStatus(status LearningResultStatus) *LearningTraceBuilder {
	b.trace.Status = status
	return b
}

func (b *LearningTraceBuilder) WithTerminationReason(reason LearningTerminationReason) *LearningTraceBuilder {
	b.trace.TerminationReason = reason
	return b
}

func (b *LearningTraceBuilder) WithTotalDuration(dur time.Duration) *LearningTraceBuilder {
	b.trace.TotalDuration = dur
	return b
}

func (b *LearningTraceBuilder) WithTimestamp(ts time.Time) *LearningTraceBuilder {
	b.trace.TraceTimestamp = ts
	return b
}

func (b *LearningTraceBuilder) WithCampaignID(campaignID string) *LearningTraceBuilder {
	b.trace.CampaignID = campaignID
	return b
}

func (b *LearningTraceBuilder) WithCampaignSummary(summary *LearningCampaignSummary) *LearningTraceBuilder {
	b.trace.CampaignSummary = summary
	return b
}

func (b *LearningTraceBuilder) WithStatisticalSummary(stats *TraceStatisticalSummary) *LearningTraceBuilder {
	b.trace.StatisticalSummary = stats
	return b
}

func (b *LearningTraceBuilder) WithLearnerPerformance(perf []*LearnerPerformanceSummary) *LearningTraceBuilder {
	b.trace.LearnerPerformance = perf
	return b
}

func (b *LearningTraceBuilder) Build() (*LearningTrace, error) {
	if err := b.trace.Validate(); err != nil {
		return nil, fmt.Errorf("learning trace builder failed validation: %w", err)
	}
	return b.trace, nil
}

// ============================================================================
// CandidateSnapshotBuilder
// ============================================================================

type CandidateSnapshotBuilder struct {
	cs *CandidateSnapshot
}

func NewCandidateSnapshotBuilder(snapshotID, semVer, schemaID string, lineage ReplayMetadata) *CandidateSnapshotBuilder {
	return &CandidateSnapshotBuilder{
		cs: &CandidateSnapshot{
			SnapshotID: snapshotID,
			SemVer:     semVer,
			SchemaID:   schemaID,
			Lifecycle:  LifecycleDraft,
			Lineage:    lineage,
		},
	}
}

func (b *CandidateSnapshotBuilder) WithLifecycle(lc CandidateLifecycle) *CandidateSnapshotBuilder {
	b.cs.Lifecycle = lc
	return b
}

func (b *CandidateSnapshotBuilder) WithProvenance(parent, ancestor string, depth uint32) *CandidateSnapshotBuilder {
	b.cs.Provenance = &CandidateLineage{
		ParentSnapshot:   parent,
		AncestorSnapshot: ancestor,
		GenerationDepth:  depth,
	}
	b.cs.Lineage.ParentSnapshot = parent
	b.cs.Lineage.AncestorSnapshot = ancestor
	b.cs.Lineage.GenerationDepth = depth
	return b
}

func (b *CandidateSnapshotBuilder) WithPayload(payload []byte) *CandidateSnapshotBuilder {
	b.cs.Payload = append([]byte(nil), payload...)
	return b
}

func (b *CandidateSnapshotBuilder) WithValidationHash(hash string) *CandidateSnapshotBuilder {
	b.cs.ValidationHash = hash
	return b
}

func (b *CandidateSnapshotBuilder) WithStructuralValidation(svr *StructuralValidationResult) *CandidateSnapshotBuilder {
	b.cs.StructuralValidation = svr
	return b
}

func (b *CandidateSnapshotBuilder) AddValidationRecord(vr ValidationResult) *CandidateSnapshotBuilder {
	b.cs.ValidationRecords = append(b.cs.ValidationRecords, vr)
	return b
}

func (b *CandidateSnapshotBuilder) Build() (*CandidateSnapshot, error) {
	if err := b.cs.Validate(); err != nil {
		return nil, fmt.Errorf("candidate snapshot builder failed validation: %w", err)
	}
	return b.cs, nil
}

// ============================================================================
// ExperimentProfileBuilder
// ============================================================================

type ExperimentProfileBuilder struct {
	ep *ExperimentProfile
}

func NewExperimentProfileBuilder(expID, domainSchemaID, targetSnapshotID string) *ExperimentProfileBuilder {
	return &ExperimentProfileBuilder{
		ep: &ExperimentProfile{
			ExperimentID:     expID,
			DomainSchemaID:   domainSchemaID,
			TargetSnapshotID: targetSnapshotID,
			ShadowRatio:      0.10,
			CanaryRatio:      0.01,
			MaxDuration:      48 * time.Hour,
			ReplaySeed:       123456789,
		},
	}
}

func (b *ExperimentProfileBuilder) WithRatios(shadowRatio, canaryRatio float64) *ExperimentProfileBuilder {
	b.ep.ShadowRatio = shadowRatio
	b.ep.CanaryRatio = canaryRatio
	return b
}

func (b *ExperimentProfileBuilder) WithMaxDuration(dur time.Duration) *ExperimentProfileBuilder {
	b.ep.MaxDuration = dur
	return b
}

func (b *ExperimentProfileBuilder) WithPriority(priority int) *ExperimentProfileBuilder {
	b.ep.Priority = priority
	return b
}

func (b *ExperimentProfileBuilder) WithReplaySeed(seed uint64) *ExperimentProfileBuilder {
	b.ep.ReplaySeed = seed
	return b
}

func (b *ExperimentProfileBuilder) Build() (*ExperimentProfile, error) {
	if err := b.ep.Validate(); err != nil {
		return nil, fmt.Errorf("experiment profile builder failed validation: %w", err)
	}
	return b.ep, nil
}

// ============================================================================
// LearningCampaignBuilder
// ============================================================================

type LearningCampaignBuilder struct {
	campaign *LearningCampaign
}

func NewLearningCampaignBuilder(id, objective string, start, end time.Time, campFP string, policyFP PolicyFingerprint) *LearningCampaignBuilder {
	return &LearningCampaignBuilder{
		campaign: &LearningCampaign{
			CampaignID:          id,
			Objective:           objective,
			StartTime:           start,
			EndTime:             end,
			CampaignFingerprint: campFP,
			PolicyFingerprint:   policyFP,
			CampaignStatus:      CampaignStatusScheduled,
		},
	}
}

func (b *LearningCampaignBuilder) WithStatus(status CampaignStatus) *LearningCampaignBuilder {
	b.campaign.CampaignStatus = status
	return b
}

func (b *LearningCampaignBuilder) WithCounters(cycles, gen, val, exp, pub uint64) *LearningCampaignBuilder {
	b.campaign.LearningCycles = cycles
	b.campaign.CandidatesGenerated = gen
	b.campaign.CandidatesValidated = val
	b.campaign.ExperimentsCreated = exp
	b.campaign.SnapshotsPublished = pub
	return b
}

func (b *LearningCampaignBuilder) Build() (*LearningCampaign, error) {
	if err := b.campaign.Validate(); err != nil {
		return nil, fmt.Errorf("learning campaign builder failed validation: %w", err)
	}
	return b.campaign, nil
}

// ============================================================================
// LearningCampaignSummaryBuilder
// ============================================================================

type LearningCampaignSummaryBuilder struct {
	summary *LearningCampaignSummary
}

func NewLearningCampaignSummaryBuilder(campaignID string) *LearningCampaignSummaryBuilder {
	return &LearningCampaignSummaryBuilder{
		summary: &LearningCampaignSummary{
			CampaignID: campaignID,
		},
	}
}

func (b *LearningCampaignSummaryBuilder) WithTotals(cycles, cand, val, pub, abst uint64) *LearningCampaignSummaryBuilder {
	b.summary.TotalCycles = cycles
	b.summary.TotalCandidates = cand
	b.summary.TotalValidated = val
	b.summary.TotalPublished = pub
	b.summary.TotalAbstained = abst
	return b
}

func (b *LearningCampaignSummaryBuilder) WithAverages(conf float64, execTimeUs uint64, fidelity float64) *LearningCampaignSummaryBuilder {
	b.summary.AverageValidationConfidence = conf
	b.summary.AverageExecutionTimeUs = execTimeUs
	b.summary.AverageReplayFidelity = fidelity
	return b
}

func (b *LearningCampaignSummaryBuilder) WithDuration(dur time.Duration) *LearningCampaignSummaryBuilder {
	b.summary.CampaignDuration = dur
	return b
}

func (b *LearningCampaignSummaryBuilder) Build() (*LearningCampaignSummary, error) {
	if err := b.summary.Validate(); err != nil {
		return nil, fmt.Errorf("learning campaign summary builder failed validation: %w", err)
	}
	return b.summary, nil
}

// ============================================================================
// TraceStatisticalSummaryBuilder
// ============================================================================

type TraceStatisticalSummaryBuilder struct {
	summary *TraceStatisticalSummary
}

func NewTraceStatisticalSummaryBuilder(totalAnalyzed int) *TraceStatisticalSummaryBuilder {
	return &TraceStatisticalSummaryBuilder{
		summary: &TraceStatisticalSummary{
			TotalArtifactsAnalyzed: totalAnalyzed,
		},
	}
}

func (b *TraceStatisticalSummaryBuilder) WithValidationScores(mean, min, max float64) *TraceStatisticalSummaryBuilder {
	b.summary.MeanValidationScore = mean
	b.summary.MinValidationScore = min
	b.summary.MaxValidationScore = max
	return b
}

func (b *TraceStatisticalSummaryBuilder) WithDriftAndCoverage(drift, fidelity, coverage float64) *TraceStatisticalSummaryBuilder {
	b.summary.EstimatedDriftScore = drift
	b.summary.ReplayFidelityRatio = fidelity
	b.summary.DomainCoverageRatio = coverage
	return b
}

func (b *TraceStatisticalSummaryBuilder) WithRejectionSummary(rej *CandidateRejectionSummary) *TraceStatisticalSummaryBuilder {
	b.summary.RejectionSummary = rej
	return b
}

func (b *TraceStatisticalSummaryBuilder) Build() (*TraceStatisticalSummary, error) {
	if err := b.summary.Validate(); err != nil {
		return nil, fmt.Errorf("trace statistical summary builder failed validation: %w", err)
	}
	return b.summary, nil
}

// ============================================================================
// LearnerPerformanceSummaryBuilder
// ============================================================================

type LearnerPerformanceSummaryBuilder struct {
	summary *LearnerPerformanceSummary
}

func NewLearnerPerformanceSummaryBuilder(learnerID, version, fingerprint string) *LearnerPerformanceSummaryBuilder {
	return &LearnerPerformanceSummaryBuilder{
		summary: &LearnerPerformanceSummary{
			LearnerID:          learnerID,
			LearnerVersion:     version,
			LearnerFingerprint: fingerprint,
		},
	}
}

func (b *LearnerPerformanceSummaryBuilder) WithCounters(exec, accepted, rejected uint64) *LearnerPerformanceSummaryBuilder {
	b.summary.Executions = exec
	b.summary.AcceptedCandidates = accepted
	b.summary.RejectedCandidates = rejected
	return b
}

func (b *LearnerPerformanceSummaryBuilder) WithMetrics(avgScore float64, avgTimeUs uint64, successRatio float64) *LearnerPerformanceSummaryBuilder {
	b.summary.AverageValidationScore = avgScore
	b.summary.AverageExecutionTimeUs = avgTimeUs
	b.summary.SuccessRatio = successRatio
	return b
}

func (b *LearnerPerformanceSummaryBuilder) Build() (*LearnerPerformanceSummary, error) {
	if err := b.summary.Validate(); err != nil {
		return nil, fmt.Errorf("learner performance summary builder failed validation: %w", err)
	}
	return b.summary, nil
}

// ============================================================================
// CandidateRejectionSummaryBuilder
// ============================================================================

type CandidateRejectionSummaryBuilder struct {
	summary *CandidateRejectionSummary
}

func NewCandidateRejectionSummaryBuilder() *CandidateRejectionSummaryBuilder {
	return &CandidateRejectionSummaryBuilder{
		summary: &CandidateRejectionSummary{},
	}
}

func (b *CandidateRejectionSummaryBuilder) WithCounts(total, stat, structFail, replay, capFail, constFail, dup, other uint64) *CandidateRejectionSummaryBuilder {
	b.summary.TotalRejected = total
	b.summary.StatisticalValidationFailures = stat
	b.summary.StructuralValidationFailures = structFail
	b.summary.ReplayValidationFailures = replay
	b.summary.CapabilityValidationFailures = capFail
	b.summary.ConstitutionalValidationFailures = constFail
	b.summary.DuplicateCandidateFailures = dup
	b.summary.OtherFailures = other
	return b
}

func (b *CandidateRejectionSummaryBuilder) Build() (*CandidateRejectionSummary, error) {
	if err := b.summary.Validate(); err != nil {
		return nil, fmt.Errorf("candidate rejection summary builder failed validation: %w", err)
	}
	return b.summary, nil
}
