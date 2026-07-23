// Package understanding implements the IDUN V3 Understanding Subsystem
// Architecture Version 2.0.0-FROZEN.
//
// Understanding is responsible for perception interpretation, semantic normalization,
// slot/referent extraction, and bounded multi-hypothesis ambiguity representation.
package understanding

import (
	"errors"
	"fmt"
)

const (
	// FrameVersion defines the canonical Version 2.0 schema version string.
	FrameVersion = "2.0"

	// MaxBeamWidth defines the maximum total hypotheses preserved in a SemanticFrame
	// (1 primary hypothesis + up to 2 runner-up hypotheses in AmbiguitySet).
	MaxBeamWidth = 3

	// DefaultAmbiguityDelta defines the default maximum difference in calibrated
	// confidence (P_eff) between the primary hypothesis and an ambiguous runner-up.
	DefaultAmbiguityDelta = 0.15
)

// Sentinel errors returned by validation and builder methods.
var (
	ErrInvalidFrameVersion = errors.New("understanding: invalid frame version")
	ErrMissingEnvelopeID   = errors.New("understanding: envelope ID is required")
	ErrMissingIntent       = errors.New("understanding: hypothesis intent is required")
	ErrInvalidConfidence   = errors.New("understanding: confidence must be within [0.0, 1.0]")
	ErrInvalidDelta        = errors.New("understanding: hypothesis delta must be non-negative and <= 1.0")
	ErrBeamOverflow        = errors.New("understanding: ambiguity beam exceeds MaxBeamWidth limit")
	ErrInvalidSlotName     = errors.New("understanding: slot name is required")
	ErrServiceClosed       = errors.New("understanding: service is closed")
	ErrNilPerception       = errors.New("understanding: perception envelope cannot be nil")
)

// SourceLayer identifies which internal evaluation specialist produced a hypothesis.
type SourceLayer string

const (
	// LayerReflexiveGrammar identifies deterministic grammar and structured pattern rules (<5ms).
	LayerReflexiveGrammar SourceLayer = "Understanding.ReflexiveGrammar"

	// LayerNeuralClassifier identifies local quantized neural classifiers (<20ms).
	LayerNeuralClassifier SourceLayer = "Understanding.NeuralClassifier"

	// LayerDeliberativeLLM identifies asynchronous deep LLM semantic parsing.
	LayerDeliberativeLLM SourceLayer = "Understanding.DeliberativeLLM"
)

// IsValid checks whether the SourceLayer is a recognized canonical layer.
func (l SourceLayer) IsValid() bool {
	switch l {
	case LayerReflexiveGrammar, LayerNeuralClassifier, LayerDeliberativeLLM:
		return true
	default:
		return false
	}
}

// InterpretationStatus summarizes the outcome state of perceptual interpretation.
type InterpretationStatus string

const (
	// StatusUnambiguous indicates a single high-confidence hypothesis exceeded threshold tau.
	StatusUnambiguous InterpretationStatus = "UNAMBIGUOUS"

	// StatusAmbiguousBeam indicates competing near-equal hypotheses were preserved in AmbiguitySet.
	StatusAmbiguousBeam InterpretationStatus = "AMBIGUOUS_BEAM"

	// StatusPreliminary indicates a fast reflexive estimate published while awaiting deep LLM parsing.
	StatusPreliminary InterpretationStatus = "PRELIMINARY"

	// StatusFailedImpasse indicates no hypothesis crossed admission threshold tau.
	StatusFailedImpasse InterpretationStatus = "FAILED_IMPASSE"
)

// IsValid checks whether the InterpretationStatus is recognized.
func (s InterpretationStatus) IsValid() bool {
	switch s {
	case StatusUnambiguous, StatusAmbiguousBeam, StatusPreliminary, StatusFailedImpasse:
		return true
	default:
		return false
	}
}

// Slot represents an extracted semantic argument or referent binding.
type Slot struct {
	// Name identifies the semantic slot or parameter (e.g., "destination", "target_event").
	Name string `json:"Name"`

	// Value records the extracted text or value representation.
	Value string `json:"Value"`

	// GroundingID references the canonical entity ID resolved from working memory or ontology.
	GroundingID string `json:"GroundingID"`

	// Confidence records the calibrated slot extraction confidence [0.0, 1.0].
	Confidence float64 `json:"Confidence"`
}

