package decision

import (
	"fmt"
	"math"
)

// ParetoDominates evaluates whether candidate A strictly Pareto-dominates candidate B
// across a set of multi-dimensional feature attributes.
// A Pareto-dominates B if A is at least as good as B in every criterion and strictly better in at least one.
func ParetoDominates(attrsA, attrsB map[string]float64, criteriaKeys []string) bool {
	if len(criteriaKeys) == 0 {
		return false
	}
	strictlyBetter := false

	for _, k := range criteriaKeys {
		valA := attrsA[k]
		valB := attrsB[k]
		// Using epsilon for floating point comparison
		if valA < valB-1e-9 {
			return false // A is worse on criterion k, cannot dominate
		}
		if valA > valB+1e-9 {
			strictlyBetter = true
		}
	}

	return strictlyBetter
}

// FindParetoFrontier analyzes a slice of candidate alternatives and partitions them
// into Pareto-efficient (non-dominated) and Pareto-dominated sets.
func FindParetoFrontier(candidates []Candidate, criteriaKeys []string) ([]Candidate, []Candidate, error) {
	if len(candidates) == 0 {
		return nil, nil, nil
	}
	if len(criteriaKeys) == 0 {
		// Default to all attribute keys present in first candidate
		keyMap := make(map[string]struct{})
		for _, c := range candidates {
			for k := range c.Attributes {
				keyMap[k] = struct{}{}
			}
		}
		for k := range keyMap {
			criteriaKeys = append(criteriaKeys, k)
		}
	}

	efficient := make([]Candidate, 0, len(candidates))
	dominated := make([]Candidate, 0)

	for i, candA := range candidates {
		isDominated := false
		for j, candB := range candidates {
			if i == j {
				continue
			}
			if ParetoDominates(candB.Attributes, candA.Attributes, criteriaKeys) {
				isDominated = true
				break
			}
		}
		if isDominated {
			dominated = append(dominated, candA)
		} else {
			efficient = append(efficient, candA)
		}
	}

	return efficient, dominated, nil
}

// ComputeTradeoffDistance computes the Euclidean distance between two candidates in multi-criteria space.
func ComputeTradeoffDistance(attrsA, attrsB map[string]float64, weights map[string]float64) float64 {
	var sumSq float64
	for k, w := range weights {
		diff := attrsA[k] - attrsB[k]
		sumSq += w * diff * diff
	}
	return math.Sqrt(sumSq)
}

// FormatParetoSummary generates a human-readable and structured summary of the Pareto frontier.
func FormatParetoSummary(efficient []Candidate, totalCandidates int) string {
	return fmt.Sprintf("Pareto frontier identified %d non-dominated candidates out of %d total candidates evaluated", len(efficient), totalCandidates)
}
