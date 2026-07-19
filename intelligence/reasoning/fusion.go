package reasoning

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
)

// BayesianFusionSpecialist implements Stage S4 Bayesian Evidence Fusion.
// It accumulates independent evidence sources and updates candidate hypothesis
// confidences via log-odds Bayesian likelihood updating.
//
// MANDATORY INVARIANTS:
// 1. Updates candidate hypothesis ReasoningConfidence ONLY.
// 2. NEVER chooses actions, performs planning, or executes decisions.
// 3. NEVER modifies Memory or Storage.
type BayesianFusionSpecialist struct{}

// NewBayesianFusionSpecialist returns an initialized BayesianFusionSpecialist.
func NewBayesianFusionSpecialist() *BayesianFusionSpecialist {
	return &BayesianFusionSpecialist{}
}

// ID returns StageS4EvidenceFusion.
func (s *BayesianFusionSpecialist) ID() StageIdentifier {
	return StageS4EvidenceFusion
}

// probToLogOdds converts probability p in (0, 1) to log-odds.
func probToLogOdds(p float64) float64 {
	if p <= 0.001 {
		p = 0.001
	}
	if p >= 0.999 {
		p = 0.999
	}
	return math.Log(p / (1.0 - p))
}

// logOddsToProb converts log-odds L to probability in (0, 1).
func logOddsToProb(l float64) float64 {
	p := 1.0 / (1.0 + math.Exp(-l))
	if p <= 0.01 {
		p = 0.01
	}
	if p >= 0.99 {
		p = 0.99
	}
	return p
}

// FuseEvidence combines evidence across candidate hypotheses and updates confidence.
// Hypotheses containing equivalent valid ProposedGoals are grouped by fingerprint into
// a single corroborated candidate before Bayesian log-odds likelihood updating.
func (s *BayesianFusionSpecialist) FuseEvidence(ctx context.Context, hypotheses []ReasoningHypothesis) ([]ReasoningHypothesis, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Step 1: Group hypotheses by valid SemanticGoal fingerprint.
	// Hypotheses without a valid ProposedGoal (or nil) cannot be grouped and remain separate candidates.
	groups := make(map[string][]ReasoningHypothesis)
	var order []string

	for _, hyp := range hypotheses {
		var fp string
		if hyp.ProposedGoal != nil && hyp.ProposedGoal.Validate() == nil {
			fp = hyp.ProposedGoal.Fingerprint()
		}
		if fp != "" {
			if _, exists := groups[fp]; !exists {
				order = append(order, fp)
			}
			groups[fp] = append(groups[fp], hyp)
		} else {
			// Ungrouped individual candidate with unique placeholder key preserving order
			key := fmt.Sprintf("ungrouped:%s:%d", hyp.ID, len(order))
			order = append(order, key)
			groups[key] = []ReasoningHypothesis{hyp}
		}
	}

	out := make([]ReasoningHypothesis, 0, len(order))

	// Step 2: For each group or individual candidate, construct the fused representative
	// and apply Bayesian log-odds likelihood updates.
	for _, key := range order {
		group := groups[key]
		if len(group) == 0 {
			continue
		}

		// Sort group descending by ReasoningConfidence (and ID for tie-breaking determinism)
		// so the highest-confidence candidate serves as the base representative.
		sortedGroup := make([]ReasoningHypothesis, len(group))
		copy(sortedGroup, group)
		sort.Slice(sortedGroup, func(i, j int) bool {
			if sortedGroup[i].ReasoningConfidence != sortedGroup[j].ReasoningConfidence {
				return sortedGroup[i].ReasoningConfidence > sortedGroup[j].ReasoningConfidence
			}
			return sortedGroup[i].ID < sortedGroup[j].ID
		})

		updated := sortedGroup[0].Clone()

		// Merge SupportingPremises and ContributingStages across all corroborating hypotheses in the group
		seenPremises := make(map[string]bool)
		var mergedPremises []string
		for _, p := range updated.SupportingPremises {
			if !seenPremises[p] {
				seenPremises[p] = true
				mergedPremises = append(mergedPremises, p)
			}
		}

		for i := 1; i < len(sortedGroup); i++ {
			corroborator := sortedGroup[i]
			for _, st := range corroborator.ContributingStages {
				updated.ContributingStages = appendUniqueStage(updated.ContributingStages, st)
			}
			for _, p := range corroborator.SupportingPremises {
				if !seenPremises[p] {
					seenPremises[p] = true
					mergedPremises = append(mergedPremises, p)
				}
			}
		}
		updated.SupportingPremises = mergedPremises

		// Step 3: Bayesian log-odds likelihood update
		priorConf := updated.ReasoningConfidence
		if priorConf <= 0.0 {
			priorConf = 0.80
		}
		priorLogOdds := probToLogOdds(priorConf)

		// Each unique supporting premise or contradiction acts as an evidence update
		for _, premise := range updated.SupportingPremises {
			lower := strings.ToLower(premise)
			if strings.Contains(lower, "contradict") || strings.Contains(lower, "conflict") {
				priorLogOdds -= 1.1 // negative likelihood update for conflicting evidence
			} else {
				priorLogOdds += 0.35 // positive likelihood update for supporting premise
			}
		}

		// Additional evidence from multiple contributing stages across corroborating hypotheses
		if len(updated.ContributingStages) > 1 {
			priorLogOdds += float64(len(updated.ContributingStages)-1) * 0.25
		}

		// Additional positive log-odds update for each corroborating hypothesis (same goal from multiple specialists)
		if len(sortedGroup) > 1 {
			for i := 1; i < len(sortedGroup); i++ {
				corConf := sortedGroup[i].ReasoningConfidence
				if corConf > 0.0 {
					// Add weighted corroboration log-odds contribution
					priorLogOdds += math.Max(0.20, probToLogOdds(corConf)*0.40)
				} else {
					priorLogOdds += 0.20
				}
			}
		}

		fusedConfidence := logOddsToProb(priorLogOdds)
		updated.ReasoningConfidence = fusedConfidence
		updated.ContributingStages = appendUniqueStage(updated.ContributingStages, StageS4EvidenceFusion)
		if len(sortedGroup) > 1 {
			updated.EvidenceTrace = fmt.Sprintf("%s; Fused across %d corroborating specialists via Stage S4 Bayesian log-odds update", updated.EvidenceTrace, len(sortedGroup))
		} else {
			updated.EvidenceTrace = updated.EvidenceTrace + "; Fused via Stage S4 Bayesian log-odds update"
		}

		out = append(out, updated)
	}

	return out, nil
}

func appendUniqueStage(stages []StageIdentifier, target StageIdentifier) []StageIdentifier {
	for _, s := range stages {
		if s == target {
			return stages
		}
	}
	return append(stages, target)
}
