package reasoning

import (
	"context"
	"encoding/json"
	"fmt"

	"idun/core/memory"
	"idun/intelligence/communication"
	"idun/intelligence/understanding"
)

// SymbolicRule defines a structured production rule for the S1 Symbolic Fast Path.
// It maps antecedent conditions (intent and optional slot operators) to a consequent SemanticGoal.
type SymbolicRule struct {
	RuleID        string                 `json:"rule_id"`
	Intent        string                 `json:"intent"`
	RequiredSlots []CompilationCondition `json:"required_slots,omitempty"`
	Consequent    *SemanticGoal          `json:"consequent"`
}

// matchRule checks if a rule matches the given intent and frame slots.
func matchRule(rule SymbolicRule, intent string, slots []understanding.Slot) bool {
	if rule.Intent != intent || rule.Consequent == nil {
		return false
	}
	if len(rule.RequiredSlots) == 0 {
		return true
	}
	slotMap := make(map[string]string, len(slots))
	for _, s := range slots {
		slotMap[s.Name] = s.Value
	}
	for _, cond := range rule.RequiredSlots {
		val, exists := slotMap[cond.Slot]
		if !exists {
			return false
		}
		switch cond.Operator {
		case "==", "eq":
			if val != cond.Value {
				return false
			}
		case "!=", "neq":
			if val == cond.Value {
				return false
			}
		}
	}
	return true
}

// SymbolicSpecialist implements Stage S1 Symbolic Fast Path.
// It forward-chains hardcoded and knowledge-rule production rules over structured SemanticFrames
// and retrieved Memory context to resolve canonical inference patterns rapidly (<2 ms).
type SymbolicSpecialist struct {
	bootstrapRules []SymbolicRule
}

// NewSymbolicSpecialist constructs an initialized SymbolicSpecialist with baseline/bootstrap rules.
func NewSymbolicSpecialist() *SymbolicSpecialist {
	return &SymbolicSpecialist{
		bootstrapRules: []SymbolicRule{
			{
				RuleID: "rule-dialogue-greet",
				Intent: "greet_user",
				Consequent: &SemanticGoal{
					Kind:          GoalKindCommunicative,
					Intent:        "greet_user",
					Target:        "user",
					DesiredState:  map[string]string{"acknowledged": "true"},
					OperationHint: "greeting",
				},
			},
			{
				RuleID: "rule-dialogue-identity",
				Intent: "query_identity",
				Consequent: &SemanticGoal{
					Kind:          GoalKindCommunicative,
					Intent:        "query_identity",
					Target:        "system_identity",
					DesiredState:  map[string]string{"identity_communicated": "true"},
					OperationHint: "identity",
				},
			},
			{
				RuleID: "rule-dialogue-wellbeing",
				Intent: "query_wellbeing",
				Consequent: &SemanticGoal{
					Kind:          GoalKindCommunicative,
					Intent:        "query_wellbeing",
					Target:        "system_status",
					DesiredState:  map[string]string{"status_communicated": "true"},
					OperationHint: "wellbeing",
				},
			},
			{
				RuleID: "rule-dialogue-farewell",
				Intent: "farewell_user",
				Consequent: &SemanticGoal{
					Kind:          GoalKindCommunicative,
					Intent:        "farewell_user",
					Target:        "user",
					DesiredState:  map[string]string{"session_concluded": "true"},
					OperationHint: "farewell",
				},
			},
			{
				RuleID: "rule-dialogue-unresolved",
				Intent: "unresolved_intent",
				Consequent: &SemanticGoal{
					Kind:          GoalKindCommunicative,
					Intent:        "unresolved_intent",
					Target:        "user",
					DesiredState:  map[string]string{"clarification_requested": "true"},
					OperationHint: "acknowledge_unclear",
				},
			},
		},
	}
}

// ID returns the cascade stage identifier for SymbolicSpecialist.
func (s *SymbolicSpecialist) ID() StageIdentifier {
	return StageS1SymbolicFast
}

