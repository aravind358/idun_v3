package decision

import (
	"fmt"
	"math"
)

// CalibrateConfidence modulates a candidate's raw confidence interval based on attribute completeness,
// top-two score margin, and active policy risk tolerance.
func CalibrateConfidence(rawConfidence float64, topMargin float64, attrCoverage float64, riskTolerance float64) (float64, []string) {
	var flags []string
	conf := rawConfidence

	// Penalize low attribute coverage
	if attrCoverage < 1.0 {
		conf *= attrCoverage
		flags = append(flags, fmt.Sprintf("LOW_ATTRIBUTE_COVERAGE: %.2f", attrCoverage))
	}

	// Penalize tight ambiguity margins
	if topMargin < 0.05 {
		penalty := (0.05 - topMargin) * 4.0 // up to 0.20 drop
		conf -= penalty
		flags = append(flags, fmt.Sprintf("TIGHT_AMBIGUITY_MARGIN: %.4f", topMargin))
	}

	// Adjust for conservative risk tolerance
	if riskTolerance < 0.30 {
		conf *= 0.95 // conservative deflation
	}

	conf = math.Max(0.01, math.Min(0.9999, conf))
	return conf, flags
}

// IdentifyInformationGaps analyzes a CandidateSet against expected strategy features
// to detect missing attributes or unverified assumptions.
func IdentifyInformationGaps(candidates []Candidate, expectedFeatures map[string]float64) []InformationGap {
	if len(candidates) == 0 || len(expectedFeatures) == 0 {
		return nil
	}

	var gaps []InformationGap
	for _, cand := range candidates {
		for featName := range expectedFeatures {
			if _, exists := cand.Attributes[featName]; !exists {
				gaps = append(gaps, InformationGap{
					CandidateID:      cand.ID,
					MissingAttribute: featName,
					Reason:           fmt.Sprintf("candidate %s missing evaluation attribute '%s'", cand.ID, featName),
					ImpactOnChoice:   "CRITICAL_UNCERTAINTY",
					TargetProvider:   "UNDERSTANDING",
				})
			}
		}
	}

	return gaps
}
