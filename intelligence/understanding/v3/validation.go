package v3

import (
	"fmt"
)

// Validate enforces the structural and semantic invariants of SemanticInterpretation (V-01 to V-25).
func (s *SemanticInterpretation) Validate() error {
	// V-01: SpecVersion must be the supported version.
	if s.specVersion != SpecVersion {
		return fmt.Errorf("%w: expected %q, got %q", ErrValidation, SpecVersion, s.specVersion)
	}
	// V-02: EnvelopeID must be non-empty.
	if s.envelopeID == "" {
		return fmt.Errorf("%w: EnvelopeID is required", ErrValidation)
	}
	// V-03: Status must be valid.
	switch s.status {
	case StatusUnambiguous, StatusAmbiguous, StatusPreliminary, StatusFailed:
		// valid
	default:
		return fmt.Errorf("%w: invalid Status %q", ErrValidation, s.status)
	}

	// V-04: Confidence must be [0.0, 1.0]
	if s.confidence < 0.0 || s.confidence > 1.0 {
		return fmt.Errorf("%w: Confidence must be between 0.0 and 1.0", ErrValidation)
	}

	// V-06: PrimaryIntent must not be empty.
	if s.primaryIntent == "" {
		return fmt.Errorf("%w: PrimaryIntent is required", ErrValidation)
	}

	// V-11: If FAILED_IMPASSE, PrimaryIntent MUST be "unresolved_intent"
	if s.status == StatusFailed && s.primaryIntent != "unresolved_intent" {
		return fmt.Errorf("%w: FAILED_IMPASSE requires PrimaryIntent to be 'unresolved_intent'", ErrValidation)
	}

	// V-12: CompoundIntentCount must be >= 1
	if s.compoundIntentCount < 1 {
		return fmt.Errorf("%w: CompoundIntentCount must be >= 1", ErrValidation)
	}

	// V-13: SecondaryIntents length must be CompoundIntentCount - 1
	if len(s.secondaryIntents) != s.compoundIntentCount-1 {
		return fmt.Errorf("%w: SecondaryIntents length must be CompoundIntentCount - 1", ErrValidation)
	}

	// V-15: If IsConditional, ConditionClause must not be empty
	if s.isConditional && s.conditionClause == "" {
		return fmt.Errorf("%w: ConditionClause required when IsConditional is true", ErrValidation)
	}
	if !s.isConditional && s.conditionClause != "" {
		return fmt.Errorf("%w: ConditionClause must be empty when IsConditional is false", ErrValidation)
	}

	// V-17: If Negated, NegationMarker must not be empty
	if s.polarity.negated && s.polarity.negationMarker == "" {
		return fmt.Errorf("%w: NegationMarker required when Negated is true", ErrValidation)
	}



	return nil
}
