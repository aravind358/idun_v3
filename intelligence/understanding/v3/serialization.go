package v3

import (
	"encoding/json"
	"fmt"
	"time"
	
	"idun/core/foundation"
)

// To handle private fields, we use mirroring structs for JSON serialization.

type jsonSlot struct {
	Name        string  `json:"Name"`
	Value       string  `json:"Value"`
	GroundingID string  `json:"GroundingID"`
	Confidence  float64 `json:"Confidence"`
}
type jsonHypothesis struct {
	Intent           string      `json:"Intent"`
	Confidence       float64     `json:"Confidence"`
	DeltaFromPrimary float64     `json:"DeltaFromPrimary"`
	SourceLayer      SourceLayer `json:"SourceLayer"`
	Slots            []jsonSlot  `json:"Slots"`
}
type jsonEntity struct {
	Surface       string     `json:"Surface"`
	Type          EntityType `json:"Type"`
	CanonicalName string     `json:"CanonicalName"`
	GroundingID   string     `json:"GroundingID"`
	Confidence    float64    `json:"Confidence"`
}
type jsonReference struct {
	Surface           string        `json:"Surface"`
	Type              ReferenceType `json:"Type"`
	AnchorHint        string        `json:"AnchorHint"`
	AnchorGroundingID string        `json:"AnchorGroundingID"`
	Resolved          bool          `json:"Resolved"`
	Confidence        float64       `json:"Confidence"`
}
type jsonTemporalAnchor struct {
	Surface    string       `json:"Surface"`
	Type       TemporalType `json:"Type"`
	Normalized string       `json:"Normalized"`
	Confidence float64      `json:"Confidence"`
}
type jsonSecondaryIntent struct {
	Intent     string     `json:"Intent"`
	Topics     []string   `json:"Topics"`
	Slots      []jsonSlot `json:"Slots"`
	Confidence float64    `json:"Confidence"`
}
type jsonPolarity struct {
	Negated        bool   `json:"Negated"`
	NegationMarker string `json:"NegationMarker"`
}
type jsonSentiment struct {
	Valence   Valence   `json:"Valence"`
	Intensity Intensity `json:"Intensity"`
	Markers   []string  `json:"Markers"`
}
type jsonInputCharacteristics struct {
	ContainsEmoji        bool      `json:"ContainsEmoji"`
	ContainsSlang        bool      `json:"ContainsSlang"`
	ContainsAbbreviation bool      `json:"ContainsAbbreviation"`
	LikelyTypoRecovered  bool      `json:"LikelyTypoRecovered"`
	Formality            Formality `json:"Formality"`
	InputLength          Length    `json:"InputLength"`
}
type jsonDialoguePosition struct {
	TurnIndex    int  `json:"TurnIndex"`
	IsFollowUp   bool `json:"IsFollowUp"`
	IsCorrection bool `json:"IsCorrection"`
	IsNewTopic   bool `json:"IsNewTopic"`
}
type jsonExecutionHints struct {
	AppearsDeterministic bool `json:"AppearsDeterministic"`
	AppearsOpenEnded     bool `json:"AppearsOpenEnded"`
	RequiresExternalData bool `json:"RequiresExternalData"`
	RequiresMemoryLookup bool `json:"RequiresMemoryLookup"`
}
type jsonValidationTrace struct {
	TypeValidation         string `json:"TypeValidation"`
	ConstraintValidation   string `json:"ConstraintValidation"`
	RelationshipValidation string `json:"RelationshipValidation"`
	TemporalValidation     string `json:"TemporalValidation"`
}
type jsonAssumption struct {
	Category string `json:"Category"`
	Details  string `json:"Details"`
}
type jsonAmbiguity struct {
	AmbiguousSpan        string   `json:"AmbiguousSpan"`
	CandidateResolutions []string `json:"CandidateResolutions"`
	Severity             string   `json:"Severity"`
}
type jsonSemanticInterpretation struct {
	SpecVersion          string                 `json:"SpecVersion"`
	ArtifactID           string                 `json:"ArtifactID"`
	ParentArtifactID     string                 `json:"ParentArtifactID,omitempty"`
	EnvelopeID           string                 `json:"EnvelopeID"`
	Timestamp            time.Time              `json:"Timestamp"`
	Status               InterpretationStatus   `json:"Status"`
	PrimaryIntent        string                 `json:"PrimaryIntent"`
	CommunicativeAct     CommunicativeAct       `json:"CommunicativeAct"`
	PrimaryHypothesis    jsonHypothesis         `json:"PrimaryHypothesis"`
	AmbiguitySet         []jsonHypothesis       `json:"AmbiguitySet"`
	Topics               []string               `json:"Topics"`
	Entities             []jsonEntity           `json:"Entities"`
	References           []jsonReference        `json:"References"`
	TemporalAnchors      []jsonTemporalAnchor   `json:"TemporalAnchors"`
	OpenSlots            []string               `json:"OpenSlots"`
	StatusReason         string                 `json:"StatusReason"`
	IsConditional        bool                   `json:"IsConditional"`
	ConditionClause      string                 `json:"ConditionClause"`
	ConsequentClause     string                 `json:"ConsequentClause"`
	CompoundIntentCount  int                    `json:"CompoundIntentCount"`
	SecondaryIntents     []jsonSecondaryIntent  `json:"SecondaryIntents"`
	RequiresContext      bool                   `json:"RequiresContext"`
	Polarity             jsonPolarity           `json:"Polarity"`
	Sentiment            jsonSentiment          `json:"Sentiment"`
	InputCharacteristics jsonInputCharacteristics `json:"InputCharacteristics"`
	DialoguePosition     jsonDialoguePosition   `json:"DialoguePosition"`
	ExecutionHints       jsonExecutionHints     `json:"ExecutionHints"`
	Assumptions          []jsonAssumption       `json:"Assumptions"`
	Ambiguities          []jsonAmbiguity        `json:"Ambiguities"`
	MissingInformation   []string               `json:"MissingInformation"`
	Completeness         float64                `json:"Completeness"`
	Confidence           float64                `json:"Confidence"`
	ProcessedDurationMs  float64                `json:"ProcessedDurationMs"`
	ValidationTrace      *jsonValidationTrace   `json:"ValidationTrace"`
	ConfidenceTrace      []string               `json:"ConfidenceTrace"`
}

