package attention

import (
	"time"

	"idun/core/logger"
)

// StimulusBuilder constructs and validates a Stimulus artifact.
type StimulusBuilder struct {
	id            string
	source        string
	payloadRef    string
	safetyFlag    bool
	salienceScore int
}

// NewStimulusBuilder creates a new StimulusBuilder.
func NewStimulusBuilder() *StimulusBuilder {
	return &StimulusBuilder{}
}

func (b *StimulusBuilder) WithID(id string) *StimulusBuilder {
	b.id = id
	return b
}

func (b *StimulusBuilder) WithSource(source string) *StimulusBuilder {
	b.source = source
	return b
}

func (b *StimulusBuilder) WithPayloadRef(payloadRef string) *StimulusBuilder {
	b.payloadRef = payloadRef
	return b
}

func (b *StimulusBuilder) WithSafetyFlag(safety bool) *StimulusBuilder {
	b.safetyFlag = safety
	return b
}

func (b *StimulusBuilder) WithSalienceScore(score int) *StimulusBuilder {
	b.salienceScore = score
	return b
}

func (b *StimulusBuilder) Build() (Stimulus, error) {
	s := Stimulus{
		ID:            b.id,
		Source:        b.source,
		PayloadRef:    b.payloadRef,
		SafetyFlag:    b.safetyFlag,
		SalienceScore: b.salienceScore,
	}
	if err := s.Validate(); err != nil {
		return Stimulus{}, err
	}
	return s, nil
}

// AttentionPolicyProfileBuilder constructs and validates an AttentionPolicyProfile.
type AttentionPolicyProfileBuilder struct {
	profile *AttentionPolicyProfile
}

// NewPolicyProfileBuilder initializes a builder with default settings.
func NewPolicyProfileBuilder() *AttentionPolicyProfileBuilder {
	return &AttentionPolicyProfileBuilder{
		profile: DefaultAttentionPolicyProfile(),
	}
}

func (b *AttentionPolicyProfileBuilder) WithVersions(policyVersion, schemaVersion string) *AttentionPolicyProfileBuilder {
	b.profile.PolicyVersion = policyVersion
	b.profile.SchemaVersion = schemaVersion
	return b
}

func (b *AttentionPolicyProfileBuilder) WithThresholds(b0, b1, b2, b3 int) *AttentionPolicyProfileBuilder {
	b.profile.Band0Threshold = b0
	b.profile.Band1Threshold = b1
	b.profile.Band2Threshold = b2
	b.profile.Band3Threshold = b3
	return b
}

func (b *AttentionPolicyProfileBuilder) WithMargins(switchMargin, interruptMargin int) *AttentionPolicyProfileBuilder {
	b.profile.SwitchMargin = switchMargin
	b.profile.InterruptMargin = interruptMargin
	return b
}

func (b *AttentionPolicyProfileBuilder) WithMaximumTracked(max int) *AttentionPolicyProfileBuilder {
	b.profile.MaximumTrackedStimuli = max
	return b
}

func (b *AttentionPolicyProfileBuilder) Build() (*AttentionPolicyProfile, error) {
	fp, err := ComputePolicyFingerprint(b.profile)
	if err != nil {
		return nil, err
	}
	b.profile.PolicyFingerprint = fp
	if err := b.profile.Validate(); err != nil {
		return nil, err
	}
	return b.profile, nil
}

// AttentionCapabilitiesBuilder constructs and validates an AttentionCapabilities.
type AttentionCapabilitiesBuilder struct {
	caps *AttentionCapabilities
}

// NewCapabilitiesBuilder initializes a builder with default capabilities.
func NewCapabilitiesBuilder() *AttentionCapabilitiesBuilder {
	return &AttentionCapabilitiesBuilder{
		caps: DefaultAttentionCapabilities(),
	}
}

func (b *AttentionCapabilitiesBuilder) WithInterruptions(supports bool) *AttentionCapabilitiesBuilder {
	b.caps.SupportsInterruptions = supports
	return b
}

func (b *AttentionCapabilitiesBuilder) WithBackgroundAttention(supports bool) *AttentionCapabilitiesBuilder {
	b.caps.SupportsBackgroundAttention = supports
	return b
}

func (b *AttentionCapabilitiesBuilder) WithFocusSwitching(supports bool) *AttentionCapabilitiesBuilder {
	b.caps.SupportsFocusSwitching = supports
	return b
}

func (b *AttentionCapabilitiesBuilder) WithMultimodalAttention(supports bool) *AttentionCapabilitiesBuilder {
	b.caps.SupportsMultimodalAttention = supports
	return b
}

func (b *AttentionCapabilitiesBuilder) WithDistributedAttention(supports bool) *AttentionCapabilitiesBuilder {
	b.caps.SupportsDistributedAttention = supports
	return b
}

