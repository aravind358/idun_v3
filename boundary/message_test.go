package boundary_test

import (
	"errors"
	"testing"

	"idun/boundary"
	"idun/intelligence/understanding"
)

func TestCommunicationMessage_ValidateAndRoundTrip(t *testing.T) {
	msg := &boundary.CommunicationMessage{
		ResponseID:  "msg-123",
		ParentRef:   "parent-456",
		Intent:      "confirm_booking",
		DialogueAct: "CONFIRM",
		Meaning:     "Booking DL123 confirmed.",
		Goal:        "Confirm flight booking",
		Confidence:  0.95,
		Tone:        "professional",
		Verbosity:   "brief",
		Language:    "en-US",
		Modality:    "text",
		Slots: []boundary.Slot{
			{Name: "flight", Value: "DL123", GroundingID: "ent-1", Confidence: 0.99},
		},
		Entities: []boundary.Entity{
			{ID: "ent-1", Type: "flight", CanonicalName: "Delta 123"},
		},
		Metadata: map[string]any{
			"trace_id": "trace-999",
		},
		DecisionRef:      "dec-1",
		PlanRef:          "plan-1",
		ReasoningRef:     "reas-1",
		SemanticFrameRef: "frame-1",
	}

	data, err := boundary.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	decoded, err := boundary.Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.ResponseID != msg.ResponseID || decoded.Intent != msg.Intent || decoded.Meaning != msg.Meaning {
		t.Errorf("mismatch in decoded fields: %+v", decoded)
	}
	if len(decoded.Slots) != 1 || decoded.Slots[0].Name != "flight" {
		t.Errorf("mismatch in decoded slots: %+v", decoded.Slots)
	}
	if len(decoded.Entities) != 1 || decoded.Entities[0].ID != "ent-1" {
		t.Errorf("mismatch in decoded entities: %+v", decoded.Entities)
	}
}

func TestCommunicationMessage_ValidationErrors(t *testing.T) {
	var nilMsg *boundary.CommunicationMessage
	if err := nilMsg.Validate(); !errors.Is(err, boundary.ErrNilMessage) {
		t.Errorf("expected ErrNilMessage, got %v", err)
	}

	emptyID := &boundary.CommunicationMessage{ResponseID: ""}
	if err := emptyID.Validate(); !errors.Is(err, boundary.ErrMissingResponseID) {
		t.Errorf("expected ErrMissingResponseID, got %v", err)
	}

	if _, err := boundary.Marshal(emptyID); err == nil {
		t.Errorf("expected Marshal to fail for invalid message")
	}

	if _, err := boundary.Unmarshal([]byte(`{"response_id":""}`)); err == nil {
		t.Errorf("expected Unmarshal to fail for missing response_id")
	}
}

func TestMappingFromUnderstanding(t *testing.T) {
	uSlots := []understanding.Slot{
		{Name: "city", Value: "Tokyo", GroundingID: "geo-tokyo", Confidence: 0.98},
	}

	bSlots := boundary.MapSlotsFromUnderstanding(uSlots)
	if len(bSlots) != 1 {
		t.Fatalf("expected 1 slot, got %d", len(bSlots))
	}
	if bSlots[0].Name != "city" || bSlots[0].Value != "Tokyo" || bSlots[0].GroundingID != "geo-tokyo" || bSlots[0].Confidence != 0.98 {
		t.Errorf("mapped slot mismatch: %+v", bSlots[0])
	}

	if boundary.MapSlotsFromUnderstanding(nil) != nil {
		t.Errorf("expected nil when mapping nil slots")
	}
}