// Validate verifies the structural validity of a Slot.
func (s Slot) Validate() error {
	if s.Name == "" {
		return ErrInvalidSlotName
	}
	if s.Confidence < 0.0 || s.Confidence > 1.0 {
		return fmt.Errorf("%w: slot %q confidence %f", ErrInvalidConfidence, s.Name, s.Confidence)
	}
	return nil
}

// Hypothesis represents a single semantic interpretation of an utterance or event.
type Hypothesis struct {
	// Intent identifies the categorized semantic intent (e.g., "book_flight").
	Intent string `json:"Intent"`

	// CalibratedConfidence records the P_eff confidence after Epistemic Calibration [0.0, 1.0].
	CalibratedConfidence float64 `json:"CalibratedConfidence"`

	// SourceLayer records which specialist layer generated this interpretation.
	SourceLayer SourceLayer `json:"SourceLayer"`

	// Slots contains extracted semantic arguments and grounded referents.
	Slots []Slot `json:"Slots"`

	// DeltaFromPrimary records P_eff(primary) - P_eff(this) for items in AmbiguitySet.
	DeltaFromPrimary float64 `json:"DeltaFromPrimary,omitempty"`
}

// Validate checks structural and range invariants for a Hypothesis.
func (h Hypothesis) Validate() error {
	if h.Intent == "" {
		return ErrMissingIntent
	}
	if h.CalibratedConfidence < 0.0 || h.CalibratedConfidence > 1.0 {
		return fmt.Errorf("%w: hypothesis %q confidence %f", ErrInvalidConfidence, h.Intent, h.CalibratedConfidence)
	}
	if h.DeltaFromPrimary < 0.0 || h.DeltaFromPrimary > 1.0 {
		return fmt.Errorf("%w: hypothesis %q delta %f", ErrInvalidDelta, h.Intent, h.DeltaFromPrimary)
	}
	for i, slot := range h.Slots {
		if err := slot.Validate(); err != nil {
			return fmt.Errorf("slot[%d]: %w", i, err)
		}
	}
	return nil
}

// Clone returns a deep copy of the Hypothesis.
func (h Hypothesis) Clone() Hypothesis {
	out := h
	if len(h.Slots) > 0 {
		out.Slots = make([]Slot, len(h.Slots))
		copy(out.Slots, h.Slots)
	} else {
		out.Slots = []Slot{}
	}
	return out
}

// EntityIdentity provides a stable semantic identity for recognized concepts.
type EntityIdentity struct {
	EntityID      string   `json:"EntityID"`
	CanonicalType string   `json:"CanonicalType"`
	CanonicalName string   `json:"CanonicalName"`
	Aliases       []string `json:"Aliases"`
	Confidence    float64  `json:"Confidence"`
	Lineage       string   `json:"Lineage"`
}

// ValidationTrace records the output of the internal Semantic Firewall.
type ValidationTrace struct {
	TypeValidation         string `json:"TypeValidation"`
	ConstraintValidation   string `json:"ConstraintValidation"`
	RelationshipValidation string `json:"RelationshipValidation"`
	TemporalValidation     string `json:"TemporalValidation"`
}

// Ambiguity represents a detected multi-path interpretation.
type Ambiguity struct {
	AmbiguousSpan        string   `json:"AmbiguousSpan"`
	CandidateResolutions []string `json:"CandidateResolutions"`
	Severity             string   `json:"Severity"` // e.g. "FATAL", "WARNING"
}

// Assumption represents a classified inference made by Understanding.
type Assumption struct {
	Category string `json:"Category"` // e.g. "Safe", "Risky", "Preferred"
	Details  string `json:"Details"`
}

// IntentNode represents a node in the Intent DAG.
type IntentNode struct {
	IntentID     string   `json:"IntentID"`
	Intent       string   `json:"Intent"`
	Dependencies []string `json:"Dependencies"` // IDs of prerequisite intents
}