func (b *AttentionCapabilitiesBuilder) WithFocusHistory(supports bool) *AttentionCapabilitiesBuilder {
	b.caps.SupportsFocusHistory = supports
	return b
}

func (b *AttentionCapabilitiesBuilder) Build() (*AttentionCapabilities, error) {
	fp, err := ComputeCapabilityFingerprint(b.caps)
	if err != nil {
		return nil, err
	}
	b.caps.CapabilityFingerprint = fp
	if err := b.caps.Validate(); err != nil {
		return nil, err
	}
	return b.caps, nil
}

// ReplayMetadataBuilder constructs and validates an AttentionReplayMetadata.
type ReplayMetadataBuilder struct {
	meta *AttentionReplayMetadata
}

// NewReplayMetadataBuilder creates a new ReplayMetadataBuilder.
func NewReplayMetadataBuilder() *ReplayMetadataBuilder {
	return &ReplayMetadataBuilder{
		meta: &AttentionReplayMetadata{
			AttentionVersion: AttentionVersion,
		},
	}
}

func (b *ReplayMetadataBuilder) WithFingerprints(policyFP, capFP string) *ReplayMetadataBuilder {
	b.meta.PolicyFingerprint = policyFP
	b.meta.CapabilityFingerprint = capFP
	return b
}

func (b *ReplayMetadataBuilder) WithVersion(version string) *ReplayMetadataBuilder {
	b.meta.AttentionVersion = version
	return b
}

func (b *ReplayMetadataBuilder) WithSeed(seed int64) *ReplayMetadataBuilder {
	b.meta.ReplaySeed = seed
	return b
}

func (b *ReplayMetadataBuilder) Build() (AttentionReplayMetadata, error) {
	if err := b.meta.Validate(); err != nil {
		return AttentionReplayMetadata{}, err
	}
	return *b.meta, nil
}

// AttentionTraceBuilder constructs and validates an AttentionTrace.
type AttentionTraceBuilder struct {
	trace *AttentionTrace
}

// NewTraceBuilder initializes a new AttentionTraceBuilder.
func NewTraceBuilder() *AttentionTraceBuilder {
	return &AttentionTraceBuilder{
		trace: &AttentionTrace{
			AttentionVersion: AttentionVersion,
			EvaluatedAt:      time.Now().UTC(),
		},
	}
}

func (b *AttentionTraceBuilder) WithIdentifiers(traceID, stimulusID, source string) *AttentionTraceBuilder {
	b.trace.TraceID = traceID
	b.trace.StimulusID = stimulusID
	b.trace.StimulusSource = source
	return b
}

func (b *AttentionTraceBuilder) WithFocusTransition(before, after string, switchOccurred bool) *AttentionTraceBuilder {
	b.trace.FocusBefore = before
	b.trace.FocusAfter = after
	b.trace.SwitchOccurred = switchOccurred
	return b
}

func (b *AttentionTraceBuilder) WithDecision(dec SalienceDecision, band PriorityBand) *AttentionTraceBuilder {
	b.trace.Decision = dec
	b.trace.PriorityBand = band
	return b
}

func (b *AttentionTraceBuilder) WithOutcome(status AttentionResultStatus, reason AttentionTerminationReason) *AttentionTraceBuilder {
	b.trace.ResultStatus = status
	b.trace.TerminationReason = reason
	return b
}

func (b *AttentionTraceBuilder) WithExecutionTime(dur time.Duration) *AttentionTraceBuilder {
	b.trace.ExecutionTime = dur
	return b
}

func (b *AttentionTraceBuilder) WithReplay(replay AttentionReplayMetadata) *AttentionTraceBuilder {
	b.trace.ReplayMetadata = replay
	b.trace.PolicyFingerprint = replay.PolicyFingerprint
	b.trace.CapabilityFingerprint = replay.CapabilityFingerprint
	return b
}

func (b *AttentionTraceBuilder) Build() (*AttentionTrace, error) {
	if err := b.trace.Validate(); err != nil {
		return nil, err
	}
	return b.trace, nil
}

// AttentionSummaryBuilder constructs and validates an AttentionSummary.
type AttentionSummaryBuilder struct {
	summary *AttentionSummary
}

// NewSummaryBuilder initializes an AttentionSummaryBuilder.
func NewSummaryBuilder() *AttentionSummaryBuilder {
	return &AttentionSummaryBuilder{
		summary: &AttentionSummary{},
	}
}

func (b *AttentionSummaryBuilder) WithCounts(total, immediate, scheduled, filtered, intAcc, intRej, switches int64) *AttentionSummaryBuilder {
	b.summary.TotalStimuli = total
	b.summary.ImmediateFocusCount = immediate
	b.summary.ScheduledCount = scheduled
	b.summary.FilteredCount = filtered
	b.summary.InterruptAccepted = intAcc
	b.summary.InterruptRejected = intRej
	b.summary.FocusSwitches = switches
	return b
}

