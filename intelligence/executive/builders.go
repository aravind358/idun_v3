package executive

import (
	"time"
)

// ExecutiveRequestBuilder builds an ExecutiveRequest with fluent validation.
type ExecutiveRequestBuilder struct {
	req *ExecutiveRequest
}

// NewExecutiveRequestBuilder initializes a new ExecutiveRequestBuilder.
func NewExecutiveRequestBuilder() *ExecutiveRequestBuilder {
	return &ExecutiveRequestBuilder{
		req: &ExecutiveRequest{
			Horizon:  HorizonDeliberative,
			Priority: PriorityBand2Interactive,
			Budget:   BudgetStandard,
		},
	}
}

func (b *ExecutiveRequestBuilder) WithRequestID(id string) *ExecutiveRequestBuilder {
	b.req.RequestID = id
	return b
}

func (b *ExecutiveRequestBuilder) WithEpisodeID(id string) *ExecutiveRequestBuilder {
	b.req.EpisodeID = id
	return b
}

func (b *ExecutiveRequestBuilder) WithWorkflow(wg *WorkflowGraph) *ExecutiveRequestBuilder {
	b.req.Workflow = wg
	return b
}

func (b *ExecutiveRequestBuilder) WithHorizon(h Horizon) *ExecutiveRequestBuilder {
	b.req.Horizon = h
	return b
}

func (b *ExecutiveRequestBuilder) WithPriority(p PriorityBand) *ExecutiveRequestBuilder {
	b.req.Priority = p
	return b
}

func (b *ExecutiveRequestBuilder) WithBudget(tier BudgetTier) *ExecutiveRequestBuilder {
	b.req.Budget = tier
	return b
}

func (b *ExecutiveRequestBuilder) WithActiveGoal(goal ActiveGoalContext) *ExecutiveRequestBuilder {
	b.req.ActiveGoal = goal
	return b
}

func (b *ExecutiveRequestBuilder) WithPolicyFingerprint(fp string) *ExecutiveRequestBuilder {
	b.req.PolicyFingerprint = fp
	return b
}

// Build constructs and validates the ExecutiveRequest.
func (b *ExecutiveRequestBuilder) Build() (*ExecutiveRequest, error) {
	if err := b.req.Validate(); err != nil {
		return nil, err
	}
	return b.req, nil
}

// ExecutivePolicyProfileBuilder builds an ExecutivePolicyProfile with fluent validation and fingerprinting.
type ExecutivePolicyProfileBuilder struct {
	profile *ExecutivePolicyProfile
}

// NewExecutivePolicyProfileBuilder initializes a new profile builder with valid default parameters.
func NewExecutivePolicyProfileBuilder() *ExecutivePolicyProfileBuilder {
	return &ExecutivePolicyProfileBuilder{
		profile: &ExecutivePolicyProfile{
			ProfileVersion: "1.0.0",
			SchemaVersion:  SchemaVersion,
			PolicySource:   "Learning",
			PriorityPolicies: PriorityPolicies{
				MaxConcurrentPerBand: map[PriorityBand]int{
					PriorityBand0CriticalSafety: 10,
					PriorityBand1RealTime:       20,
					PriorityBand2Interactive:    50,
					PriorityBand3Background:     100,
					PriorityBand4Idle:           200,
				},
				PreemptionAllowed:     true,
				EmergencyBand0Timeout: 5 * time.Second,
			},
			BudgetPolicies: BudgetPolicies{
				MaxCycleBudgetUnits:      10000,
				MaxFuelPerWorkflow:       500,
				MaxReflectionDepth:       3,
				EscalationUnitMultiplier: 1.5,
			},
			RetryPolicies: RetryPolicies{
				MaxRetries:        3,
				InitialBackoff:    50 * time.Millisecond,
				MaxBackoff:        2 * time.Second,
				BackoffMultiplier: 2.0,
			},
			CancellationPolicies: CancellationPolicies{
				CooperativeTimeout:  500 * time.Millisecond,
				ForceKillOnTimeout:  true,
				PropagateToChildren: true,
			},
			EscalationPolicies: EscalationPolicies{
				MaxEscalationsPerWorkflow:   3,
				RequireUserConfirmationTier: BudgetDeliberative,
				AmbiguityThreshold:          0.65,
			},
			WorkspacePolicies: WorkspacePolicies{
				AdmissionThreshold:     0.40,
				MaxPendingBidsPerTopic: 100,
				ImpasseUrgency:         50,
				ImpasseTimeout:         3 * time.Second,
			},
			HomeostasisPolicies: HomeostasisPolicies{
				IdleThreshold:         30 * time.Second,
				ConsolidationInterval: 5 * time.Minute,
				MaxPeriodicFuel:       200,
			},
		},
	}
}