// SemanticFrame is the canonical, version-invariant output contract published
// by Understanding to the Global Workspace (TopicPerceivedIntent).
type SemanticFrame struct {
	// FrameVersion specifies the canonical schema version ("2.0").
	FrameVersion string `json:"FrameVersion"`

	// EnvelopeID correlates this frame with the incoming perception Envelope ID.
	EnvelopeID string `json:"EnvelopeID"`

	// Status records the interpretation status.
	Status InterpretationStatus `json:"Status"`

	// PrimaryHypothesis holds the highest calibrated P_eff interpretation.
	PrimaryHypothesis Hypothesis `json:"PrimaryHypothesis"`

	// AmbiguitySet holds bounded runner-up hypotheses within Delta threshold (MaxBeamWidth <= 3 total).
	// Always non-nil ([]Hypothesis{}) to guarantee uniform JSON serialization.
	AmbiguitySet []Hypothesis `json:"AmbiguitySet"`

	// TopDownPriorApplied records any top-down dialogue or attentional prior applied.
	TopDownPriorApplied string `json:"TopDownPriorApplied,omitempty"`

	// ProcessedDurationMs records total processing latency in milliseconds.
	ProcessedDurationMs float64 `json:"ProcessedDurationMs"`

	// --- Phase 2B Semantic Interpretation Fields ---

	// Intents represents the Intent DAG.
	Intents []IntentNode `json:"Intents,omitempty"`

	// Entities contains strongly-typed ontological objects.
	Entities map[string]EntityIdentity `json:"Entities,omitempty"`

	// Constraints contains hard boundaries extracted from text.
	Constraints []string `json:"Constraints,omitempty"`

	// ValidationTrace contains detailed structural validation results.
	ValidationTrace *ValidationTrace `json:"ValidationTrace,omitempty"`

	// Ambiguities contains detected multi-path interpretations.
	Ambiguities []Ambiguity `json:"Ambiguities,omitempty"`

	// Assumptions contains inferences made by Understanding.
	Assumptions []Assumption `json:"Assumptions,omitempty"`

	// MissingInformation contains known required slots that were absent.
	MissingInformation []string `json:"MissingInformation,omitempty"`

	// Completeness records the semantic completeness score [0.0, 1.0].
	Completeness float64 `json:"Completeness,omitempty"`

	// ConfidenceTrace records modifiers applied to the base confidence.
	ConfidenceTrace []string `json:"ConfidenceTrace,omitempty"`
}

// Validate checks whether the SemanticFrame satisfies all structural and beam invariants.
func (f SemanticFrame) Validate() error {
	if f.FrameVersion != FrameVersion {
		return fmt.Errorf("%w: expected %q, got %q", ErrInvalidFrameVersion, FrameVersion, f.FrameVersion)
	}
	if f.EnvelopeID == "" {
		return ErrMissingEnvelopeID
	}
	if !f.Status.IsValid() {
		return fmt.Errorf("understanding: invalid status %q", f.Status)
	}
	if err := f.PrimaryHypothesis.Validate(); err != nil {
		return fmt.Errorf("primary hypothesis: %w", err)
	}
	if len(f.AmbiguitySet)+1 > MaxBeamWidth {
		return fmt.Errorf("%w: total hypotheses %d exceeds MaxBeamWidth %d", ErrBeamOverflow, len(f.AmbiguitySet)+1, MaxBeamWidth)
	}
	for i, hyp := range f.AmbiguitySet {
		if err := hyp.Validate(); err != nil {
			return fmt.Errorf("ambiguity_set[%d]: %w", i, err)
		}
	}
	return nil
}

// Clone returns a deep, immutable copy of the SemanticFrame.
func (f SemanticFrame) Clone() SemanticFrame {
	out := f
	out.PrimaryHypothesis = f.PrimaryHypothesis.Clone()
	if len(f.AmbiguitySet) > 0 {
		out.AmbiguitySet = make([]Hypothesis, len(f.AmbiguitySet))
		for i, h := range f.AmbiguitySet {
			out.AmbiguitySet[i] = h.Clone()
		}
	} else {
		out.AmbiguitySet = []Hypothesis{}
	}
	return out
}
