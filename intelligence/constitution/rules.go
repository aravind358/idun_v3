package constitution

import (
	"context"
	"fmt"

	"idun/intelligence/communication"
)

// FunctionalRule wraps an arbitrary evaluation function into a Rule.
type FunctionalRule struct {
	id          string
	description string
	fn          func(ctx context.Context, env communication.Envelope) (Verdict, string, error)
}

// NewFunctionalRule constructs a new functional rule.
func NewFunctionalRule(id, description string, fn func(ctx context.Context, env communication.Envelope) (Verdict, string, error)) *FunctionalRule {
	return &FunctionalRule{
		id:          id,
		description: description,
		fn:          fn,
	}
}

func (r *FunctionalRule) ID() string {
	return r.id
}

func (r *FunctionalRule) Description() string {
	return r.description
}

func (r *FunctionalRule) Evaluate(ctx context.Context, env communication.Envelope) (Verdict, string, error) {
	if r.fn == nil {
		return VerdictApproved, "", nil
	}
	return r.fn(ctx, env)
}

// NewSafetyMetadataRule creates an invariant rule verifying that candidate action envelopes
// have valid structural metadata before physical execution.
func NewSafetyMetadataRule() Rule {
	return NewFunctionalRule(
		"core.safety.metadata",
		"Verifies candidate action envelopes possess valid structural metadata and provenance.",
		func(ctx context.Context, env communication.Envelope) (Verdict, string, error) {
			if err := env.Validate(); err != nil {
				return VerdictVetoed, fmt.Sprintf("invalid envelope metadata: %v", err), nil
			}
			return VerdictApproved, "", nil
		},
	)
}

// NewMaxUrgencyEscalationRule creates an invariant rule that escalates actions with
// extreme urgency thresholds to the user before irreversible execution.
func NewMaxUrgencyEscalationRule(threshold int) Rule {
	return NewFunctionalRule(
		"core.safety.urgency-escalation",
		fmt.Sprintf("Escalates irreversible actions with urgency >= %d for user confirmation.", threshold),
		func(ctx context.Context, env communication.Envelope) (Verdict, string, error) {
			if env.Urgency >= threshold {
				return VerdictEscalateToUser, fmt.Sprintf("action urgency %d >= escalation threshold %d", env.Urgency, threshold), nil
			}
			return VerdictApproved, "", nil
		},
	)
}