func (b *ExecutivePolicyProfileBuilder) WithProfileID(id string) *ExecutivePolicyProfileBuilder {
	b.profile.ProfileID = id
	return b
}

func (b *ExecutivePolicyProfileBuilder) WithVersion(version string) *ExecutivePolicyProfileBuilder {
	b.profile.ProfileVersion = version
	return b
}

func (b *ExecutivePolicyProfileBuilder) WithSource(source string) *ExecutivePolicyProfileBuilder {
	b.profile.PolicySource = source
	return b
}

func (b *ExecutivePolicyProfileBuilder) WithPriorityPolicies(pp PriorityPolicies) *ExecutivePolicyProfileBuilder {
	b.profile.PriorityPolicies = pp
	return b
}

func (b *ExecutivePolicyProfileBuilder) WithBudgetPolicies(bp BudgetPolicies) *ExecutivePolicyProfileBuilder {
	b.profile.BudgetPolicies = bp
	return b
}

func (b *ExecutivePolicyProfileBuilder) WithRetryPolicies(rp RetryPolicies) *ExecutivePolicyProfileBuilder {
	b.profile.RetryPolicies = rp
	return b
}

func (b *ExecutivePolicyProfileBuilder) WithCancellationPolicies(cp CancellationPolicies) *ExecutivePolicyProfileBuilder {
	b.profile.CancellationPolicies = cp
	return b
}

func (b *ExecutivePolicyProfileBuilder) WithEscalationPolicies(ep EscalationPolicies) *ExecutivePolicyProfileBuilder {
	b.profile.EscalationPolicies = ep
	return b
}

func (b *ExecutivePolicyProfileBuilder) WithWorkspacePolicies(wp WorkspacePolicies) *ExecutivePolicyProfileBuilder {
	b.profile.WorkspacePolicies = wp
	return b
}

func (b *ExecutivePolicyProfileBuilder) WithHomeostasisPolicies(hp HomeostasisPolicies) *ExecutivePolicyProfileBuilder {
	b.profile.HomeostasisPolicies = hp
	return b
}

// Build generates the fingerprint, validates all policy rules, and returns the immutable profile.
func (b *ExecutivePolicyProfileBuilder) Build() (*ExecutivePolicyProfile, error) {
	b.profile.PolicyFingerprint = b.profile.GeneratePolicyFingerprint()
	if err := b.profile.Validate(); err != nil {
		return nil, err
	}
	return b.profile, nil
}

// ExecutiveCapabilitiesBuilder builds ExecutiveCapabilities with fluent validation and fingerprinting.
type ExecutiveCapabilitiesBuilder struct {
	caps *ExecutiveCapabilities
}

// NewExecutiveCapabilitiesBuilder initializes a capabilities builder with standard defaults.
func NewExecutiveCapabilitiesBuilder() *ExecutiveCapabilitiesBuilder {
	return &ExecutiveCapabilitiesBuilder{
		caps: &ExecutiveCapabilities{
			SupportsInterrupts:           true,
			SupportsPauseResume:          true,
			SupportsScheduling:           true,
			SupportsRecovery:             true,
			SupportsCheckpointing:        true,
			SupportsBackgroundTasks:      true,
			SupportsWorkspaceArbitration: true,
			SupportsCalibration:          true,
			SupportsConstitution:         true,
			MaxConcurrentEpisodes:        100,
			MaxRetryBudget:               1000,
		},
	}
}

func (b *ExecutiveCapabilitiesBuilder) WithInterrupts(supported bool) *ExecutiveCapabilitiesBuilder {
	b.caps.SupportsInterrupts = supported
	return b
}

func (b *ExecutiveCapabilitiesBuilder) WithPauseResume(supported bool) *ExecutiveCapabilitiesBuilder {
	b.caps.SupportsPauseResume = supported
	return b
}

func (b *ExecutiveCapabilitiesBuilder) WithScheduling(supported bool) *ExecutiveCapabilitiesBuilder {
	b.caps.SupportsScheduling = supported
	return b
}

func (b *ExecutiveCapabilitiesBuilder) WithRecovery(supported bool) *ExecutiveCapabilitiesBuilder {
	b.caps.SupportsRecovery = supported
	return b
}

func (b *ExecutiveCapabilitiesBuilder) WithCheckpointing(supported bool) *ExecutiveCapabilitiesBuilder {
	b.caps.SupportsCheckpointing = supported
	return b
}

func (b *ExecutiveCapabilitiesBuilder) WithBackgroundTasks(supported bool) *ExecutiveCapabilitiesBuilder {
	b.caps.SupportsBackgroundTasks = supported
	return b
}

