package v3

import (
	"encoding/json"
	"fmt"
	"idun/core/foundation"
	understanding "idun/intelligence/understanding/v3"
	"time"
)

type jsonGroundedEntity struct {
	SurfaceText string  `json:"SurfaceText"`
	MemoryID    string  `json:"MemoryID"`
	Confidence  float64 `json:"Confidence"`
}
type jsonResolvedReference struct {
	Pronoun      string  `json:"Pronoun"`
	TargetEntity string  `json:"TargetEntity"`
	MemoryID     string  `json:"MemoryID"`
	Confidence   float64 `json:"Confidence"`
}
type jsonContextEvidence struct {
	Source    string  `json:"Source"`
	Content   string  `json:"Content"`
	Relevance float64 `json:"Relevance"`
}
type jsonEnrichedSlot struct {
	Name          string  `json:"Name"`
	Value         string  `json:"Value"`
	GroundingID   string  `json:"GroundingID"`
	Confidence    float64 `json:"Confidence"`
	EnrichedValue string  `json:"EnrichedValue"`
}
type jsonReasoningContext struct {
	SpecVersion        string                  `json:"SpecVersion"`
	ArtifactID         string                  `json:"ArtifactID"`
	ParentArtifactID   string                  `json:"ParentArtifactID"`
	EnvelopeID         string                  `json:"EnvelopeID"`
	Timestamp          time.Time               `json:"Timestamp"`
	ResolvedIntent     string                  `json:"ResolvedIntent"`
	CanProceed         bool                    `json:"CanProceed"`
	CommunicativeAct   understanding.CommunicativeAct `json:"CommunicativeAct"`
	ResolvedConfidence float64                 `json:"ResolvedConfidence"`
	EnrichedSlots      []jsonEnrichedSlot      `json:"EnrichedSlots"`
	GroundedEntities   []jsonGroundedEntity    `json:"GroundedEntities"`
	ResolvedReferences []jsonResolvedReference `json:"ResolvedReferences"`
	RetrievedContexts  []jsonContextEvidence   `json:"RetrievedContexts"`
	ConditionEvaluated bool                    `json:"ConditionEvaluated"`
	ConditionMet       bool                    `json:"ConditionMet"`
	TruthEvaluated     bool                    `json:"TruthEvaluated"`
	IsFactuallyTrue    bool                    `json:"IsFactuallyTrue"`
}

func (r *ReasoningContext) MarshalJSON() ([]byte, error) {
	if err := r.Validate(); err != nil {
		return nil, fmt.Errorf("cannot marshal invalid ReasoningContext: %w", err)
	}

	j := jsonReasoningContext{
		SpecVersion:        r.specVersion,
		ArtifactID:         string(r.artifactID),
		ParentArtifactID:   string(r.parentArtifactID),
		EnvelopeID:         string(r.envelopeID),
		Timestamp:          time.Time(r.timestamp),
		ResolvedIntent:     r.resolvedIntent,
		CanProceed:         r.canProceed,
		CommunicativeAct:   r.communicativeAct,
		ResolvedConfidence: r.resolvedConfidence,
		ConditionEvaluated: r.conditionEvaluated,
		ConditionMet:       r.conditionMet,
		TruthEvaluated:     r.truthEvaluated,
		IsFactuallyTrue:    r.isFactuallyTrue,
	}

	for _, s := range r.enrichedSlots {
		j.EnrichedSlots = append(j.EnrichedSlots, jsonEnrichedSlot{
			Name:          s.original.Name(),
			Value:         s.original.Value(),
			GroundingID:   s.original.GroundingID(),
			Confidence:    s.original.Confidence(),
			EnrichedValue: s.enrichedValue,
		})
	}
	for _, g := range r.groundedEntities {
		j.GroundedEntities = append(j.GroundedEntities, jsonGroundedEntity{SurfaceText: g.surfaceText, MemoryID: g.memoryID, Confidence: g.confidence})
	}
	for _, ref := range r.resolvedReferences {
		j.ResolvedReferences = append(j.ResolvedReferences, jsonResolvedReference{Pronoun: ref.pronoun, TargetEntity: ref.targetEntity, MemoryID: ref.memoryID, Confidence: ref.confidence})
	}
	for _, ctx := range r.retrievedContexts {
		j.RetrievedContexts = append(j.RetrievedContexts, jsonContextEvidence{Source: ctx.source, Content: ctx.content, Relevance: ctx.relevance})
	}

	return json.Marshal(j)
}

func (r *ReasoningContext) UnmarshalJSON(data []byte) error {
	var j jsonReasoningContext
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}

	r.specVersion = j.SpecVersion
	r.artifactID = foundation.ArtifactID(j.ArtifactID)
	r.parentArtifactID = foundation.ParentArtifactID(j.ParentArtifactID)
	r.envelopeID = foundation.EnvelopeID(j.EnvelopeID)
	r.timestamp = foundation.Timestamp(j.Timestamp)
	r.resolvedIntent = j.ResolvedIntent
	r.canProceed = j.CanProceed
	r.communicativeAct = j.CommunicativeAct
	r.resolvedConfidence = j.ResolvedConfidence
	r.conditionEvaluated = j.ConditionEvaluated
	r.conditionMet = j.ConditionMet
	r.truthEvaluated = j.TruthEvaluated
	r.isFactuallyTrue = j.IsFactuallyTrue

	r.enrichedSlots = make([]EnrichedSlot, len(j.EnrichedSlots))
	for i, s := range j.EnrichedSlots {
		orig := understanding.NewSlot(s.Name, s.Value, s.GroundingID, s.Confidence)
		r.enrichedSlots[i] = NewEnrichedSlot(orig, s.EnrichedValue)
	}

	r.groundedEntities = make([]GroundedEntity, len(j.GroundedEntities))
	for i, g := range j.GroundedEntities {
		r.groundedEntities[i] = NewGroundedEntity(g.SurfaceText, g.MemoryID, g.Confidence)
	}

	r.resolvedReferences = make([]ResolvedReference, len(j.ResolvedReferences))
	for i, ref := range j.ResolvedReferences {
		r.resolvedReferences[i] = NewResolvedReference(ref.Pronoun, ref.TargetEntity, ref.MemoryID, ref.Confidence)
	}

	r.retrievedContexts = make([]ContextEvidence, len(j.RetrievedContexts))
	for i, ctx := range j.RetrievedContexts {
		r.retrievedContexts[i] = NewContextEvidence(ctx.Source, ctx.Content, ctx.Relevance)
	}

	return r.Validate()
}
