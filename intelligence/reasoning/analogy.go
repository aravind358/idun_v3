package reasoning

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"idun/core/memory"
	"idun/intelligence/communication"
	"idun/intelligence/infrastructure/embedding"
	"idun/intelligence/understanding"
)

// HistoricalCasePayload defines the structured schema for stored case/episode records in memory.
type HistoricalCasePayload struct {
	Intent         string        `json:"intent,omitempty"`
	ResolvedGoal   *SemanticGoal `json:"resolved_goal,omitempty"`
	SemanticGoal   *SemanticGoal `json:"semantic_goal,omitempty"`
	ParameterSlots []string      `json:"parameter_slots,omitempty"`
}

// CaseAnalogySpecialist implements Stage S5 Case-Based / Analogical Reasoning.
// It retrieves structurally or semantically similar past experiences and generates
// analogical reasoning hypotheses.
//
// MANDATORY INVARIANTS:
// 1. Read-only experience retrieval: NEVER modifies memory records or embeddings.
// 2. NEVER creates new rules or compiles learning candidates.
type CaseAnalogySpecialist struct {
	embedder embedding.EmbeddingService
	mem      MemoryProvider
}

// NewCaseAnalogySpecialist constructs an initialized CaseAnalogySpecialist.
func NewCaseAnalogySpecialist(embedder embedding.EmbeddingService, mem MemoryProvider) *CaseAnalogySpecialist {
	return &CaseAnalogySpecialist{
		embedder: embedder,
		mem:      mem,
	}
}

// ID returns StageS5CaseAnalogy.
func (s *CaseAnalogySpecialist) ID() StageIdentifier {
	return StageS5CaseAnalogy
}

// EvaluateAnalogy retrieves similar cases and generates analogical hypotheses.
func (s *CaseAnalogySpecialist) EvaluateAnalogy(
	ctx context.Context,
	perceptionEnv communication.Envelope,
	frame *understanding.SemanticFrame,
	memContext []memory.Record,
) ([]ReasoningHypothesis, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	candidates := make([]memory.Record, 0, len(memContext))
	for _, rec := range memContext {
		if rec.Type == "case" || rec.Type == "episode" || strings.HasPrefix(rec.ID, "case/") {
			candidates = append(candidates, rec)
		}
	}
	if s.mem != nil {
		for _, t := range []string{"case", "episode"} {
			recs, err := s.mem.ListRecordsByType(t)
			if err == nil {
				candidates = append(candidates, recs...)
			}
		}
	}

	var hyps []ReasoningHypothesis
	queryIntent := ""
	if frame != nil {
		queryIntent = frame.PrimaryHypothesis.Intent
	}

	for i, rec := range candidates {
		simScore := 0.70 // Default baseline analogical similarity for matching case records
		if s.embedder != nil && perceptionEnv.PayloadRef != "" && rec.ID != "" {
			score, err := s.embedder.Similarity(ctx, perceptionEnv.PayloadRef, rec.ID)
			if err == nil && score >= 0.0 && score <= 1.0 {
				simScore = score
			}
		}

		if simScore >= 0.60 {
			conclusion := fmt.Sprintf("Analogical hypothesis derived from case %s (similarity=%.2f)", rec.ID, simScore)
			if queryIntent != "" {
				conclusion = fmt.Sprintf("Analogical match for intent %q from case %s (similarity=%.2f)", queryIntent, rec.ID, simScore)
			}

			var proposedGoal *SemanticGoal
			var missingArtifactReason string
			if len(rec.Payload) > 0 {
				var wrapper HistoricalCasePayload
				var candidateGoal *SemanticGoal
				var paramSlots []string

				if err := json.Unmarshal(rec.Payload, &wrapper); err == nil && (wrapper.ResolvedGoal != nil || wrapper.SemanticGoal != nil) {
					if wrapper.ResolvedGoal != nil && wrapper.ResolvedGoal.Validate() == nil {
						candidateGoal = wrapper.ResolvedGoal.Clone()
					} else if wrapper.SemanticGoal != nil && wrapper.SemanticGoal.Validate() == nil {
						candidateGoal = wrapper.SemanticGoal.Clone()
					}
					paramSlots = wrapper.ParameterSlots
				} else {
					// Try unmarshaling directly into SemanticGoal
					var directGoal SemanticGoal
					if err := json.Unmarshal(rec.Payload, &directGoal); err == nil && directGoal.Validate() == nil {
						candidateGoal = directGoal.Clone()
					}
				}

				if candidateGoal != nil {
					// Deterministic parameter adaptation grounded in current SemanticFrame slots
					if frame != nil && len(frame.PrimaryHypothesis.Slots) > 0 && candidateGoal.DesiredState != nil {
						slotMap := make(map[string]string, len(frame.PrimaryHypothesis.Slots))
						for _, slot := range frame.PrimaryHypothesis.Slots {
							slotMap[slot.Name] = slot.Value
						}
						// Adapt ONLY explicitly parameterizable fields
						if len(paramSlots) > 0 {
							for _, param := range paramSlots {
								if newVal, ok := slotMap[param]; ok {
									if _, hasKey := candidateGoal.DesiredState[param]; hasKey {
										candidateGoal.DesiredState[param] = newVal
									}
								}
							}
						} else if candidateGoal.Constraints != nil {
							// Also check if parameter_slots stored inside constraints
							if paramsStr, ok := candidateGoal.Constraints["parameter_slots"]; ok {
								params := strings.Split(paramsStr, ",")
								for _, param := range params {
									param = strings.TrimSpace(param)
									if newVal, exists := slotMap[param]; exists && param != "" {
										if _, hasKey := candidateGoal.DesiredState[param]; hasKey {
											candidateGoal.DesiredState[param] = newVal
										}
									}
								}
							}
						}
					}
					proposedGoal = candidateGoal
				} else {
					missingArtifactReason = "case record payload does not contain structured ResolvedGoal or SemanticGoal"
				}
			} else {
				missingArtifactReason = "case record payload is empty (Learning integration has not yet stored structured goals)"
			}

			premises := []string{fmt.Sprintf("analogical_case=%s", rec.ID)}
			if proposedGoal != nil {
				premises = append(premises, "stored_goal_recovered=true")
			} else if missingArtifactReason != "" {
				premises = append(premises, fmt.Sprintf("missing_learning_artifact=%s", missingArtifactReason))
			}

			hyps = append(hyps, ReasoningHypothesis{
				ID:                  fmt.Sprintf("s5-hyp-%s-%d", perceptionEnv.ID, i),
				Type:                HypothesisAnalogy,
				Conclusion:          conclusion,
				ProposedGoal:        proposedGoal,
				ReasoningConfidence: simScore,
				ContributingStages:  []StageIdentifier{StageS5CaseAnalogy},
				SupportingPremises:  premises,
				EvidenceTrace:       fmt.Sprintf("Retrieved via S5 Case-Based Analogy (score=%.2f)", simScore),
			})
		}
	}

	return hyps, nil
}
