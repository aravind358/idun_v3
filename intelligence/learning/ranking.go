package learning

import (
	"context"
	"fmt"
	"sort"
)

// CandidateRankingEngine provides deterministic ranking of candidate snapshots
// prior to validation gate processing, prioritizing high-impact proposals while preserving reproducibility.
type CandidateRankingEngine interface {
	RankCandidates(ctx context.Context, candidates []*CandidateSnapshot, summary *AggregationSummary, profile *LearningPolicyProfile) ([]*CandidateSnapshot, error)
}

// DefaultCandidateRankingEngine implements CandidateRankingEngine using deterministic scoring
// and stable sorting O(N log N) with tie-breaking on cryptographic snapshot IDs.
type DefaultCandidateRankingEngine struct{}

func NewDefaultCandidateRankingEngine() *DefaultCandidateRankingEngine {
	return &DefaultCandidateRankingEngine{}
}

func (e *DefaultCandidateRankingEngine) RankCandidates(ctx context.Context, candidates []*CandidateSnapshot, summary *AggregationSummary, profile *LearningPolicyProfile) ([]*CandidateSnapshot, error) {
	if len(candidates) <= 1 {
		return candidates, nil
	}

	ranked := make([]*CandidateSnapshot, len(candidates))
	copy(ranked, candidates)

	scores := make(map[*CandidateSnapshot]float64, len(ranked))
	for i, c := range ranked {
		if c == nil {
			return nil, fmt.Errorf("%w: candidate at index %d is nil during ranking", ErrValidationFailed, i)
		}
		var score float64 = 1.0

		// Boost priority if schema matches domain weights in active policy profile
		if profile != nil && profile.DomainWeights != nil {
			if w, ok := profile.DomainWeights[c.SchemaID]; ok {
				score += w
			}
		}

		// Factor in evolutionary generation depth (preference for stable lineage)
		if c.Provenance != nil && c.Provenance.GenerationDepth > 0 {
			score += 1.0 / float64(c.Provenance.GenerationDepth+1)
		} else {
			score += 0.5 // Baseline for root/immediate candidates
		}

		scores[c] = score
	}

	sort.SliceStable(ranked, func(i, j int) bool {
		ci := ranked[i]
		cj := ranked[j]
		si := scores[ci]
		sj := scores[cj]
		if si != sj {
			return si > sj // Higher score ranked first
		}
		// Deterministic tie-breaker using SnapshotID to ensure bit-identical replay
		return ci.SnapshotID < cj.SnapshotID
	})

	return ranked, nil
}
