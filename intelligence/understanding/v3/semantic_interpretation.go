package v3

import (
	"errors"
)

const (
	SpecVersion = "3.0"
)

// InterpretationStatus
type InterpretationStatus string

const (
	StatusUnambiguous  InterpretationStatus = "UNAMBIGUOUS"
	StatusAmbiguous    InterpretationStatus = "AMBIGUOUS_BEAM"
	StatusPreliminary  InterpretationStatus = "PRELIMINARY"
	StatusFailed       InterpretationStatus = "FAILED_IMPASSE"
)

// CommunicativeAct
type CommunicativeAct string

const (
	ActQuestion      CommunicativeAct = "QUESTION"
	ActCommand       CommunicativeAct = "COMMAND"
	ActRequest       CommunicativeAct = "REQUEST"
	ActStatement     CommunicativeAct = "STATEMENT"
	ActDeclaration   CommunicativeAct = "DECLARATION"
	ActCorrection    CommunicativeAct = "CORRECTION"
	ActConfirmation  CommunicativeAct = "CONFIRMATION"
	ActRefusal       CommunicativeAct = "REFUSAL"
	ActClarification CommunicativeAct = "CLARIFICATION"
	ActExpression    CommunicativeAct = "EXPRESSION"
	ActConversation  CommunicativeAct = "CONVERSATION"
)

// SourceLayer
type SourceLayer string

const (
	LayerReflexiveGrammar SourceLayer = "Understanding.ReflexiveGrammar"
	LayerNeuralClassifier SourceLayer = "Understanding.NeuralClassifier"
	LayerDeliberativeLLM  SourceLayer = "Understanding.DeliberativeLLM"
)

// EntityType
type EntityType string

const (
	EntityPerson           EntityType = "PERSON"
	EntityOrganization     EntityType = "ORGANIZATION"
	EntityLocation         EntityType = "LOCATION"
	EntityHistorical       EntityType = "HISTORICAL_ENTITY"
	EntityProduct          EntityType = "PRODUCT"
	EntityDateTime         EntityType = "DATE_TIME"
	EntityQuantity         EntityType = "QUANTITY"
	EntityConcept          EntityType = "CONCEPT"
	EntityEvent            EntityType = "EVENT"
	EntityFile             EntityType = "FILE"
	EntityUnknown          EntityType = "UNKNOWN"
)

// ReferenceType
type ReferenceType string

const (
	RefPronoun             ReferenceType = "PRONOUN"
	RefDemonstrative       ReferenceType = "DEMONSTRATIVE"
	RefDefiniteDescription ReferenceType = "DEFINITE_DESCRIPTION"
)

// TemporalType
type TemporalType string

const (
	TempAbsolute   TemporalType = "ABSOLUTE"
	TempRelative   TemporalType = "RELATIVE"
	TempDuration   TemporalType = "DURATION"
	TempRecurrence TemporalType = "RECURRENCE"
)

// Valence
type Valence string

const (
	ValencePositive Valence = "POSITIVE"
	ValenceNeutral  Valence = "NEUTRAL"
	ValenceNegative Valence = "NEGATIVE"
)

// Intensity
type Intensity string

const (
	IntensityLow    Intensity = "LOW"
	IntensityMedium Intensity = "MEDIUM"
	IntensityHigh   Intensity = "HIGH"
)

// Formality
type Formality string

const (
	FormalityFormal   Formality = "FORMAL"
	FormalityInformal Formality = "INFORMAL"
	FormalityCasual   Formality = "CASUAL"
)

// Length
type Length string

const (
	LengthShort  Length = "SHORT"
	LengthMedium Length = "MEDIUM"
	LengthLong   Length = "LONG"
)

// Sub-types

type Slot struct {
	name        string
	value       string
	groundingID string
	confidence  float64
}

func NewSlot(name, value, groundingID string, confidence float64) Slot {
	return Slot{name: name, value: value, groundingID: groundingID, confidence: confidence}
}
func (s Slot) Name() string        { return s.name }
func (s Slot) Value() string       { return s.value }
func (s Slot) GroundingID() string { return s.groundingID }
func (s Slot) Confidence() float64 { return s.confidence }

type Hypothesis struct {
	intent           string
	confidence       float64
	deltaFromPrimary float64
	sourceLayer      SourceLayer
	slots            []Slot
}

