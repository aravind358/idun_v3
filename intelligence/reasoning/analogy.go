package reasoning

import (
	"context"
	"fmt"
	"strings"

	"idun/core/memory"
	"idun/intelligence/communication"
	"idun/intelligence/infrastructure/embedding"
	"idun/intelligence/understanding"
)

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
			hyps = append(hyps, ReasoningHypothesis{
				ID:                  fmt.Sprintf("s5-hyp-%s-%d", perceptionEnv.ID, i),
				Type:                HypothesisAnalogy,
				Conclusion:          conclusion,
				ReasoningConfidence: simScore,
				ContributingStages:  []StageIdentifier{StageS5CaseAnalogy},
				SupportingPremises:   []string{fmt.Sprintf("analogical_case=%s", rec.ID)},
				EvidenceTrace:        fmt.Sprintf("Retrieved via S5 Case-Based Analogy (score=%.2f)", simScore),
			})
		}
	}

	return hyps, nil
}
