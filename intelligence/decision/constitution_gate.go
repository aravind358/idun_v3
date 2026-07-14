package decision

import (
	"context"
	"strings"
)

// ProhibitedRiskTags defines explicit constitutional violation tags that trigger automatic Tier 1 rejection.
var ProhibitedRiskTags = map[string]struct{}{
	"SAFETY_VIOLATION":          {},
	"CONSTITUTIONAL_VETO":       {},
	"IRREVERSIBLE_UNSAFE":       {},
	"ETHICAL_BOUNDARY_BREACH":   {},
	"AXIOM_VIOLATION":           {},
}

// DefaultConstitutionalGate implements Tier1ConstitutionalGate, enforcing hard binary safety filters
// before any Tier 2 objective utility scoring occurs.
type DefaultConstitutionalGate struct{}

// NewDefaultConstitutionalGate constructs the Tier 1 constitutional hard gate.
func NewDefaultConstitutionalGate() *DefaultConstitutionalGate {
	return &DefaultConstitutionalGate{}
}

// Filter evaluates CandidateSet against non-negotiable constitutional safety invariants.
// Any violating candidates are immediately eliminated into RejectedAlternative records.
func (g *DefaultConstitutionalGate) Filter(ctx context.Context, cs CandidateSet) ([]Candidate, []RejectedAlternative, error) {
	if err := cs.Validate(); err != nil {
		return nil, nil, err
	}

	surviving := make([]Candidate, 0, len(cs.Candidates))
	rejected := make([]RejectedAlternative, 0)

	for _, cand := range cs.Candidates {
		vetoed := false
		reason := ""

		// 1. Check prohibited risk tags
		for _, flag := range cand.FlaggedRisks {
			upperFlag := strings.ToUpper(strings.TrimSpace(flag))
			if _, prohibited := ProhibitedRiskTags[upperFlag]; prohibited {
				vetoed = true
				reason = "CONSTITUTIONAL_GATE_VIOLATION: " + upperFlag
				break
			}
		}

		// 2. Check explicit constitutional_safety attribute if provided
		if !vetoed && cand.Attributes != nil {
			if safetyVal, exists := cand.Attributes["constitutional_safety"]; exists && safetyVal < 0.0 {
				vetoed = true
				reason = "CONSTITUTIONAL_GATE_VIOLATION: negative constitutional_safety attribute"
			}
		}

		if vetoed {
			rejected = append(rejected, RejectedAlternative{
				CandidateID:    cand.ID,
				RejectionStage: "TIER_1_CONSTITUTION",
				PrimaryReason:  reason,
				ScoreDelta:     0.0, // Set during Tier 2 or left 0 for Tier 1 hard elimination
			})
		} else {
			surviving = append(surviving, cand)
		}
	}

	return surviving, rejected, nil
}