func (s *SemanticInterpretation) MarshalJSON() ([]byte, error) {
	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("cannot marshal invalid SemanticInterpretation: %w", err)
	}

	j := jsonSemanticInterpretation{
		SpecVersion:         s.specVersion,
		ArtifactID:          string(s.artifactID),
		ParentArtifactID:    string(s.parentArtifactID),
		EnvelopeID:          string(s.envelopeID),
		Timestamp:           time.Time(s.timestamp),
		Status:              s.status,
		PrimaryIntent:       s.primaryIntent,
		CommunicativeAct:    s.communicativeAct,
		PrimaryHypothesis:   marshalHypothesis(s.primaryHypothesis),
		Topics:              s.topics,
		OpenSlots:           s.openSlots,
		StatusReason:        s.statusReason,
		IsConditional:       s.isConditional,
		ConditionClause:     s.conditionClause,
		ConsequentClause:    s.consequentClause,
		CompoundIntentCount: s.compoundIntentCount,
		RequiresContext:     s.requiresContext,
		MissingInformation:  s.missingInformation,
		Completeness:        s.completeness,
		Confidence:          s.confidence,
		ProcessedDurationMs: s.processedDurationMs,
		ConfidenceTrace:     s.confidenceTrace,
		Polarity: jsonPolarity{
			Negated:        s.polarity.negated,
			NegationMarker: s.polarity.negationMarker,
		},
		Sentiment: jsonSentiment{
			Valence:   s.sentiment.valence,
			Intensity: s.sentiment.intensity,
			Markers:   s.sentiment.markers,
		},
		InputCharacteristics: jsonInputCharacteristics{
			ContainsEmoji:        s.inputCharacteristics.containsEmoji,
			ContainsSlang:        s.inputCharacteristics.containsSlang,
			ContainsAbbreviation: s.inputCharacteristics.containsAbbreviation,
			LikelyTypoRecovered:  s.inputCharacteristics.likelyTypoRecovered,
			Formality:            s.inputCharacteristics.formality,
			InputLength:          s.inputCharacteristics.inputLength,
		},
		DialoguePosition: jsonDialoguePosition{
			TurnIndex:    s.dialoguePosition.turnIndex,
			IsFollowUp:   s.dialoguePosition.isFollowUp,
			IsCorrection: s.dialoguePosition.isCorrection,
			IsNewTopic:   s.dialoguePosition.isNewTopic,
		},
		ExecutionHints: jsonExecutionHints{
			AppearsDeterministic: s.executionHints.appearsDeterministic,
			AppearsOpenEnded:     s.executionHints.appearsOpenEnded,
			RequiresExternalData: s.executionHints.requiresExternalData,
			RequiresMemoryLookup: s.executionHints.requiresMemoryLookup,
		},
	}

	for _, h := range s.ambiguitySet {
		j.AmbiguitySet = append(j.AmbiguitySet, marshalHypothesis(h))
	}
	for _, e := range s.entities {
		j.Entities = append(j.Entities, jsonEntity{Surface: e.surface, Type: e.eType, CanonicalName: e.canonicalName, GroundingID: e.groundingID, Confidence: e.confidence})
	}
	for _, r := range s.references {
		j.References = append(j.References, jsonReference{Surface: r.surface, Type: r.rType, AnchorHint: r.anchorHint, AnchorGroundingID: r.anchorGroundingID, Resolved: r.resolved, Confidence: r.confidence})
	}
	for _, t := range s.temporalAnchors {
		j.TemporalAnchors = append(j.TemporalAnchors, jsonTemporalAnchor{Surface: t.surface, Type: t.tType, Normalized: t.normalized, Confidence: t.confidence})
	}
	for _, si := range s.secondaryIntents {
		jsi := jsonSecondaryIntent{Intent: si.intent, Topics: si.topics, Confidence: si.confidence}
		for _, sl := range si.slots {
			jsi.Slots = append(jsi.Slots, jsonSlot{Name: sl.name, Value: sl.value, GroundingID: sl.groundingID, Confidence: sl.confidence})
		}
		j.SecondaryIntents = append(j.SecondaryIntents, jsi)
	}
	for _, a := range s.assumptions {
		j.Assumptions = append(j.Assumptions, jsonAssumption{Category: a.category, Details: a.details})
	}
	for _, a := range s.ambiguities {
		j.Ambiguities = append(j.Ambiguities, jsonAmbiguity{AmbiguousSpan: a.ambiguousSpan, CandidateResolutions: a.candidateResolutions, Severity: a.severity})
	}
	if s.validationTrace != nil {
		j.ValidationTrace = &jsonValidationTrace{
			TypeValidation:         s.validationTrace.typeValidation,
			ConstraintValidation:   s.validationTrace.constraintValidation,
			RelationshipValidation: s.validationTrace.relationshipValidation,
			TemporalValidation:     s.validationTrace.temporalValidation,
		}
	}

	return json.Marshal(j)
}

