package reasoning

import (
	"errors"
	"fmt"
)

// SchemaVersion defines the canonical invariant schema version ("2.0").
const SchemaVersion = "2.0"

// MaxBeamWidth limits the maximum number of retained reasoning hypotheses (Primary + up to 2 runners-up).
const MaxBeamWidth = 3

var (
	ErrInvalidSchemaVersion = errors.New("reasoning: invalid schema version")
	ErrMissingEnvelopeID    = errors.New("reasoning: missing envelope ID")
	ErrMissingSourceFrameID = errors.New("reasoning: missing source frame ID")
	ErrMissingConclusion    = errors.New("reasoning: missing hypothesis conclusion")
	ErrInvalidConfidence    = errors.New("reasoning: confidence value out of bounds [0.0, 1.0]")
	ErrBeamOverflow         = errors.New("reasoning: hypothesis beam exceeds MaxBeamWidth")
	ErrServiceClosed        = errors.New("reasoning: service closed")
	ErrReasoningCancelled   = errors.New("reasoning: context cancelled")
	ErrReasoningTimeout     = errors.New("reasoning: operation timed out")
	ErrInvalidGraphLimits   = errors.New("reasoning: invalid graph resource limits")
)

// ReasoningOutcomeStatus summarizes the outcome state of a reasoning episode.
type ReasoningOutcomeStatus string

const (
	StatusUnambiguousSolved ReasoningOutcomeStatus = "UNAMBIGUOUS_SOLVED"
	StatusAmbiguousBeam     ReasoningOutcomeStatus = "AMBIGUOUS_BEAM"
	StatusPreliminary       ReasoningOutcomeStatus = "PRELIMINARY"
	StatusFailedImpasse     ReasoningOutcomeStatus = "FAILED_IMPASSE"
)

// IsValid checks whether ReasoningOutcomeStatus is a known enum value.
func (s ReasoningOutcomeStatus) IsValid() bool {
	switch s {
	case StatusUnambiguousSolved, StatusAmbiguousBeam, StatusPreliminary, StatusFailedImpasse:
		return true
	default:
		return false
	}
}

// HypothesisType classifies the logical or structural nature of a derived conclusion.
type HypothesisType string

const (
	HypothesisInference     HypothesisType = "INFERENCE"
	HypothesisRelation      HypothesisType = "RELATION"
	HypothesisAnalogy       HypothesisType = "ANALOGY"
	HypothesisContradiction HypothesisType = "CONTRADICTION"
	HypothesisAbduction     HypothesisType = "ABDUCTION"
	HypothesisDeliberative  HypothesisType = "DELIBERATIVE"
)

// IsValid checks whether HypothesisType is a known enum value.
func (h HypothesisType) IsValid() bool {
	switch h {
	case HypothesisInference, HypothesisRelation, HypothesisAnalogy, HypothesisContradiction, HypothesisAbduction, HypothesisDeliberative:
		return true
	default:
		return false
	}
}

// StageIdentifier identifies a specific execution stage in the 11-stage Reasoning Cascade.
type StageIdentifier string

const (
	StageS0ContextAssembly  StageIdentifier = "S0_CONTEXT_ASSEMBLY"
	StageS1SymbolicFast     StageIdentifier = "S1_SYMBOLIC_FAST"
	StageS2RelationalGraph  StageIdentifier = "S2_RELATIONAL_GRAPH"
	StageS3CSPCheck         StageIdentifier = "S3_CSP_CHECK"
	StageS4EvidenceFusion   StageIdentifier = "S4_EVIDENCE_FUSION"
	StageS5CaseAnalogy      StageIdentifier = "S5_CASE_ANALOGY"
	StageS6BeamSelection    StageIdentifier = "S6_BEAM_SELECTION"
	StageS7Calibration      StageIdentifier = "S7_CALIBRATION"
	StageS8DeliberativeLLM  StageIdentifier = "S8_DELIBERATIVE_LLM"
	StageS9Constitution     StageIdentifier = "S9_CONSTITUTION"
	StageS10ResultAssembly  StageIdentifier = "S10_RESULT_ASSEMBLY"
)

// StrategyIdentifier identifies a dynamic reasoning routing policy.
type StrategyIdentifier string

const (
	StrategySymbolicFast          StrategyIdentifier = "STRATEGY_SYMBOLIC_FAST"
	StrategyAnalogicalBayes       StrategyIdentifier = "STRATEGY_ANALOGICAL_BAYES"
	StrategyGraphDeliberative     StrategyIdentifier = "STRATEGY_GRAPH_DELIBERATIVE"
	StrategyDeliberativeEscalate  StrategyIdentifier = "STRATEGY_DELIBERATIVE_ESCALATE"
	StrategyDynamicHybrid         StrategyIdentifier = "STRATEGY_DYNAMIC_HYBRID"
)