func (b *ExecutiveCapabilitiesBuilder) WithWorkspaceArbitration(supported bool) *ExecutiveCapabilitiesBuilder {
	b.caps.SupportsWorkspaceArbitration = supported
	return b
}

func (b *ExecutiveCapabilitiesBuilder) WithCalibration(supported bool) *ExecutiveCapabilitiesBuilder {
	b.caps.SupportsCalibration = supported
	return b
}

func (b *ExecutiveCapabilitiesBuilder) WithConstitution(supported bool) *ExecutiveCapabilitiesBuilder {
	b.caps.SupportsConstitution = supported
	return b
}

func (b *ExecutiveCapabilitiesBuilder) WithMaxConcurrentEpisodes(max int) *ExecutiveCapabilitiesBuilder {
	b.caps.MaxConcurrentEpisodes = max
	return b
}

func (b *ExecutiveCapabilitiesBuilder) WithMaxRetryBudget(max int) *ExecutiveCapabilitiesBuilder {
	b.caps.MaxRetryBudget = max
	return b
}

// Build generates the fingerprint, validates limits, and returns the immutable capabilities.
func (b *ExecutiveCapabilitiesBuilder) Build() (*ExecutiveCapabilities, error) {
	b.caps.CapabilityFingerprint = b.caps.GenerateCapabilityFingerprint()
	if err := b.caps.Validate(); err != nil {
		return nil, err
	}
	return b.caps, nil
}

// ReplayMetadataBuilder builds ReplayMetadata with fluent validation.
type ReplayMetadataBuilder struct {
	meta *ReplayMetadata
}

// NewReplayMetadataBuilder initializes a new ReplayMetadataBuilder.
func NewReplayMetadataBuilder() *ReplayMetadataBuilder {
	return &ReplayMetadataBuilder{
		meta: &ReplayMetadata{
			ExecutiveVersion: DefaultExecutiveVersion,
			ReplayFidelity:   "EXACT",
			ReplayTimestamp:  time.Now().UTC(),
		},
	}
}

func (b *ReplayMetadataBuilder) WithPolicyFingerprint(fp string) *ReplayMetadataBuilder {
	b.meta.PolicyFingerprint = fp
	return b
}

func (b *ReplayMetadataBuilder) WithCapabilityFingerprint(fp string) *ReplayMetadataBuilder {
	b.meta.CapabilityFingerprint = fp
	return b
}

func (b *ReplayMetadataBuilder) WithReplaySeed(seed uint64) *ReplayMetadataBuilder {
	b.meta.ReplaySeed = seed
	return b
}

func (b *ReplayMetadataBuilder) WithExecutiveVersion(version string) *ReplayMetadataBuilder {
	b.meta.ExecutiveVersion = version
	return b
}

func (b *ReplayMetadataBuilder) WithConfigurationFingerprint(fp string) *ReplayMetadataBuilder {
	b.meta.ConfigurationFingerprint = fp
	return b
}

func (b *ReplayMetadataBuilder) WithFidelity(fidelity string) *ReplayMetadataBuilder {
	b.meta.ReplayFidelity = fidelity
	return b
}

// Build validates and returns the ReplayMetadata.
func (b *ReplayMetadataBuilder) Build() (*ReplayMetadata, error) {
	if err := b.meta.Validate(); err != nil {
		return nil, err
	}
	return b.meta, nil
}

// ExecutiveTraceBuilder builds an ExecutiveTrace with fluent validation.
type ExecutiveTraceBuilder struct {
	trace *ExecutiveTrace
}

// NewExecutiveTraceBuilder initializes a new ExecutiveTraceBuilder.
func NewExecutiveTraceBuilder() *ExecutiveTraceBuilder {
	return &ExecutiveTraceBuilder{
		trace: &ExecutiveTrace{
			SchemaVersion:    SchemaVersion,
			Timestamp:        time.Now().UTC(),
			ExecutiveVersion: DefaultExecutiveVersion,
		},
	}
}

func (b *ExecutiveTraceBuilder) WithTraceID(id string) *ExecutiveTraceBuilder {
	b.trace.TraceID = id
	return b
}

func (b *ExecutiveTraceBuilder) WithEpisodeID(id string) *ExecutiveTraceBuilder {
	b.trace.EpisodeID = id
	return b
}

func (b *ExecutiveTraceBuilder) WithBudgetsAllocated(events []BudgetAllocationEvent) *ExecutiveTraceBuilder {
	b.trace.BudgetsAllocated = events
	return b
}

func (b *ExecutiveTraceBuilder) WithPriorityChosen(events []PriorityChosenEvent) *ExecutiveTraceBuilder {
	b.trace.PriorityChosen = events
	return b
}

