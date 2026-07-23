package template

import (
	"context"
	"testing"

	"idun/capabilities"
)

func TestMetadata(t *testing.T) {
	meta := Metadata()
	if meta.Name == "" {
		t.Error("Expected Capability Metadata Name to be populated")
	}
}

func TestExecuteExample(t *testing.T) {
	mockProvider := NewMockProvider(false)
	cap := New(nil, mockProvider)

	req := capabilities.CapabilityRequest{
		RequirementID: "req-1",
		Parameters: map[string]string{
			"operation": string(OperationExample),
		},
	}

	res, err := cap.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected nil err from Execute, got %v", err)
	}
	if !res.Success {
		t.Errorf("Expected Success=true, got Error=%v", res.Error)
	}
	if res.Data["provider"] != "mock" {
		t.Errorf("Expected Data['provider'] to be mock, got %v", res.Data["provider"])
	}
	
	// Test Metrics
	metrics := cap.Metrics()
	if metrics.ExecutionCount != 1 || metrics.SuccessCount != 1 {
		t.Errorf("Expected ExecutionCount=1 and SuccessCount=1, got %d and %d", metrics.ExecutionCount, metrics.SuccessCount)
	}
}

func TestExecuteInvalidOperation(t *testing.T) {
	mockProvider := NewMockProvider(false)
	cap := New(nil, mockProvider)

	req := capabilities.CapabilityRequest{
		RequirementID: "req-2",
		Parameters: map[string]string{
			"operation": "invalid_op",
		},
	}

	res, err := cap.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected nil error return, got %v", err)
	}
	if res.Success {
		t.Error("Expected Success=false")
	}
	if res.Error == nil || res.Error.Code != "Validation" {
		t.Errorf("Expected Validation error, got %v", res.Error)
	}
}

func TestExecuteProviderFailure(t *testing.T) {
	mockProvider := NewMockProvider(true) // Configured to fail
	cap := New(nil, mockProvider)

	req := capabilities.CapabilityRequest{
		RequirementID: "req-3",
		Parameters: map[string]string{
			"operation": string(OperationExample),
		},
	}

	res, err := cap.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected nil error return, got %v", err)
	}
	if res.Success {
		t.Error("Expected Success=false")
	}
	if res.Error == nil || res.Error.Code != "Execution" {
		t.Errorf("Expected Execution error, got %v", res.Error)
	}
}
