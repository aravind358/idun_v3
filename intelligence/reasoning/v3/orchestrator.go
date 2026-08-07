package v3

import (
	"context"
	"fmt"
	"idun/core/foundation"
	reasoning "idun/intelligence/reasoning"
	"idun/intelligence/communication"
	understanding "idun/intelligence/understanding/v3"
	"time"
)

// Orchestrator coordinates the Reasoning pipeline.
type Orchestrator struct {
	memory      MemoryProvider
	deliberative *reasoning.DeliberativeSpecialist
}

// NewOrchestrator creates a new reasoning Orchestrator.
// deliberative may be nil; if nil, Step 8 (LLM escalation) is skipped.
func NewOrchestrator(memory MemoryProvider, deliberative *reasoning.DeliberativeSpecialist) *Orchestrator {
	return &Orchestrator{memory: memory, deliberative: deliberative}
}

// Reason performs context enrichment on the incoming interpretation.
func (o *Orchestrator) Reason(ctx context.Context, interpretation *understanding.SemanticInterpretation) (*ReasoningContext, error) {
	uuidStr, _ := foundation.NewUUID()
	artifactID := foundation.ArtifactID(uuidStr)

	builder := NewBuilder().
		ArtifactID(artifactID).
		ParentArtifactID(foundation.ParentArtifactID(interpretation.ArtifactID())).
		EnvelopeID(interpretation.EnvelopeID()).
		Timestamp(foundation.Timestamp(time.Now()))

	// 1. Entity Grounding
	var groundedEntities []GroundedEntity
	for _, ent := range interpretation.Entities() {
		memID, conf, err := o.memory.RetrieveEntity(ctx, ent.Surface())
		if err == nil && memID != "" {
			groundedEntities = append(groundedEntities, NewGroundedEntity(ent.Surface(), memID, conf))
		}
	}
	builder.GroundedEntities(groundedEntities)

	// 2. Reference Resolution
	var resolvedRefs []ResolvedReference
	for _, ref := range interpretation.References() {
		target, memID, conf, err := o.memory.ResolveReference(ctx, ref.Surface())
		if err == nil && memID != "" {
			resolvedRefs = append(resolvedRefs, NewResolvedReference(ref.Surface(), target, memID, conf))
		}
	}
	builder.ResolvedReferences(resolvedRefs)

	// 3. Ambiguity Collapse
	// In a real implementation, ambiguity collapse would use context evidence to select
	// the best intent. For this phase, we deterministically select the primary intent.
	resolvedIntent := interpretation.PrimaryIntent()
	
	// If it's an unresolved impasse, but we have an ambiguity set, perhaps reasoning can salvage it
	// using context. Here, we'll just stick to the primary intent from Understanding.
	if interpretation.Status() == understanding.StatusAmbiguous && len(interpretation.AmbiguitySet()) > 0 {
		// Mock logic: choose primary.
		resolvedIntent = interpretation.PrimaryHypothesis().Intent()
	}
	builder.ResolvedIntent(resolvedIntent)

	// 4. Slot Enrichment
	// We only map slots that were transformed or enriched (e.g. grounded).
	// Mock: we'll just carry over the primary slots without enrichment for the stub, unless they match an entity.
	var enrichedSlots []EnrichedSlot
	for _, slot := range interpretation.PrimaryHypothesis().Slots() {
		// Check if we grounded this slot's value to an entity
		enrichedVal := slot.Value()
		for _, ge := range groundedEntities {
			if ge.SurfaceText() == slot.Value() {
				enrichedVal = ge.MemoryID()
				break
			}
		}
		
		// Check if this slot was temporally normalized
		if enrichedVal == slot.Value() {
			if len(interpretation.ComposedTimestamps()) > 0 && (slot.Name() == "time" || slot.Name() == "date" || slot.Name() == "temporal") {
				// Use the canonically composed timestamp if available for temporal slots
				enrichedVal = interpretation.ComposedTimestamps()[0]
			} else {
				for _, anchor := range interpretation.TemporalAnchors() {
					if anchor.Surface() == slot.Value() && anchor.Normalized() != "" {
						enrichedVal = anchor.Normalized()
						break
					}
				}
			}
		}

		// Always append the slot to ensure it reaches planning
		enrichedSlots = append(enrichedSlots, NewEnrichedSlot(slot, enrichedVal))
	}
	builder.EnrichedSlots(enrichedSlots)

	// 5. Context Retrieval
	contexts, _ := o.memory.RetrieveContext(ctx, resolvedIntent, interpretation.Topics())
	builder.RetrievedContexts(contexts)

	// 6. Conditional Evaluation
	if interpretation.IsConditional() {
		met, _ := o.memory.EvaluateCondition(ctx, interpretation.ConditionClause())
		builder.ConditionEvaluated(true, met)
	}

	// 7. Truth Evaluation
	// Mock: If it's a statement, evaluate its truth.
	if interpretation.CommunicativeAct() == understanding.ActStatement {
		isTrue, _ := o.memory.EvaluateFact(ctx, interpretation.PrimaryIntent())
		builder.TruthEvaluated(true, isTrue)
	}

	// 8. Deliberative Escalation
	// If the resolved intent is empty or ambiguous and a DeliberativeSpecialist is wired,
	// escalate to the LLM to attempt a higher-quality intent resolution.
	// This mirrors the S8 escalation in the V1 reasoning.Service (service.go:492-530).
	if o.deliberative != nil && (interpretation.Status() == understanding.StatusAmbiguous || interpretation.Status() == understanding.StatusFailed) {
		// Build a minimal envelope for the deliberative specialist to operate on.
		// The specialist uses the envelope's PayloadRef to retrieve raw perception data.
		synthEnv := communication.Envelope{
			ID:         string(interpretation.ArtifactID()),
			Topic:      communication.TopicResolvedIntent,
			CreatedAt:  time.Now().UTC(),
			PayloadRef: "Analyze ambiguous intent: " + interpretation.PrimaryIntent(),
		}
		
		fmt.Printf(">>> Reasoning V3: Ambiguity/Failure detected (Status: %s). Escalating to DeliberativeSpecialist (S8)...\n", interpretation.Status())
		
		_, err := o.deliberative.EvaluateDeliberative(ctx, synthEnv, 0.0, 0.65)
		if err != nil {
			fmt.Printf(">>> Reasoning V3: DeliberativeSpecialist error: %v\n", err)
		}
		// Results are discarded for this MVP integration;
		// the framework integration is confirmed active when the specialist executes.
		// A follow-up will wire the hypotheses back into the ReasoningContext builder.
	}

	return builder.Build()
}