// CompilationCondition specifies an antecedent condition for learning compilation.
type CompilationCondition struct {
	Slot     string `json:"slot"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

// Validate checks whether CompilationCondition is well-formed.
func (c CompilationCondition) Validate() error {
	if c.Slot == "" || c.Operator == "" {
		return errors.New("reasoning: compilation condition missing slot or operator")
	}
	return nil
}

// CompilationConsequent specifies the action or conclusion derived by a compiled rule.
type CompilationConsequent struct {
	Intent     string `json:"intent"`
	RuleAction string `json:"rule_action"`
}

// Validate checks whether CompilationConsequent is well-formed.
func (c CompilationConsequent) Validate() error {
	if c.Intent == "" {
		return errors.New("reasoning: compilation consequent missing intent")
	}
	return nil
}

// CompilationCandidate holds a structured machine-readable representation that Learning
// can convert into fast symbolic production rules without parsing free-form prose.
type CompilationCandidate struct {
	EligibleForLearning  bool                   `json:"eligible_for_learning"`
	SourceStage          StageIdentifier        `json:"source_stage"`
	StrategyUsed         StrategyIdentifier     `json:"strategy_used"`
	CalibratedConfidence float64                `json:"calibrated_confidence"`
	Antecedents          []CompilationCondition `json:"antecedents"`
	DerivedConsequent    CompilationConsequent  `json:"derived_consequent"`
}

// Validate checks structural and confidence invariants for CompilationCandidate.
func (c *CompilationCandidate) Validate() error {
	if c == nil {
		return nil
	}
	if c.CalibratedConfidence < 0.0 || c.CalibratedConfidence > 1.0 {
		return fmt.Errorf("%w: compilation candidate confidence %f", ErrInvalidConfidence, c.CalibratedConfidence)
	}
	for i, ant := range c.Antecedents {
		if err := ant.Validate(); err != nil {
			return fmt.Errorf("antecedent[%d]: %w", i, err)
		}
	}
	if err := c.DerivedConsequent.Validate(); err != nil {
		return fmt.Errorf("consequent: %w", err)
	}
	return nil
}

// Clone returns a deep copy of CompilationCandidate.
func (c *CompilationCandidate) Clone() *CompilationCandidate {
	if c == nil {
		return nil
	}
	out := *c
	if len(c.Antecedents) > 0 {
		out.Antecedents = make([]CompilationCondition, len(c.Antecedents))
		copy(out.Antecedents, c.Antecedents)
	} else {
		out.Antecedents = []CompilationCondition{}
	}
	return &out
}

// ContradictionFlag records an internal logical conflict flagged against existing beliefs.
type ContradictionFlag struct {
	BeliefID          string          `json:"belief_id"`
	ConflictingClaim  string          `json:"conflicting_claim"`
	Confidence        float64         `json:"confidence"`
	DetectedAtStage   StageIdentifier `json:"detected_at_stage"`
}

// Validate verifies ContradictionFlag fields and confidence bounds.
func (cf ContradictionFlag) Validate() error {
	if cf.BeliefID == "" || cf.ConflictingClaim == "" {
		return errors.New("reasoning: contradiction flag missing belief ID or claim")
	}
	if cf.Confidence < 0.0 || cf.Confidence > 1.0 {
		return fmt.Errorf("%w: contradiction confidence %f", ErrInvalidConfidence, cf.Confidence)
	}
	return nil
}

// BeliefUpdateProposal suggests a non-binding update to Memory's belief store.
type BeliefUpdateProposal struct {
	Subject            string  `json:"subject"`
	Predicate          string  `json:"predicate"`
	Object             string  `json:"object"`
	ProposedConfidence float64 `json:"proposed_confidence"`
	SourceHypothesisID string  `json:"source_hypothesis_id"`
}

// Validate checks whether BeliefUpdateProposal is well-formed.
func (b BeliefUpdateProposal) Validate() error {
	if b.Subject == "" || b.Predicate == "" || b.Object == "" {
		return errors.New("reasoning: belief update proposal missing subject/predicate/object")
	}
	if b.ProposedConfidence < 0.0 || b.ProposedConfidence > 1.0 {
		return fmt.Errorf("%w: belief update confidence %f", ErrInvalidConfidence, b.ProposedConfidence)
	}
	return nil
}

// StageTraceLog records diagnostic trace events for an executed cascade stage.
type StageTraceLog struct {
	Stage               StageIdentifier `json:"stage"`
	ExecutionDurationMs float64         `json:"execution_duration_ms"`
	Description         string          `json:"description"`
	OutputSummary       string          `json:"output_summary"`
}

// StrategyTelemetry captures execution metadata allowing Learning to optimize strategy selection.
type StrategyTelemetry struct {
	EpisodeID           string                 `json:"episode_id"`
	StrategySelected    StrategyIdentifier     `json:"strategy_selected"`
	SpecialistsExecuted []StageIdentifier      `json:"specialists_executed"`
	ExecutionDurationMs float64                `json:"execution_duration_ms"`
	CalibratedConfidence float64               `json:"calibrated_confidence"`
	ResourceCostTier    string                 `json:"resource_cost_tier"`
	EscalatedToLLM      bool                   `json:"escalated_to_llm"`
	OutcomeStatus       ReasoningOutcomeStatus `json:"outcome_status"`
	Metadata            map[string]string      `json:"metadata,omitempty"`
}

// Validate checks confidence bounds and required identifiers in StrategyTelemetry.
func (st StrategyTelemetry) Validate() error {
	if st.EpisodeID == "" {
		return errors.New("reasoning: telemetry missing episode ID")
	}
	if st.CalibratedConfidence < 0.0 || st.CalibratedConfidence > 1.0 {
		return fmt.Errorf("%w: telemetry confidence %f", ErrInvalidConfidence, st.CalibratedConfidence)
	}
	return nil
}

// Clone returns a deep copy of StrategyTelemetry.
func (st StrategyTelemetry) Clone() StrategyTelemetry {
	out := st
	if len(st.SpecialistsExecuted) > 0 {
		out.SpecialistsExecuted = make([]StageIdentifier, len(st.SpecialistsExecuted))
		copy(out.SpecialistsExecuted, st.SpecialistsExecuted)
	} else {
		out.SpecialistsExecuted = []StageIdentifier{}
	}
	if st.Metadata != nil {
		out.Metadata = make(map[string]string, len(st.Metadata))
		for k, v := range st.Metadata {
			out.Metadata[k] = v
		}
	} else {
		out.Metadata = map[string]string{}
	}
	return out
}

// ReasoningHypothesis represents a single derived conclusion produced by the Reasoning Cascade.
type ReasoningHypothesis struct {
	ID                   string            `json:"id"`
	Type                 HypothesisType    `json:"type"`
	Conclusion           string            `json:"conclusion"`
	ReasoningConfidence  float64           `json:"reasoning_confidence,omitempty"`
	CalibratedConfidence float64           `json:"calibrated_confidence"`
	ContributingStages   []StageIdentifier `json:"contributing_stages"`
	SupportingPremises   []string          `json:"supporting_premises"`
	EvidenceTrace        string            `json:"evidence_trace"`
}

// Validate checks structural and range invariants for ReasoningHypothesis.
func (h ReasoningHypothesis) Validate() error {
	if h.ID == "" {
		return errors.New("reasoning: hypothesis missing ID")
	}
	if h.Conclusion == "" {
		return ErrMissingConclusion
	}
	if !h.Type.IsValid() {
		return fmt.Errorf("reasoning: invalid hypothesis type %q", h.Type)
	}
	if h.CalibratedConfidence < 0.0 || h.CalibratedConfidence > 1.0 {
		return fmt.Errorf("%w: hypothesis %q confidence %f", ErrInvalidConfidence, h.ID, h.CalibratedConfidence)
	}
	return nil
}

// Clone returns a deep copy of ReasoningHypothesis.
func (h ReasoningHypothesis) Clone() ReasoningHypothesis {
	out := h
	if len(h.ContributingStages) > 0 {
		out.ContributingStages = make([]StageIdentifier, len(h.ContributingStages))
		copy(out.ContributingStages, h.ContributingStages)
	} else {
		out.ContributingStages = []StageIdentifier{}
	}
	if len(h.SupportingPremises) > 0 {
		out.SupportingPremises = make([]string, len(h.SupportingPremises))
		copy(out.SupportingPremises, h.SupportingPremises)
	} else {
		out.SupportingPremises = []string{}
	}
	return out
}

// ReasoningResult is the canonical, version-invariant output contract published
// by Reasoning to Global Workspace. Downstream modules (Planning, Decision, Reflection)
// consume this structured envelope without parsing natural language prose.
type ReasoningResult struct {
	SchemaVersion           string                 `json:"schema_version"`
	EnvelopeID              string                 `json:"envelope_id"`
	SourceFrameID           string                 `json:"source_frame_id"`
	Status                  ReasoningOutcomeStatus `json:"status"`
	StrategyUsed            StrategyIdentifier     `json:"strategy_used"`
	PrimaryHypothesis       ReasoningHypothesis    `json:"primary_hypothesis"`
	AmbiguitySet            []ReasoningHypothesis  `json:"ambiguity_set"`
	ContradictionsFlagged   []ContradictionFlag    `json:"contradictions_flagged"`
	ProposedBeliefUpdates   []BeliefUpdateProposal `json:"proposed_belief_updates"`
	CompilationCandidate    *CompilationCandidate  `json:"compilation_candidate,omitempty"`
	StrategyTelemetry       StrategyTelemetry      `json:"strategy_telemetry"`
	ConstitutionAnnotations []string               `json:"constitution_annotations"`
	ReasoningTrace          []StageTraceLog        `json:"reasoning_trace"`
	OfflineMode             bool                   `json:"offline_mode"`
	ProcessedDurationMs     float64                `json:"processed_duration_ms"`
}

// Validate checks whether ReasoningResult satisfies all structural and beam invariants.
func (r ReasoningResult) Validate() error {
	if r.SchemaVersion != SchemaVersion {
		return fmt.Errorf("%w: expected %q, got %q", ErrInvalidSchemaVersion, SchemaVersion, r.SchemaVersion)
	}
	if r.EnvelopeID == "" {
		return ErrMissingEnvelopeID
	}
	if r.SourceFrameID == "" {
		return ErrMissingSourceFrameID
	}
	if !r.Status.IsValid() {
		return fmt.Errorf("reasoning: invalid status %q", r.Status)
	}
	if err := r.PrimaryHypothesis.Validate(); err != nil {
		return fmt.Errorf("primary hypothesis: %w", err)
	}
	if len(r.AmbiguitySet)+1 > MaxBeamWidth {
		return fmt.Errorf("%w: total hypotheses %d exceeds MaxBeamWidth %d", ErrBeamOverflow, len(r.AmbiguitySet)+1, MaxBeamWidth)
	}
	for i, hyp := range r.AmbiguitySet {
		if err := hyp.Validate(); err != nil {
			return fmt.Errorf("ambiguity_set[%d]: %w", i, err)
		}
	}
	for i, cf := range r.ContradictionsFlagged {
		if err := cf.Validate(); err != nil {
			return fmt.Errorf("contradictions_flagged[%d]: %w", i, err)
		}
	}
	for i, bu := range r.ProposedBeliefUpdates {
		if err := bu.Validate(); err != nil {
			return fmt.Errorf("proposed_belief_updates[%d]: %w", i, err)
		}
	}
	if err := r.CompilationCandidate.Validate(); err != nil {
		return fmt.Errorf("compilation_candidate: %w", err)
	}
	if err := r.StrategyTelemetry.Validate(); err != nil {
		return fmt.Errorf("strategy_telemetry: %w", err)
	}
	return nil
}

// Clone returns a deep copy of ReasoningResult.
func (r ReasoningResult) Clone() ReasoningResult {
	out := r
	out.PrimaryHypothesis = r.PrimaryHypothesis.Clone()

	if len(r.AmbiguitySet) > 0 {
		out.AmbiguitySet = make([]ReasoningHypothesis, len(r.AmbiguitySet))
		for i, h := range r.AmbiguitySet {
			out.AmbiguitySet[i] = h.Clone()
		}
	} else {
		out.AmbiguitySet = []ReasoningHypothesis{}
	}

	if len(r.ContradictionsFlagged) > 0 {
		out.ContradictionsFlagged = make([]ContradictionFlag, len(r.ContradictionsFlagged))
		copy(out.ContradictionsFlagged, r.ContradictionsFlagged)
	} else {
		out.ContradictionsFlagged = []ContradictionFlag{}
	}

	if len(r.ProposedBeliefUpdates) > 0 {
		out.ProposedBeliefUpdates = make([]BeliefUpdateProposal, len(r.ProposedBeliefUpdates))
		copy(out.ProposedBeliefUpdates, r.ProposedBeliefUpdates)
	} else {
		out.ProposedBeliefUpdates = []BeliefUpdateProposal{}
	}

	out.CompilationCandidate = r.CompilationCandidate.Clone()
	out.StrategyTelemetry = r.StrategyTelemetry.Clone()

	if len(r.ConstitutionAnnotations) > 0 {
		out.ConstitutionAnnotations = make([]string, len(r.ConstitutionAnnotations))
		copy(out.ConstitutionAnnotations, r.ConstitutionAnnotations)
	} else {
		out.ConstitutionAnnotations = []string{}
	}

	if len(r.ReasoningTrace) > 0 {
		out.ReasoningTrace = make([]StageTraceLog, len(r.ReasoningTrace))
		copy(out.ReasoningTrace, r.ReasoningTrace)
	} else {
		out.ReasoningTrace = []StageTraceLog{}
	}

	return out
}
