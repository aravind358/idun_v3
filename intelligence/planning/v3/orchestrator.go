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

// Plan constructs a DAG of operations to fulfill the resolved intents.
func (o *Orchestrator) Plan(ctx context.Context, interp *understanding.SemanticInterpretation, reasonCtxs []*reasoning.ReasoningContext) (*ExecutionPlan, error) {
	uuidStr, _ := foundation.NewUUID()
	artifactID := foundation.ArtifactID(uuidStr)

	// We trace back to the first context if available
	parentArtifactID := ""
	envID := foundation.EnvelopeID("")
	if len(reasonCtxs) > 0 {
		parentArtifactID = string(reasonCtxs[0].ArtifactID())
		envID = reasonCtxs[0].EnvelopeID()
	}

	builder := NewBuilder().
		ArtifactID(artifactID).
		ParentArtifactID(foundation.ParentArtifactID(parentArtifactID)).
		EnvelopeID(envID).
		Timestamp(foundation.Timestamp(time.Now()))

	var nodes []PlanNode
	var edges []Dependency

	var previousNodeID string

	for _, reasonCtx := range reasonCtxs {
		// 1. Goal Decomposition & Capability Selection
		caps, err := o.registry.Discover(ctx, reasonCtx.ResolvedIntent())
		if err != nil {
			return nil, fmt.Errorf("failed to discover capabilities: %w", err)
		}

		var selectedCap CapabilityDescriptor
		if len(caps) == 0 {
			selectedCap = CapabilityDescriptor{
				ID:          CapabilityID("mock.capability"),
				Description: "Mock capability for " + reasonCtx.ResolvedIntent(),
			}
		} else {
			selectedCap = caps[0]
		}

		// 2. Parameter Binding
		boundParams := make(map[string]any)
		for _, reqParam := range selectedCap.Params {
			for _, slot := range reasonCtx.EnrichedSlots() {
				if slot.Original().Name() == reqParam {
					boundParams[reqParam] = slot.EnrichedValue()
				}
			}
		}

		// 3. Task Sequencing & DAG Construction
		nodeIDStr, _ := foundation.NewUUID()
		node := NewPlanNode(nodeIDStr, selectedCap.ID, boundParams, "NONE")
		nodes = append(nodes, node)

		// Create sequential dependency if not the first node
		if previousNodeID != "" {
			edges = append(edges, NewDependency(previousNodeID, nodeIDStr))
		}
		previousNodeID = nodeIDStr
	}

	builder.Nodes(nodes)
	builder.Edges(edges)

	return builder.Build()
}
