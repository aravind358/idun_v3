package v3

import (
	"context"
	"fmt"
	"idun/core/foundation"
	reasoning "idun/intelligence/reasoning/v3"
	understanding "idun/intelligence/understanding/v3"
	"time"
)

// Orchestrator coordinates the Planning pipeline.
type Orchestrator struct {
	registry CapabilityRegistry
}

// NewOrchestrator creates a new planning Orchestrator with the given registry.
func NewOrchestrator(registry CapabilityRegistry) *Orchestrator {
	return &Orchestrator{registry: registry}
}

// Plan constructs a DAG of operations to fulfill the resolved intent.
func (o *Orchestrator) Plan(ctx context.Context, interp *understanding.SemanticInterpretation, reasonCtx *reasoning.ReasoningContext) (*ExecutionPlan, error) {
	uuidStr, _ := foundation.NewUUID()
	artifactID := foundation.ArtifactID(uuidStr)

	builder := NewBuilder().
		ArtifactID(artifactID).
		ParentArtifactID(foundation.ParentArtifactID(reasonCtx.ArtifactID())). // Trace back to ReasoningContext
		EnvelopeID(reasonCtx.EnvelopeID()).
		Timestamp(foundation.Timestamp(time.Now()))

	// 1. Goal Decomposition & Capability Selection
	// Mock: Match the ResolvedIntent directly to a capability
	caps, err := o.registry.Discover(ctx, reasonCtx.ResolvedIntent())
	if err != nil {
		return nil, fmt.Errorf("failed to discover capabilities: %w", err)
	}

	if len(caps) == 0 {
		return nil, fmt.Errorf("no capability found for goal: %s", reasonCtx.ResolvedIntent())
	}
	selectedCap := caps[0] // just pick the first matching capability for mock

	// 2. Parameter Binding
	// Bind EnrichedSlots and GroundedEntities from ReasoningContext to the required params.
	boundParams := make(map[string]any)
	for _, reqParam := range selectedCap.Params {
		// Attempt to find it in EnrichedSlots
		for _, slot := range reasonCtx.EnrichedSlots() {
			if slot.Original().Name() == reqParam {
				boundParams[reqParam] = slot.EnrichedValue()
			}
		}
		// If not found in slots, check grounded entities or resolved references if applicable
		// For the mock, we assume the slot mapped exactly if it was present.
	}

	// 3. Task Sequencing & DAG Construction (GraphBuilder)
	// For this phase, we generate a simple single node. In complex queries, this would
	// iterate sub-goals and create Dependency edges.
	nodeID, _ := foundation.NewUUID()
	node := NewPlanNode(nodeID, selectedCap.ID, boundParams)

	// Since there's only one node, there are no edges.
	builder.Nodes([]PlanNode{node})
	builder.Edges(nil)

	return builder.Build()
}