func (b *ExecutiveTraceBuilder) WithCancellationEvents(events []CancellationEvent) *ExecutiveTraceBuilder {
	b.trace.CancellationEvents = events
	return b
}

func (b *ExecutiveTraceBuilder) WithWorkspaceArbitration(events []WorkspaceArbitrationEvent) *ExecutiveTraceBuilder {
	b.trace.WorkspaceArbitration = events
	return b
}

func (b *ExecutiveTraceBuilder) WithConstitutionInvocation(events []ConstitutionInvocationEvent) *ExecutiveTraceBuilder {
	b.trace.ConstitutionInvocation = events
	return b
}

func (b *ExecutiveTraceBuilder) WithCalibrationInvocation(events []CalibrationInvocationEvent) *ExecutiveTraceBuilder {
	b.trace.CalibrationInvocation = events
	return b
}

func (b *ExecutiveTraceBuilder) WithExecutionDuration(d time.Duration) *ExecutiveTraceBuilder {
	b.trace.ExecutionDuration = d
	return b
}

func (b *ExecutiveTraceBuilder) WithProvenance(policyFP, capFP, execVersion string) *ExecutiveTraceBuilder {
	b.trace.PolicyFingerprint = policyFP
	b.trace.CapabilityFingerprint = capFP
	b.trace.ExecutiveVersion = execVersion
	return b
}

func (b *ExecutiveTraceBuilder) WithReplayMetadata(meta ReplayMetadata) *ExecutiveTraceBuilder {
	b.trace.ReplayMetadata = meta
	return b
}

// Build validates and returns the ExecutiveTrace.
func (b *ExecutiveTraceBuilder) Build() (*ExecutiveTrace, error) {
	if err := b.trace.Validate(); err != nil {
		return nil, err
	}
	return b.trace, nil
}

// ExecutiveResultBuilder builds an ExecutiveResult with fluent validation.
type ExecutiveResultBuilder struct {
	result *ExecutiveResult
}

// NewExecutiveResultBuilder initializes a new ExecutiveResultBuilder.
func NewExecutiveResultBuilder() *ExecutiveResultBuilder {
	return &ExecutiveResultBuilder{
		result: &ExecutiveResult{
			Status:            StatusSuccess,
			TerminationReason: ReasonSuccess,
			ExecutiveVersion:  DefaultExecutiveVersion,
		},
	}
}

func (b *ExecutiveResultBuilder) WithEpisodeID(id string) *ExecutiveResultBuilder {
	b.result.EpisodeID = id
	return b
}

func (b *ExecutiveResultBuilder) WithWorkflowID(id string) *ExecutiveResultBuilder {
	b.result.WorkflowID = id
	return b
}

func (b *ExecutiveResultBuilder) WithStatus(status ExecutiveResultStatus) *ExecutiveResultBuilder {
	b.result.Status = status
	return b
}

func (b *ExecutiveResultBuilder) WithTerminationReason(reason ExecutiveTerminationReason) *ExecutiveResultBuilder {
	b.result.TerminationReason = reason
	return b
}

func (b *ExecutiveResultBuilder) WithOutputRef(ref string) *ExecutiveResultBuilder {
	b.result.OutputRef = ref
	return b
}

func (b *ExecutiveResultBuilder) WithError(err error) *ExecutiveResultBuilder {
	b.result.Error = err
	if err != nil {
		b.result.ErrorString = err.Error()
	}
	return b
}

func (b *ExecutiveResultBuilder) WithCoordinationSummary(summary ExecutiveCoordinationSummary) *ExecutiveResultBuilder {
	b.result.CoordinationSummary = summary
	return b
}

func (b *ExecutiveResultBuilder) WithTerminationSummary(summary CoordinationTerminationSummary) *ExecutiveResultBuilder {
	b.result.TerminationSummary = summary
	b.result.CoordinationSummary.TerminationSummary = summary
	return b
}

func (b *ExecutiveResultBuilder) WithTraceID(id string) *ExecutiveResultBuilder {
	b.result.TraceID = id
	return b
}

func (b *ExecutiveResultBuilder) WithProvenance(policyFP, capFP, execVersion string) *ExecutiveResultBuilder {
	b.result.PolicyFingerprint = policyFP
	b.result.CapabilityFingerprint = capFP
	b.result.ExecutiveVersion = execVersion
	return b
}

func (b *ExecutiveResultBuilder) WithReplayMetadata(meta ReplayMetadata) *ExecutiveResultBuilder {
	b.result.ReplayMetadata = meta
	return b
}

// Build validates and returns the ExecutiveResult.
func (b *ExecutiveResultBuilder) Build() (*ExecutiveResult, error) {
	if err := b.result.Validate(); err != nil {
		return nil, err
	}
	return b.result, nil
}