func NewHypothesis(intent string, confidence, delta float64, source SourceLayer, slots []Slot) Hypothesis {
	s := make([]Slot, len(slots))
	copy(s, slots)
	return Hypothesis{intent: intent, confidence: confidence, deltaFromPrimary: delta, sourceLayer: source, slots: s}
}
func (h Hypothesis) Intent() string             { return h.intent }
func (h Hypothesis) Confidence() float64        { return h.confidence }
func (h Hypothesis) DeltaFromPrimary() float64  { return h.deltaFromPrimary }
func (h Hypothesis) SourceLayer() SourceLayer   { return h.sourceLayer }
func (h Hypothesis) Slots() []Slot              { s := make([]Slot, len(h.slots)); copy(s, h.slots); return s }

type Entity struct {
	surface       string
	eType         EntityType
	canonicalName string
	groundingID   string
	confidence    float64
}

func NewEntity(surface string, t EntityType, canonical, grounding string, conf float64) Entity {
	return Entity{surface: surface, eType: t, canonicalName: canonical, groundingID: grounding, confidence: conf}
}
func (e Entity) Surface() string       { return e.surface }
func (e Entity) Type() EntityType      { return e.eType }
func (e Entity) CanonicalName() string { return e.canonicalName }
func (e Entity) GroundingID() string   { return e.groundingID }
func (e Entity) Confidence() float64   { return e.confidence }

type Reference struct {
	surface           string
	rType             ReferenceType
	anchorHint        string
	anchorGroundingID string
	resolved          bool
	confidence        float64
}

func NewReference(surface string, t ReferenceType, hint, grounding string, resolved bool, conf float64) Reference {
	return Reference{surface: surface, rType: t, anchorHint: hint, anchorGroundingID: grounding, resolved: resolved, confidence: conf}
}
func (r Reference) Surface() string           { return r.surface }
func (r Reference) Type() ReferenceType       { return r.rType }
func (r Reference) AnchorHint() string        { return r.anchorHint }
func (r Reference) AnchorGroundingID() string { return r.anchorGroundingID }
func (r Reference) Resolved() bool            { return r.resolved }
func (r Reference) Confidence() float64       { return r.confidence }

type TemporalAnchor struct {
	surface    string
	tType      TemporalType
	normalized string
	confidence float64
}

func NewTemporalAnchor(surface string, t TemporalType, normalized string, conf float64) TemporalAnchor {
	return TemporalAnchor{surface: surface, tType: t, normalized: normalized, confidence: conf}
}
func (t TemporalAnchor) Surface() string       { return t.surface }
func (t TemporalAnchor) Type() TemporalType    { return t.tType }
func (t TemporalAnchor) Normalized() string    { return t.normalized }
func (t TemporalAnchor) Confidence() float64   { return t.confidence }

type SecondaryIntent struct {
	intent     string
	topics     []string
	slots      []Slot
	confidence float64
}

func NewSecondaryIntent(intent string, topics []string, slots []Slot, conf float64) SecondaryIntent {
	t := make([]string, len(topics))
	copy(t, topics)
	s := make([]Slot, len(slots))
	copy(s, slots)
	return SecondaryIntent{intent: intent, topics: t, slots: s, confidence: conf}
}
func (s SecondaryIntent) Intent() string      { return s.intent }
func (s SecondaryIntent) Topics() []string    { t := make([]string, len(s.topics)); copy(t, s.topics); return t }
func (s SecondaryIntent) Slots() []Slot       { sl := make([]Slot, len(s.slots)); copy(sl, s.slots); return sl }
func (s SecondaryIntent) Confidence() float64 { return s.confidence }

type Polarity struct {
	negated        bool
	negationMarker string
}

func NewPolarity(negated bool, marker string) Polarity {
	return Polarity{negated: negated, negationMarker: marker}
}
func (p Polarity) Negated() bool         { return p.negated }
func (p Polarity) NegationMarker() string { return p.negationMarker }

type Sentiment struct {
	valence   Valence
	intensity Intensity
	markers   []string
}

func NewSentiment(v Valence, i Intensity, markers []string) Sentiment {
	m := make([]string, len(markers))
	copy(m, markers)
	return Sentiment{valence: v, intensity: i, markers: m}
}
func (s Sentiment) Valence() Valence     { return s.valence }
func (s Sentiment) Intensity() Intensity { return s.intensity }
func (s Sentiment) Markers() []string    { m := make([]string, len(s.markers)); copy(m, s.markers); return m }

