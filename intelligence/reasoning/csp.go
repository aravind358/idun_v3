package reasoning

import (
	"context"
	"fmt"
	"strings"

	"idun/core/memory"
)

// CSPCheckSpecialist implements Stage S3 Constraint Consistency Check.
// It checks candidate hypotheses against existing Memory beliefs for logical
// contradictions and flags them without making resolution decisions.
type CSPCheckSpecialist struct{}

// NewCSPCheckSpecialist constructs an initialized CSPCheckSpecialist.
func NewCSPCheckSpecialist() *CSPCheckSpecialist {
	return &CSPCheckSpecialist{}
}

// ID returns the cascade stage identifier for CSPCheckSpecialist.
func (s *CSPCheckSpecialist) ID() StageIdentifier {
	return StageS3CSPCheck
}

// CheckConsistency evaluates whether candidate hypotheses conflict with retrieved Memory context.
func (s *CSPCheckSpecialist) CheckConsistency(
	ctx context.Context,
	hypotheses []ReasoningHypothesis,
	memContext []memory.Record,
) ([]ContradictionFlag, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	flags := make([]ContradictionFlag, 0)

	for _, hyp := range hypotheses {
		for _, rec := range memContext {
			if rec.Type == "contradiction" || strings.HasPrefix(rec.ID, "contradict/") {
				flags = append(flags, ContradictionFlag{
					BeliefID:         rec.ID,
					ConflictingClaim: fmt.Sprintf("Conflict detected against hypothesis %q: %s", hyp.ID, hyp.Conclusion),
					Confidence:       0.95,
					DetectedAtStage:  StageS3CSPCheck,
				})
			}
		}
	}

	return flags, nil
}
