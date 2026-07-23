package communication

import (
	"context"
	"testing"

	"idun/capabilities"
)

func TestMetadata(t *testing.T) {
	meta := Metadata()
	if meta.Name != "NativeCommunicationCapability" {
		t.Errorf("Expected NativeCommunicationCapability, got %s", meta.Name)
	}
	if meta.Category != capabilities.CategoryCommunication {
		t.Errorf("Expected CategoryCommunication, got %s", meta.Category)
	}
}

func TestExecuteSendMessage(t *testing.T) {
	mockProvider := NewMockProvider(false)
	cap := New(nil, mockProvider)

	req := capabilities.CapabilityRequest{
		RequirementID: "req-send",
		Parameters: map[string]string{
			"operation":    string(OperationSendMessage),
			"destination":  "user-123",
			"payload_text": "hello",
		},
	}

	res, err := cap.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected nil err, got %v", err)
	}
	if !res.Success {
		t.Fatalf("Expected Success=true, got error: %v", res.Error)
	}
	
	if res.Data["message_id"] != "msg-12345" {
		t.Errorf("Expected message_id msg-12345, got %v", res.Data["message_id"])
	}
	
	metrics := cap.Metrics()
	if metrics.ExecutionCount != 1 || metrics.SuccessCount != 1 {
		t.Errorf("Metrics mismatch")
	}
}

func TestExecuteInvalidOperation(t *testing.T) {
	mockProvider := NewMockProvider(false)
	cap := New(nil, mockProvider)

	req := capabilities.CapabilityRequest{
		RequirementID: "req-invalid",
		Parameters: map[string]string{
			"operation": "telepathic_broadcast",
		},
	}

	res, err := cap.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected nil err return, got %v", err)
	}
	if res.Success {
		t.Error("Expected Success=false")
	}
	if res.Error == nil || res.Error.Code != "Validation" {
		t.Errorf("Expected Validation error, got %v", res.Error)
	}
}

func TestProviderFailure(t *testing.T) {
	mockProvider := NewMockProvider(true) // Configured to fail
	cap := New(nil, mockProvider)

	req := capabilities.CapabilityRequest{
		RequirementID: "req-fail",
		Parameters: map[string]string{
			"operation":   string(OperationGetHistory),
			"thread_id":   "thread-1",
		},
	}

	res, err := cap.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected nil err return, got %v", err)
	}
	if res.Success {
		t.Error("Expected Success=false")
	}
	if res.Error == nil || res.Error.Code != "Execution" {
		t.Errorf("Expected Execution error, got %v", res.Error)
	}
}