type InputCharacteristics struct {
	containsEmoji        bool
	containsSlang        bool
	containsAbbreviation bool
	likelyTypoRecovered  bool
	formality            Formality
	inputLength          Length
}

func NewInputCharacteristics(emoji, slang, abbr, typo bool, form Formality, length Length) InputCharacteristics {
	return InputCharacteristics{containsEmoji: emoji, containsSlang: slang, containsAbbreviation: abbr, likelyTypoRecovered: typo, formality: form, inputLength: length}
}
func (i InputCharacteristics) ContainsEmoji() bool        { return i.containsEmoji }
func (i InputCharacteristics) ContainsSlang() bool        { return i.containsSlang }
func (i InputCharacteristics) ContainsAbbreviation() bool { return i.containsAbbreviation }
func (i InputCharacteristics) LikelyTypoRecovered() bool  { return i.likelyTypoRecovered }
func (i InputCharacteristics) Formality() Formality       { return i.formality }
func (i InputCharacteristics) InputLength() Length        { return i.inputLength }

type DialoguePosition struct {
	turnIndex    int
	isFollowUp   bool
	isCorrection bool
	isNewTopic   bool
}

func NewDialoguePosition(turn int, follow, correction, newTopic bool) DialoguePosition {
	return DialoguePosition{turnIndex: turn, isFollowUp: follow, isCorrection: correction, isNewTopic: newTopic}
}
func (d DialoguePosition) TurnIndex() int      { return d.turnIndex }
func (d DialoguePosition) IsFollowUp() bool    { return d.isFollowUp }
func (d DialoguePosition) IsCorrection() bool  { return d.isCorrection }
func (d DialoguePosition) IsNewTopic() bool    { return d.isNewTopic }

type ExecutionHints struct {
	appearsDeterministic bool
	appearsOpenEnded     bool
	requiresExternalData bool
	requiresMemoryLookup bool
}

func NewExecutionHints(det, open, ext, mem bool) ExecutionHints {
	return ExecutionHints{appearsDeterministic: det, appearsOpenEnded: open, requiresExternalData: ext, requiresMemoryLookup: mem}
}
func (e ExecutionHints) AppearsDeterministic() bool { return e.appearsDeterministic }
func (e ExecutionHints) AppearsOpenEnded() bool     { return e.appearsOpenEnded }
func (e ExecutionHints) RequiresExternalData() bool { return e.requiresExternalData }
func (e ExecutionHints) RequiresMemoryLookup() bool { return e.requiresMemoryLookup }

type ValidationTrace struct {
	typeValidation         string
	constraintValidation   string
	relationshipValidation string
	temporalValidation     string
}

func NewValidationTrace(t, c, r, temp string) ValidationTrace {
	return ValidationTrace{typeValidation: t, constraintValidation: c, relationshipValidation: r, temporalValidation: temp}
}
func (v ValidationTrace) TypeValidation() string         { return v.typeValidation }
func (v ValidationTrace) ConstraintValidation() string   { return v.constraintValidation }
func (v ValidationTrace) RelationshipValidation() string { return v.relationshipValidation }
func (v ValidationTrace) TemporalValidation() string     { return v.temporalValidation }

type Assumption struct {
	category string
	details  string
}

func NewAssumption(cat, details string) Assumption {
	return Assumption{category: cat, details: details}
}
func (a Assumption) Category() string { return a.category }
func (a Assumption) Details() string  { return a.details }

type Ambiguity struct {
	ambiguousSpan        string
	candidateResolutions []string
	severity             string
}

func NewAmbiguity(span string, candidates []string, severity string) Ambiguity {
	c := make([]string, len(candidates))
	copy(c, candidates)
	return Ambiguity{ambiguousSpan: span, candidateResolutions: c, severity: severity}
}
func (a Ambiguity) AmbiguousSpan() string         { return a.ambiguousSpan }
func (a Ambiguity) CandidateResolutions() []string { c := make([]string, len(a.candidateResolutions)); copy(c, a.candidateResolutions); return c }
func (a Ambiguity) Severity() string              { return a.severity }

// Error declarations
var (
	ErrValidation = errors.New("semantic interpretation validation failed")
)
