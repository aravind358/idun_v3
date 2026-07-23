package capabilities

import (
	"context"
	"encoding/json"
	"fmt"

	"idun/intelligence/communication"
	"idun/intelligence/planning"
)

// PayloadRetriever defines the interface to fetch payloads from CAS.
type PayloadRetriever interface {
	Retrieve(ctx context.Context, key string) ([]byte, error)
}

// WorkspaceSubscriber defines the interface for registering topic handlers.
type WorkspaceSubscriber interface {
	Subscribe(topic communication.TopicID, subscriberID string, handler func(ctx context.Context, env communication.Envelope) error) (interface{}, error)
}

// ActionExecutionHandler consumes TopicActionExecution, extracts CapabilityRequirements from CandidatePlans,
// resolves them, and executes them through the Capability Framework.
type ActionExecutionHandler struct {
	manager   CapabilityManager
	retriever PayloadRetriever
}

func NewActionExecutionHandler(manager CapabilityManager, retriever PayloadRetriever) *ActionExecutionHandler {
	return &ActionExecutionHandler{
		manager:   manager,
		retriever: retriever,
	}
}

// HandleActionExecution processes the execution envelope.
func (h *ActionExecutionHandler) HandleActionExecution(ctx context.Context, env communication.Envelope) error {
	if env.Topic != communication.TopicActionExecution {
		return fmt.Errorf("capability_framework: expected TopicActionExecution, got %s", env.Topic)
	}
	if env.PayloadRef == "" {
		return fmt.Errorf("capability_framework: missing payload reference in envelope")
	}

	data, err := h.retriever.Retrieve(ctx, env.PayloadRef)
	if err != nil {
		return fmt.Errorf("capability_framework: failed to retrieve payload: %w", err)
	}

	// In Phase 2D, the payload is a CandidatePlan or DecisionRecord containing a CandidatePlan.
	// We parse it as a basic structure to find the Plan and its ExecutionSteps.
	var plan planning.CandidatePlan
	
	// Attempt to unmarshal as CandidatePlan
	if err := json.Unmarshal(data, &plan); err != nil {
		// If it's a DecisionRecord, extract the CandidatePlan (simplified for V1 integration)
		return fmt.Errorf("capability_framework: unable to parse CandidatePlan payload: %v", err)
	}

	for _, step := range plan.ExecutionSteps {
		if step.Action == "EXECUTE_CAPABILITY" {
			// Extract capability requirements from the plan
			var capabilityName string
			var params map[string]string
			for _, req := range plan.CapabilityRequirements {
				if req.RequirementID == step.CapabilityReqID {
					capabilityName = req.CapabilityName
					params = req.Parameters
					break
				}
			}

			if capabilityName == "" {
				continue
			}

			// Resolve the capability
			cap, err := h.manager.Resolver().Resolve(ctx, step.CapabilityReqID, capabilityName, params)
			if err != nil {
				// Record error but continue attempting other steps if parallel (simplified for V1)
				fmt.Printf("Capability resolution failed for %s: %v\n", capabilityName, err)
				continue
			}

			// Execute the capability
			req := CapabilityRequest{
				RequirementID: step.CapabilityReqID,
				Parameters:    params,
				ContextID:     env.ID,
			}
			res, err := cap.Execute(ctx, req)
			if err != nil {
				fmt.Printf("Capability execution failed for %s: %v\n", capabilityName, err)
				continue
			}

			// Normalize the result
			// For V1, we just print or log it. The result would typically be pushed to TopicActionResults.
			fmt.Printf("Capability execution succeeded for %s: %v\n", capabilityName, res.Success)
		}
	}

	return nil
}
