package reasoning

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"idun/intelligence/communication"
	"idun/intelligence/infrastructure/inference"
)

// DeliberativeSpecialist implements Stage S8 Deliberative LLM Reasoning.
// It integrates shared idun/intelligence/infrastructure/inference to escalate
// complex semantic reasoning when deterministic stages lack confidence.
//
// MANDATORY INVARIANTS:
// 1. Reuses shared InferenceService only; NEVER creates custom LLM clients.
// 2. Escalates ONLY when primary ReasoningConfidence falls below EscalationThreshold.
// 3. Emits structured ReasoningHypothesis only; NEVER injects raw free-form prose into Workspace.
type DeliberativeSpecialist struct {
	infer inference.InferenceService
}

// NewDeliberativeSpecialist returns an initialized DeliberativeSpecialist.
func NewDeliberativeSpecialist(infer inference.InferenceService) *DeliberativeSpecialist {
	return &DeliberativeSpecialist{infer: infer}
}

// ID returns StageS8DeliberativeLLM.
func (s *DeliberativeSpecialist) ID() StageIdentifier {
	return StageS8DeliberativeLLM
}

// EvaluateDeliberative escalates to shared InferenceService when confidence is below threshold.
func (s *DeliberativeSpecialist) EvaluateDeliberative(
	ctx context.Context,
	perceptionEnv communication.Envelope,
	currentConf float64,
	escalationThreshold float64,
	storer ...PayloadStorer,
) ([]ReasoningHypothesis, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if escalationThreshold <= 0.0 {
		escalationThreshold = 0.65
	}
	// Do not escalate if confidence already satisfies threshold
	if currentConf >= escalationThreshold {
		return nil, nil
	}
	if s.infer == nil {
		return nil, nil
	}

	req := inference.InferenceRequest{
		ModelID:  "reasoning-deliberative-llm",
		InputRef: perceptionEnv.PayloadRef,
		Modality: inference.ModalityStructured,
		Budget:   "DELIBERATIVE",
		CallerID: "Intelligence.Reasoning",
	}

	res, err := s.infer.Execute(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("stage S8 deliberative inference failed: %w", err)
	}

	var rawData []byte
	if len(storer) > 0 && storer[0] != nil {
		if retrieved, rErr := storer[0].Retrieve(ctx, res.OutputRef); rErr == nil && len(retrieved) > 0 {
			rawData = retrieved
		}
	}
	if len(rawData) == 0 && len(res.OutputRef) > 0 {
		trimmed := strings.TrimSpace(res.OutputRef)
		if strings.HasPrefix(trimmed, "{") {
			rawData = []byte(trimmed)
		}
	}

	var proposedGoal *SemanticGoal
	var rejectionStatus string
	if len(rawData) > 0 {
		var candidate SemanticGoal
		if err := json.Unmarshal(rawData, &candidate); err == nil && candidate.Validate() == nil {
			proposedGoal = candidate.Clone()
		} else {
			rejectionStatus = "rejected_malformed_goal"
		}
	}

	premises := []string{fmt.Sprintf("inference_ref=%s", res.OutputRef)}
	if proposedGoal != nil {
		premises = append(premises, "provenance=LLM_FALLBACK")
	} else if rejectionStatus != "" {
		premises = append(premises, fmt.Sprintf("status=%s", rejectionStatus))
	}

	hyp := ReasoningHypothesis{
		ID:                  fmt.Sprintf("s8-hyp-%s", perceptionEnv.ID),
		Type:                HypothesisDeliberative,
		Conclusion:          fmt.Sprintf("Deliberative structured synthesis for %s via %s", perceptionEnv.ID, res.OutputRef),
		ProposedGoal:        proposedGoal,
		ReasoningConfidence: 0.85,
		ContributingStages:  []StageIdentifier{StageS8DeliberativeLLM},
		SupportingPremises:  premises,
		EvidenceTrace:       fmt.Sprintf("Escalated to Stage S8 Deliberative LLM (model=%s, compute=%d)", res.ModelID, res.ComputeUnits),
	}

	if err := hyp.Validate(); err != nil {
		return nil, fmt.Errorf("stage S8 generated hypothesis schema invalid: %w", err)
	}

	return []ReasoningHypothesis{hyp}, nil
}
