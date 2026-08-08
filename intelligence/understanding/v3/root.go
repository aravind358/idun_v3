package v3

import (
	"idun/core/foundation"
)

// SemanticInterpretation represents the immutable output contract of the
// IDUN V3 Understanding Subsystem.
type SemanticInterpretation struct {
	specVersion          string
	artifactID           foundation.ArtifactID
	parentArtifactID     foundation.ParentArtifactID
	envelopeID           foundation.EnvelopeID
	timestamp            foundation.Timestamp
	status               InterpretationStatus
	primaryIntent        string
	communicativeAct     CommunicativeAct
	primaryHypothesis    Hypothesis
	ambiguitySet         []Hypothesis
	topics               []string
	entities             []Entity
	references           []Reference
	temporalAnchors      []TemporalAnchor
	composedTimestamps   []string
	openSlots            []string
	isConditional        bool
	conditionClause      string
	consequentClause     string
	compoundIntentCount  int
	secondaryIntents     []SecondaryIntent
	requiresContext      bool
	polarity             Polarity
	sentiment            Sentiment
	inputCharacteristics InputCharacteristics
	dialoguePosition     DialoguePosition
	executionHints       ExecutionHints
	assumptions          []Assumption
	ambiguities          []Ambiguity
	confidence           float64
	metadata             foundation.InteractionMetadata
}

// IsImmutable satisfies foundation.Immutable.
func (s *SemanticInterpretation) IsImmutable() bool { return true }

func (s *SemanticInterpretation) ArtifactID() foundation.ArtifactID         { return s.artifactID }
func (s *SemanticInterpretation) ParentArtifactID() foundation.ParentArtifactID { return s.parentArtifactID }
func (s *SemanticInterpretation) Timestamp() foundation.Timestamp           { return s.timestamp }
func (s *SemanticInterpretation) SpecVersion() foundation.Version           { return foundation.Version(s.specVersion) }
func (s *SemanticInterpretation) EnvelopeID() foundation.EnvelopeID         { return s.envelopeID }
func (s *SemanticInterpretation) Status() InterpretationStatus              { return s.status }
func (s *SemanticInterpretation) PrimaryIntent() string                     { return s.primaryIntent }
func (s *SemanticInterpretation) CommunicativeAct() CommunicativeAct        { return s.communicativeAct }
func (s *SemanticInterpretation) PrimaryHypothesis() Hypothesis             { return s.primaryHypothesis }
func (s *SemanticInterpretation) AmbiguitySet() []Hypothesis                { a := make([]Hypothesis, len(s.ambiguitySet)); copy(a, s.ambiguitySet); return a }
func (s *SemanticInterpretation) Topics() []string                          { t := make([]string, len(s.topics)); copy(t, s.topics); return t }
func (s *SemanticInterpretation) Entities() []Entity                        { e := make([]Entity, len(s.entities)); copy(e, s.entities); return e }
func (s *SemanticInterpretation) References() []Reference                   { r := make([]Reference, len(s.references)); copy(r, s.references); return r }
func (s *SemanticInterpretation) TemporalAnchors() []TemporalAnchor         { t := make([]TemporalAnchor, len(s.temporalAnchors)); copy(t, s.temporalAnchors); return t }
func (s *SemanticInterpretation) ComposedTimestamps() []string              { c := make([]string, len(s.composedTimestamps)); copy(c, s.composedTimestamps); return c }
func (s *SemanticInterpretation) OpenSlots() []string                       { o := make([]string, len(s.openSlots)); copy(o, s.openSlots); return o }
func (s *SemanticInterpretation) IsConditional() bool                       { return s.isConditional }
func (s *SemanticInterpretation) ConditionClause() string                   { return s.conditionClause }
func (s *SemanticInterpretation) ConsequentClause() string                  { return s.consequentClause }
func (s *SemanticInterpretation) CompoundIntentCount() int                  { return s.compoundIntentCount }
func (s *SemanticInterpretation) SecondaryIntents() []SecondaryIntent       { si := make([]SecondaryIntent, len(s.secondaryIntents)); copy(si, s.secondaryIntents); return si }
func (s *SemanticInterpretation) RequiresContext() bool                     { return s.requiresContext }
func (s *SemanticInterpretation) Polarity() Polarity                        { return s.polarity }
func (s *SemanticInterpretation) Sentiment() Sentiment                      { return s.sentiment }
func (s *SemanticInterpretation) InputCharacteristics() InputCharacteristics { return s.inputCharacteristics }
func (s *SemanticInterpretation) DialoguePosition() DialoguePosition        { return s.dialoguePosition }
func (s *SemanticInterpretation) ExecutionHints() ExecutionHints            { return s.executionHints }
func (s *SemanticInterpretation) Assumptions() []Assumption                 { a := make([]Assumption, len(s.assumptions)); copy(a, s.assumptions); return a }
func (s *SemanticInterpretation) Ambiguities() []Ambiguity                  { a := make([]Ambiguity, len(s.ambiguities)); copy(a, s.ambiguities); return a }
func (s *SemanticInterpretation) Confidence() float64                       { return s.confidence }
func (s *SemanticInterpretation) Metadata() foundation.InteractionMetadata  { return s.metadata }
