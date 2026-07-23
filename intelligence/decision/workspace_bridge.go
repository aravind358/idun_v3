package decision

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"idun/boundary"
	"idun/intelligence/communication"
)

// ShouldPublishToWorkspace determines whether a decision record should be broadcast
// over the Global Workspace communication substrate.
// Per Section 4.1 of Version 2.0.0-FROZEN:
// - Reflexive micro-decisions SKIP workspace broadcast to prevent bus saturation and deadlock.
// - Deliberative macro-decisions MUST be published to TopicEvaluatedOptions.
func ShouldPublishToWorkspace(depth DeliberationDepth) bool {
	return depth == DepthDeliberative
}

// WorkspacePublisher defines the functional interface required to emit Envelopes to the Global Workspace.
type WorkspacePublisher interface {
	Publish(ctx context.Context, env communication.Envelope) error
}

// WorkspaceSubscription manages the lifecycle of an active subscription.
type WorkspaceSubscription interface {
	Cancel() error
}

// WorkspaceSubscriber defines the interface for registering handlers on Workspace topics.
type WorkspaceSubscriber interface {
	Subscribe(topic communication.TopicID, subscriberID string, handler func(ctx context.Context, env communication.Envelope) error) (WorkspaceSubscription, error)
}

// PayloadStorer defines the functional interface required to persist DecisionRecord payloads to CAS storage.
type PayloadStorer interface {
	Store(ctx context.Context, data []byte) (string, error)
	Retrieve(ctx context.Context, key string) ([]byte, error)
}



// PublishDeliberativeDecision serializes a Deliberative DecisionRecord (or boundary.CommunicationMessage if selectedCand is provided),
// stores its payload reference, packages it into a canonical communication.Envelope, and publishes it to TopicEvaluatedOptions.
func PublishDeliberativeDecision(
	ctx context.Context,
	rec *DecisionRecord,
	selectedCand *Candidate,
	storer PayloadStorer,
	publisher WorkspacePublisher,
	parentRefs ...string,
) (communication.Envelope, error) {
	if rec == nil {
		return communication.Envelope{}, ErrInvalidDecisionRecord
	}
	if err := rec.Validate(); err != nil {
		return communication.Envelope{}, fmt.Errorf("decision validation firewall rejected record: %w", err)
	}
	if !ShouldPublishToWorkspace(rec.DeliberationDepth) {
		return communication.Envelope{}, fmt.Errorf("decision: reflexive decision records must not be published to Global Workspace")
	}
	if storer == nil || publisher == nil {
		return communication.Envelope{}, fmt.Errorf("decision: storer and publisher cannot be nil")
	}

	var data []byte
	var err error
	if selectedCand != nil {
		parentID := ""
		if len(parentRefs) > 0 && parentRefs[0] != "" {
			parentID = parentRefs[0]
		}

		commMsg := &boundary.CommunicationMessage{
			ResponseID:  "resp-" + rec.DecisionID,
			ParentRef:   parentID,
			Goal:        selectedCand.Description,
			Meaning:     selectedCand.Description,
			Tone:        "conversational",
			Verbosity:   "standard",
			Language:    "en-US",
			Modality:    "text",
			Confidence:  rec.Confidence,
			DecisionRef: rec.DecisionID,
		}
		if parentID != "" {
			commMsg.ReasoningRef = parentID
		}
		
		if selectedCand.Metadata != nil {
			if rgStr, ok := selectedCand.Metadata["resolved_goal"]; ok {
				var rg struct {
					Intent      string            `json:"intent"`
					Target      string            `json:"target"`
					Constraints map[string]string `json:"constraints"`
				}
				if json.Unmarshal([]byte(rgStr), &rg) == nil {
					commMsg.Intent = rg.Intent
					for k, v := range rg.Constraints {
						if strings.HasPrefix(k, "slot_") {
							commMsg.Slots = append(commMsg.Slots, boundary.Slot{
								Name:  strings.TrimPrefix(k, "slot_"),
								Value: v,
							})
						}
					}
				}
			}
			if pdStr, ok := selectedCand.Metadata["presentation_directives"]; ok {
				var pd struct {
					Tone      string `json:"tone"`
					Verbosity string `json:"verbosity"`
					Language  string `json:"language"`
				}
				if json.Unmarshal([]byte(pdStr), &pd) == nil {
					if pd.Tone != "" {
						commMsg.Tone = pd.Tone
					}
					if pd.Verbosity != "" {
						commMsg.Verbosity = pd.Verbosity
					}
					if pd.Language != "" {
						commMsg.Language = pd.Language
					}
				}
			}
		}

		if commMsg.Intent == "" || commMsg.Intent == "inform" {
			if idx := strings.Index(selectedCand.Description, "for intent \""); idx != -1 {
				sub := selectedCand.Description[idx+len("for intent \""):]
				if endIdx := strings.Index(sub, "\""); endIdx != -1 {
					commMsg.Intent = sub[:endIdx]
				}
			}
		}
		if commMsg.Intent == "" {
			commMsg.Intent = "inform"
		}
		if commMsg.DialogueAct == "" {
			commMsg.DialogueAct = "INFORM"
		}
		if commMsg.Meaning == "" {
			commMsg.Meaning = selectedCand.Description
		}

		data, err = boundary.Marshal(commMsg)
		if err != nil {
			return communication.Envelope{}, fmt.Errorf("decision: failed to marshal communication message: %w", err)
		}
	} else {
		data, err = MarshalDecisionRecord(rec)
		if err != nil {
			return communication.Envelope{}, err
		}
	}

	payloadRef, err := storer.Store(ctx, data)
	if err != nil {
		return communication.Envelope{}, fmt.Errorf("decision: failed to store payload: %w", err)
	}

	env, err := EnvelopeFromDecisionRecord(rec, payloadRef)
	if err != nil {
		return communication.Envelope{}, err
	}
	if selectedCand != nil {
		env.PayloadModality = "communication-message"
	}

	if len(parentRefs) > 0 && parentRefs[0] != "" {
		env.ParentRef = parentRefs[0]
	}

	if err := publisher.Publish(ctx, env); err != nil {
		return communication.Envelope{}, fmt.Errorf("decision: failed to publish decision record to workspace: %w", err)
	}

	return env, nil
}
