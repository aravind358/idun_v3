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
	entities []Entity,
	refs []Reference,
	anchors []TemporalAnchor,
	composed []string,
	secondary []SecondaryIntent,
	goalIndex int,
	totalGoals int,
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
		CompoundIntentCount(1). // default mock value
		CommunicativeAct(ActConversation). // default mock value
		Entities(entities).
		References(refs).
		TemporalAnchors(anchors).
		ComposedTimestamps(composed).
		SecondaryIntents(secondary).
		Metadata(foundation.InteractionMetadata{
			GoalID:     string(artifactID), 
			GoalIndex:  goalIndex,
			TotalGoals: totalGoals,
		})

	if status == StatusFailed {
		builder.PrimaryIntent("unresolved_intent")
	}

	return builder.Build()
}
