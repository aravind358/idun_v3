package reasoning

import (
	"context"
	"fmt"

	"idun/intelligence/communication"
	"idun/intelligence/constitution"
)

// ConstitutionSpecialist implements Stage S9 Constitution Integration.
// Every completed ReasoningResult passes through idun/intelligence/constitution
// before broadcast publication.
//
// MANDATORY INVARIANTS:
// 1. Reasoning NEVER evaluates constitutional rules directly.
// 2. Delegates 100% of constitutional evaluation to constitution.ActionGate.
type ConstitutionSpecialist struct {
	gate constitution.ActionGate
}

// NewConstitutionSpecialist returns an initialized ConstitutionSpecialist.
func NewConstitutionSpecialist(gate constitution.ActionGate) *ConstitutionSpecialist {
	return &ConstitutionSpecialist{gate: gate}
}

// ID returns StageS9Constitution.
func (s *ConstitutionSpecialist) ID() StageIdentifier {
	return StageS9Constitution
}

// EvaluateResult submits the candidate ReasoningResult envelope to the Constitutional Action Gate.
func (s *ConstitutionSpecialist) EvaluateResult(ctx context.Context, result *ReasoningResult) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if result == nil {
		return nil
	}
	if s.gate == nil {
		return nil
	}

	env := communication.Envelope{
		ID:              result.EnvelopeID,
		Source:          "intelligence.reasoning",
		Topic:           communication.TopicActiveGoals,
		PayloadRef:      result.EnvelopeID,
		PayloadModality: "reasoning-result",
		RawConfidence:   result.PrimaryHypothesis.CalibratedConfidence,
	}

	eval, err := s.gate.EvaluateAction(ctx, env)
	if err != nil {
		return fmt.Errorf("stage S9 constitutional evaluation failed: %w", err)
	}

	switch eval.Verdict {
	case constitution.VerdictApproved:
		result.ConstitutionAnnotations = append(result.ConstitutionAnnotations, fmt.Sprintf("CONSTITUTION_APPROVED:%s", eval.Signature))
	case constitution.VerdictVetoed:
		result.ConstitutionAnnotations = append(result.ConstitutionAnnotations, fmt.Sprintf("CONSTITUTION_VETOED:%s", eval.Reason))
		return fmt.Errorf("reasoning result vetoed by constitution rule %s: %s", eval.RuleViolated, eval.Reason)
	case constitution.VerdictEscalateToUser:
		result.ConstitutionAnnotations = append(result.ConstitutionAnnotations, "CONSTITUTION_ESCALATE_USER")
	}

	return nil
}