func marshalHypothesis(h Hypothesis) jsonHypothesis {
	jh := jsonHypothesis{
		Intent:           h.intent,
		Confidence:       h.confidence,
		DeltaFromPrimary: h.deltaFromPrimary,
		SourceLayer:      h.sourceLayer,
	}
	for _, s := range h.slots {
		jh.Slots = append(jh.Slots, jsonSlot{Name: s.name, Value: s.value, GroundingID: s.groundingID, Confidence: s.confidence})
	}
	return jh
}

func unmarshalHypothesis(jh jsonHypothesis) Hypothesis {
	slots := make([]Slot, len(jh.Slots))
	for i, s := range jh.Slots {
		slots[i] = NewSlot(s.Name, s.Value, s.GroundingID, s.Confidence)
	}
	return NewHypothesis(jh.Intent, jh.Confidence, jh.DeltaFromPrimary, jh.SourceLayer, slots)
}

func (s *SemanticInterpretation) UnmarshalJSON(data []byte) error {
	var j jsonSemanticInterpretation
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	
	s.specVersion = j.SpecVersion
	s.artifactID = foundation.ArtifactID(j.ArtifactID)
	s.parentArtifactID = foundation.ParentArtifactID(j.ParentArtifactID)
	s.envelopeID = foundation.EnvelopeID(j.EnvelopeID)
	s.timestamp = foundation.Timestamp(j.Timestamp)
	s.status = j.Status
	s.primaryIntent = j.PrimaryIntent
	s.communicativeAct = j.CommunicativeAct
	s.primaryHypothesis = unmarshalHypothesis(j.PrimaryHypothesis)
	s.topics = j.Topics
	s.openSlots = j.OpenSlots
	s.statusReason = j.StatusReason
	s.isConditional = j.IsConditional
	s.conditionClause = j.ConditionClause
	s.consequentClause = j.ConsequentClause
	s.compoundIntentCount = j.CompoundIntentCount
	s.requiresContext = j.RequiresContext
	s.missingInformation = j.MissingInformation
	s.completeness = j.Completeness
	s.confidence = j.Confidence
	s.processedDurationMs = j.ProcessedDurationMs
	s.confidenceTrace = j.ConfidenceTrace

	s.polarity = NewPolarity(j.Polarity.Negated, j.Polarity.NegationMarker)
	s.sentiment = NewSentiment(j.Sentiment.Valence, j.Sentiment.Intensity, j.Sentiment.Markers)
	s.inputCharacteristics = NewInputCharacteristics(j.InputCharacteristics.ContainsEmoji, j.InputCharacteristics.ContainsSlang, j.InputCharacteristics.ContainsAbbreviation, j.InputCharacteristics.LikelyTypoRecovered, j.InputCharacteristics.Formality, j.InputCharacteristics.InputLength)
	s.dialoguePosition = NewDialoguePosition(j.DialoguePosition.TurnIndex, j.DialoguePosition.IsFollowUp, j.DialoguePosition.IsCorrection, j.DialoguePosition.IsNewTopic)
	s.executionHints = NewExecutionHints(j.ExecutionHints.AppearsDeterministic, j.ExecutionHints.AppearsOpenEnded, j.ExecutionHints.RequiresExternalData, j.ExecutionHints.RequiresMemoryLookup)

	if j.ValidationTrace != nil {
		vt := NewValidationTrace(j.ValidationTrace.TypeValidation, j.ValidationTrace.ConstraintValidation, j.ValidationTrace.RelationshipValidation, j.ValidationTrace.TemporalValidation)
		s.validationTrace = &vt
	}

	s.ambiguitySet = make([]Hypothesis, len(j.AmbiguitySet))
	for i, a := range j.AmbiguitySet {
		s.ambiguitySet[i] = unmarshalHypothesis(a)
	}
	s.entities = make([]Entity, len(j.Entities))
	for i, e := range j.Entities {
		s.entities[i] = NewEntity(e.Surface, e.Type, e.CanonicalName, e.GroundingID, e.Confidence)
	}
	s.references = make([]Reference, len(j.References))
	for i, r := range j.References {
		s.references[i] = NewReference(r.Surface, r.Type, r.AnchorHint, r.AnchorGroundingID, r.Resolved, r.Confidence)
	}
	s.temporalAnchors = make([]TemporalAnchor, len(j.TemporalAnchors))
	for i, t := range j.TemporalAnchors {
		s.temporalAnchors[i] = NewTemporalAnchor(t.Surface, t.Type, t.Normalized, t.Confidence)
	}
	s.secondaryIntents = make([]SecondaryIntent, len(j.SecondaryIntents))
	for i, si := range j.SecondaryIntents {
		slots := make([]Slot, len(si.Slots))
		for k, sl := range si.Slots {
			slots[k] = NewSlot(sl.Name, sl.Value, sl.GroundingID, sl.Confidence)
		}
		s.secondaryIntents[i] = NewSecondaryIntent(si.Intent, si.Topics, slots, si.Confidence)
	}
	s.assumptions = make([]Assumption, len(j.Assumptions))
	for i, a := range j.Assumptions {
		s.assumptions[i] = NewAssumption(a.Category, a.Details)
	}
	s.ambiguities = make([]Ambiguity, len(j.Ambiguities))
	for i, a := range j.Ambiguities {
		s.ambiguities[i] = NewAmbiguity(a.AmbiguousSpan, a.CandidateResolutions, a.Severity)
	}

	if err := s.Validate(); err != nil {
		return fmt.Errorf("unmarshaled invalid SemanticInterpretation: %w", err)
	}

	return nil
}