// Evaluate applies symbolic rules to derive candidate ReasoningHypothesis conclusions and proposed goals.
func (s *SymbolicSpecialist) Evaluate(
	ctx context.Context,
	perceptionEnv communication.Envelope,
	frame *understanding.SemanticFrame,
	memContext []memory.Record,
) ([]ReasoningHypothesis, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var conclusion string
	var confidence float64
	var proposedGoal *SemanticGoal
	supportingPremises := make([]string, 0, len(memContext)+2)

	intent := ""
	var slots []understanding.Slot
	if frame != nil {
		if len(frame.Intents) > 0 {
			intent = frame.Intents[0].Intent
			slots = frame.PrimaryHypothesis.Slots
			conclusion = fmt.Sprintf("Derived symbolic conclusion for semantic intent graph root %q", intent)
			confidence = frame.PrimaryHypothesis.CalibratedConfidence
			if confidence <= 0.0 {
				confidence = 0.90
			}
			for _, slot := range slots {
				supportingPremises = append(supportingPremises, fmt.Sprintf("slot:%s=%s", slot.Name, slot.Value))
			}
			if frame.ValidationTrace != nil {
				supportingPremises = append(supportingPremises, fmt.Sprintf("validation_trace:Type=%s,Constraint=%s", frame.ValidationTrace.TypeValidation, frame.ValidationTrace.ConstraintValidation))
			}
		} else if frame.PrimaryHypothesis.Intent != "" {
			intent = frame.PrimaryHypothesis.Intent
			slots = frame.PrimaryHypothesis.Slots
			conclusion = fmt.Sprintf("Derived symbolic conclusion for intent %q", intent)
			confidence = frame.PrimaryHypothesis.CalibratedConfidence
			if confidence <= 0.0 {
				confidence = 0.90
			}
			for _, slot := range slots {
				supportingPremises = append(supportingPremises, fmt.Sprintf("slot:%s=%s", slot.Name, slot.Value))
			}
		}
	}

	if intent == "" {
		conclusion = fmt.Sprintf("Derived symbolic conclusion for perception envelope %s", perceptionEnv.ID)
		confidence = perceptionEnv.RawConfidence
		if confidence <= 0.0 {
			confidence = 0.85
		}
	}

	// First: Check if any memory record is a knowledge/symbolic rule that matches
	var matchedRule *SymbolicRule
	if intent != "" {
		for _, rec := range memContext {
			if rec.Type == "knowledge_rule" || rec.Type == "rule" || rec.Type == "symbolic_rule" {
				var rule SymbolicRule
				if err := json.Unmarshal(rec.Payload, &rule); err == nil && rule.Consequent != nil {
					if matchRule(rule, intent, slots) {
						matchedRule = &rule
						supportingPremises = append(supportingPremises, fmt.Sprintf("memory:%s(%s)", rec.Type, rec.ID))
						break
					}
				}
				// Also check if stored as CompilationCandidate with DerivedConsequent
				var cc CompilationCandidate
				if err := json.Unmarshal(rec.Payload, &cc); err == nil && cc.DerivedConsequent.Intent == intent {
					var ruleActionGoal SemanticGoal
					if err := json.Unmarshal([]byte(cc.DerivedConsequent.RuleAction), &ruleActionGoal); err == nil && ruleActionGoal.Validate() == nil {
						ruleFromCC := SymbolicRule{
							RuleID:        rec.ID,
							Intent:        intent,
							RequiredSlots: cc.Antecedents,
							Consequent:    &ruleActionGoal,
						}
						if matchRule(ruleFromCC, intent, slots) {
							matchedRule = &ruleFromCC
							supportingPremises = append(supportingPremises, fmt.Sprintf("memory:%s(%s)", rec.Type, rec.ID))
							break
						}
					}
				}
			} else {
				supportingPremises = append(supportingPremises, fmt.Sprintf("memory:%s(%s)", rec.Type, rec.ID))
			}
		}

		// Second: If no memory rule matched, check bootstrap rules
		if matchedRule == nil {
			for _, r := range s.bootstrapRules {
				if matchRule(r, intent, slots) {
					ruleCopy := r
					matchedRule = &ruleCopy
					break
				}
			}
		}

		if matchedRule != nil && matchedRule.Consequent != nil && matchedRule.Consequent.Validate() == nil {
			proposedGoal = matchedRule.Consequent.Clone()
			if proposedGoal.Constraints == nil {
				proposedGoal.Constraints = make(map[string]string)
			}
			for _, slot := range slots {
				proposedGoal.Constraints["slot_"+slot.Name] = slot.Value
			}
			supportingPremises = append(supportingPremises, fmt.Sprintf("matched_rule=%s", matchedRule.RuleID))

			// --- Phase 2C Deliberative Population ---
			proposedGoal.GoalID = fmt.Sprintf("goal-%s", perceptionEnv.ID)
			proposedGoal.Status = StateCreated
			proposedGoal.Confidence = confidence
			proposedGoal.GoalCompleteness = 1.0 // Heuristic: start at 1.0
			proposedGoal.SemanticFrameID = perceptionEnv.ID

			if frame != nil {
				if len(frame.Constraints) > 0 {
					for _, c := range frame.Constraints {
						proposedGoal.SemanticConstraints = append(proposedGoal.SemanticConstraints, Constraint{
							Name:     "semantic_constraint",
							Value:    c,
							Priority: 50,
						})
					}
				}
				if len(frame.Assumptions) > 0 {
					for _, a := range frame.Assumptions {
						proposedGoal.GoalAssumptions = append(proposedGoal.GoalAssumptions, Assumption{
							Category: a.Category,
							Details:  a.Details,
						})
						proposedGoal.AcceptedAssumptions = append(proposedGoal.AcceptedAssumptions, a.Category)
					}
				}
			}
		}
	} else {
		for _, rec := range memContext {
			supportingPremises = append(supportingPremises, fmt.Sprintf("memory:%s(%s)", rec.Type, rec.ID))
		}
	}

	primary := ReasoningHypothesis{
		ID:                  fmt.Sprintf("s1-hyp-%s", perceptionEnv.ID),
		Type:                HypothesisInference,
		Conclusion:          conclusion,
		ProposedGoal:        proposedGoal,
		ReasoningConfidence: confidence,
		ContributingStages:  []StageIdentifier{StageS1SymbolicFast},
		SupportingPremises:  supportingPremises,
		EvidenceTrace:       "Evaluated via S1 Symbolic Fast Path forward-chaining rules",
	}

	return []ReasoningHypothesis{primary}, nil
}
