package reasoning

import (
	"context"
	"fmt"

	"idun/intelligence/communication"
	"idun/intelligence/executive"
)

// StrategySpec specifies a dynamic routing policy that dictates which reasoning cascade
// stages run and enforces session-scoped graph limits during evaluation.
type StrategySpec struct {
	StrategyID    StrategyIdentifier `json:"strategy_id"`
	EnabledStages []StageIdentifier  `json:"enabled_stages"`
	PriorityOrder []StageIdentifier  `json:"priority_order"`
	MaxBudgetMs   float64            `json:"max_budget_ms"`
	MaxGraphNodes int                `json:"max_graph_nodes"`
	MaxGraphEdges       int                `json:"max_graph_edges"`
	MaxGraphDepth       int                `json:"max_graph_depth"`
	EscalationThreshold float64            `json:"escalation_threshold,omitempty"`
}

// Validate checks whether StrategySpec satisfies all structural and resource invariant bounds.
func (s StrategySpec) Validate() error {
	if s.StrategyID == "" {
		return fmt.Errorf("reasoning: strategy specification missing ID")
	}
	if s.MaxGraphNodes <= 0 || s.MaxGraphEdges <= 0 || s.MaxGraphDepth <= 0 {
		return fmt.Errorf("%w: max nodes %d, edges %d, depth %d must be > 0", ErrInvalidGraphLimits, s.MaxGraphNodes, s.MaxGraphEdges, s.MaxGraphDepth)
	}
	return nil
}

// Clone returns a deep copy of StrategySpec.
func (s StrategySpec) Clone() StrategySpec {
	out := s
	if len(s.EnabledStages) > 0 {
		out.EnabledStages = make([]StageIdentifier, len(s.EnabledStages))
		copy(out.EnabledStages, s.EnabledStages)
	} else {
		out.EnabledStages = []StageIdentifier{}
	}
	if len(s.PriorityOrder) > 0 {
		out.PriorityOrder = make([]StageIdentifier, len(s.PriorityOrder))
		copy(out.PriorityOrder, s.PriorityOrder)
	} else {
		out.PriorityOrder = []StageIdentifier{}
	}
	return out
}

// IsStageEnabled checks whether a specific cascade stage is enabled in this specification.
func (s StrategySpec) IsStageEnabled(stage StageIdentifier) bool {
	for _, enabled := range s.EnabledStages {
		if enabled == stage {
			return true
		}
	}
	return false
}

// ReasoningService defines the public contract for CognitiveAbility.Reasoning.
// It implements executive.ReasoningAbility and provides structured envelope-based reasoning.
type ReasoningService interface {
	executive.ReasoningAbility

	// Start boots the Reasoning service and initializes dependencies.
	Start() error

	// Close gracefully shuts down the Reasoning service.
	Close() error

	// ReasonEnvelope derives calibrated conclusions from an incoming perceptual Envelope
	// under an explicitly supplied StrategySpec routing policy.
	ReasonEnvelope(ctx context.Context, perceptionEnv communication.Envelope, spec StrategySpec) (ReasoningResult, error)
}

// StrategySelector defines the abstraction for selecting or evaluating a StrategySpec
// for a given perceptual input envelope and context.
type StrategySelector interface {
	SelectStrategy(ctx context.Context, perceptionEnv communication.Envelope) (StrategySpec, error)
}

// Specialist defines the common contract for all Reasoning Cascade specialists.
type Specialist interface {
	// ID returns the cascade stage identifier for this specialist.
	ID() StageIdentifier
}

// PayloadStorer defines the interface required to persist and retrieve payloads to/from CAS storage.
// Reasoning uses this to store the ReasoningResult before publishing the envelope PayloadRef.
type PayloadStorer interface {
	Store(ctx context.Context, data []byte) (string, error)
	Retrieve(ctx context.Context, key string) ([]byte, error)
}