func (b *AttentionSummaryBuilder) WithTimes(avg, total time.Duration) *AttentionSummaryBuilder {
	b.summary.AverageEvaluationTime = avg
	b.summary.TotalEvaluationTime = total
	return b
}

func (b *AttentionSummaryBuilder) Build() (AttentionSummary, error) {
	if err := b.summary.Validate(); err != nil {
		return AttentionSummary{}, err
	}
	return *b.summary, nil
}

// FocusHistoryBuilder constructs and validates a FocusHistoryEntry.
type FocusHistoryBuilder struct {
	entry *FocusHistoryEntry
}

// NewFocusHistoryBuilder creates a new FocusHistoryBuilder.
func NewFocusHistoryBuilder() *FocusHistoryBuilder {
	return &FocusHistoryBuilder{
		entry: &FocusHistoryEntry{
			Timestamp: time.Now().UTC(),
		},
	}
}

func (b *FocusHistoryBuilder) WithTransition(prev, curr, reason string) *FocusHistoryBuilder {
	b.entry.PreviousFocus = prev
	b.entry.CurrentFocus = curr
	b.entry.SwitchReason = reason
	return b
}

func (b *FocusHistoryBuilder) WithTimestamp(ts time.Time) *FocusHistoryBuilder {
	b.entry.Timestamp = ts
	return b
}

func (b *FocusHistoryBuilder) Build() (FocusHistoryEntry, error) {
	if err := b.entry.Validate(); err != nil {
		return FocusHistoryEntry{}, err
	}
	return *b.entry, nil
}

// AttentionEventSummaryBuilder constructs and validates an AttentionEventSummary.
type AttentionEventSummaryBuilder struct {
	summary *AttentionEventSummary
}

// NewEventSummaryBuilder creates a new AttentionEventSummaryBuilder.
func NewEventSummaryBuilder() *AttentionEventSummaryBuilder {
	return &AttentionEventSummaryBuilder{
		summary: &AttentionEventSummary{},
	}
}

func (b *AttentionEventSummaryBuilder) WithCounters(focus, intAcc, intRej, goal, safety int64) *AttentionEventSummaryBuilder {
	b.summary.FocusChangedCount = focus
	b.summary.InterruptAcceptedCount = intAcc
	b.summary.InterruptRejectedCount = intRej
	b.summary.GoalSwitchCount = goal
	b.summary.SafetyTripwireCount = safety
	return b
}

func (b *AttentionEventSummaryBuilder) Build() (AttentionEventSummary, error) {
	if err := b.summary.Validate(); err != nil {
		return AttentionEventSummary{}, err
	}
	return *b.summary, nil
}

// ConfigurationBuilder constructs and validates a Configuration.
type ConfigurationBuilder struct {
	cfg *Configuration
}

// NewConfigurationBuilder creates a new ConfigurationBuilder.
func NewConfigurationBuilder() *ConfigurationBuilder {
	return &ConfigurationBuilder{
		cfg: DefaultConfiguration(),
	}
}

func (b *ConfigurationBuilder) WithLogger(log logger.Writer) *ConfigurationBuilder {
	b.cfg.Logger = log
	return b
}

func (b *ConfigurationBuilder) WithPolicyProfile(profile *AttentionPolicyProfile) *ConfigurationBuilder {
	b.cfg.PolicyProfile = profile
	return b
}

func (b *ConfigurationBuilder) WithCapabilities(caps *AttentionCapabilities) *ConfigurationBuilder {
	b.cfg.Capabilities = caps
	return b
}

func (b *ConfigurationBuilder) WithReplaySeed(seed int64) *ConfigurationBuilder {
	b.cfg.ReplaySeed = seed
	return b
}

func (b *ConfigurationBuilder) WithWorkspacePublisher(pub WorkspacePublisher) *ConfigurationBuilder {
	b.cfg.WorkspacePublisher = pub
	return b
}

func (b *ConfigurationBuilder) Build() (*Configuration, error) {
	if err := b.cfg.Validate(); err != nil {
		return nil, err
	}
	return b.cfg, nil
}

// ServiceBuilder constructs and initializes a new Attention Service.
type ServiceBuilder struct {
	cfg *Configuration
}

// NewServiceBuilder creates a ServiceBuilder from a Configuration.
func NewServiceBuilder(cfg *Configuration) *ServiceBuilder {
	if cfg == nil {
		cfg = DefaultConfiguration()
	}
	return &ServiceBuilder{cfg: cfg}
}

func (b *ServiceBuilder) Build() (*Service, error) {
	if err := b.cfg.Validate(); err != nil {
		return nil, err
	}
	svc := NewServiceFromConfig(b.cfg)
	return svc, nil
}
