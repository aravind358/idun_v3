// Package communication implements the Cognitive Communication Substrate for IDUN V3.
//
// Architecture Version: 2.0.0-FROZEN
//
// The communication package defines the canonical control-plane Envelope schema,
// leveled TopicID channels, and content-blind messaging invariants required by
// the Global Workspace and Executive Functions.
package communication

// TopicID identifies an orthogonal leveled workspace channel.
// Cognitive abilities subscribe only to the semantic topics relevant to their competence.
type TopicID string

const (
	// TopicPerception carries raw or pre-processed world/sensory stimuli.
	TopicPerception TopicID = "perception"

	// TopicUserIntent carries interpreted human intent representations.
	TopicUserIntent TopicID = "user-intent"

	// TopicActiveGoals carries active goal headers and decomposed target states.
	TopicActiveGoals TopicID = "active-goals"

	// TopicCandidatePlans carries candidate Hierarchical Task Networks (HTN) or plans.
	TopicCandidatePlans TopicID = "candidate-plans"

	// TopicEvaluatedOptions carries evaluated decision options and trade-off matrices.
	TopicEvaluatedOptions TopicID = "evaluated-options"

	// TopicReflections carries metacognitive audits, error analyses, and calibration updates.
	TopicReflections TopicID = "reflections"

	// TopicValueFlags carries constitutional alerts, ethics flags, and safety overrides.
	TopicValueFlags TopicID = "value-flags"

	// TopicImpasses carries content-blind SOAR-style impasse events when no confident bid wins.
	TopicImpasses TopicID = "impasses"

	// TopicActionExecution carries candidate external actions subject to Pre-Broadcast Constitutional Interception.
	TopicActionExecution TopicID = "action-execution"
)

// IsValid returns true if t is a registered canonical TopicID.
func (t TopicID) IsValid() bool {
	switch t {
	case TopicPerception,
		TopicUserIntent,
		TopicActiveGoals,
		TopicCandidatePlans,
		TopicEvaluatedOptions,
		TopicReflections,
		TopicValueFlags,
		TopicImpasses,
		TopicActionExecution:
		return true
	default:
		return false
	}
}

// AllTopics returns a slice of all canonical leveled Workspace topics.
func AllTopics() []TopicID {
	return []TopicID{
		TopicPerception,
		TopicUserIntent,
		TopicActiveGoals,
		TopicCandidatePlans,
		TopicEvaluatedOptions,
		TopicReflections,
		TopicValueFlags,
		TopicImpasses,
		TopicActionExecution,
	}
}
