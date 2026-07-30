package v3

import (
	"idun/boundary/perception"
	"idun/core/foundation"
	"time"
)

// Synthesize constructs the final SemanticInterpretation artifact from the evaluated hypotheses
// and the original PerceptionEnvelope.
func Synthesize(
	env *perception.PerceptionEnvelope,
	primary Hypothesis,
	ambSet []Hypothesis,
	status InterpretationStatus,
) (*SemanticInterpretation, error) {
	
	// Create a new cognitive artifact ID
	uuidStr, _ := foundation.NewUUID()
	artifactID := foundation.ArtifactID(uuidStr)
	
	builder := NewBuilder().
		ArtifactID(artifactID).
		ParentArtifactID(foundation.ParentArtifactID(env.ArtifactID())).
		EnvelopeID(env.EnvelopeID()).
		Timestamp(foundation.Timestamp(time.Now())).
		Status(status).
		PrimaryIntent(primary.Intent()).
		PrimaryHypothesis(primary).
		AmbiguitySet(ambSet).
		Confidence(primary.Confidence()).
		Completeness(1.0). // default mock value
		CompoundIntentCount(1). // default mock value
		CommunicativeAct(ActConversation) // default mock value

	// In a real implementation, Topics, Entities, References, etc., would be mapped from
	// the primary Hypothesis slots or other outputs from the orchestrator logic.
	// For Phase 2 (Architecture validation), we ensure valid default states.

	if status == StatusFailed {
		builder.PrimaryIntent("unresolved_intent")
	}

	return builder.Build()
}
