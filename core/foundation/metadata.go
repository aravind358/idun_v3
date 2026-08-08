package foundation

// InteractionMetadata represents the canonical transport metadata shared
// across the cognitive pipeline. It flows from Understanding all the way
// to the Output layer, ensuring semantic execution ordering and provenance
// are preserved.
type InteractionMetadata struct {
	GoalID     string `json:"goal_id,omitempty"`
	GoalIndex  int    `json:"goal_index"`
	TotalGoals int    `json:"total_goals"`
	PayloadRef string `json:"payload_ref,omitempty"`
}
