package reasoning

import (
	"context"
	"math"
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
func (s *BayesianFusionSpecialist) FuseEvidence(ctx context.Context, hypotheses []ReasoningHypothesis) ([]ReasoningHypothesis, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	out := make([]ReasoningHypothesis, len(hypotheses))

	for i, hyp := range hypotheses {
		updated := hyp.Clone()

		priorConf := updated.ReasoningConfidence
		if priorConf <= 0.0 {
			priorConf = 0.80
		}
		priorLogOdds := probToLogOdds(priorConf)

		// Each supporting premise or contributing stage acts as an evidence update
		for _, premise := range updated.SupportingPremises {
			lower := strings.ToLower(premise)
			if strings.Contains(lower, "contradict") || strings.Contains(lower, "conflict") {
				priorLogOdds -= 1.1 // negative likelihood update for conflicting evidence
			} else {
				priorLogOdds += 0.35 // positive likelihood update for supporting premise
			}
		}

		// Additional evidence from multiple contributing stages
		if len(updated.ContributingStages) > 1 {
			priorLogOdds += float64(len(updated.ContributingStages)-1) * 0.25
		}

		fusedConfidence := logOddsToProb(priorLogOdds)
		updated.ReasoningConfidence = fusedConfidence
		updated.ContributingStages = appendUniqueStage(updated.ContributingStages, StageS4EvidenceFusion)
		updated.EvidenceTrace = updated.EvidenceTrace + "; Fused via Stage S4 Bayesian log-odds update"

		out[i] = updated
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
