package system

import (
	"context"
	"testing"

	"idun/capabilities"
)

func TestMetadata(t *testing.T) {
	meta := Metadata()
	if meta.Category != capabilities.CategorySystem {
		t.Errorf("Expected CategorySystem, got %s", meta.Category)
	}
	if meta.Name != "NativeSystemCapability" {
		t.Errorf("Expected NativeSystemCapability, got %s", meta.Name)
	}
	if meta.Version != "3.1.2" {
		t.Errorf("Expected Version 3.1.2, got %s", meta.Version)
	}
}

func TestExecuteInfo(t *testing.T) {
	mockProvider := NewMockProvider(false)
	cap := New(nil, mockProvider, nil)

	req := capabilities.CapabilityRequest{
		RequirementID: "req-1",
		Parameters: map[string]string{
			"operation": string(OperationSystemInfo),
		},
	}

	res, err := cap.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected nil err from Execute, got %v", err)
	}
	if !res.Success {
		t.Errorf("Expected Success=true, got Error=%v", res.Error)
	}
	if res.Data["os"] != "mock_os" {
		t.Errorf("Expected Data['os'] to be mock_os, got %v", res.Data["os"])
	}
	
	// Test Metrics
	metrics := cap.Metrics()
	if metrics.ExecutionCount != 1 || metrics.SuccessCount != 1 {
		t.Errorf("Expected ExecutionCount=1 and SuccessCount=1, got %d and %d", metrics.ExecutionCount, metrics.SuccessCount)
	}
}

func TestExecutePower(t *testing.T) {
	mockProvider := NewMockProvider(false)
	cap := New(nil, mockProvider, nil)

	req := capabilities.CapabilityRequest{
		RequirementID: "req-2",
		Parameters: map[string]string{
			"operation": string(OperationShutdown),
		},
	}

	res, err := cap.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected nil err from Execute, got %v", err)
	}
	if !res.Success {
		t.Errorf("Expected Success=true, got Error=%v", res.Error)
	}
	if res.Data["action"] != string(OperationShutdown) {
		t.Errorf("Expected action=shutdown, got %v", res.Data["action"])
	}
}

func TestExecuteInvalidOperation(t *testing.T) {
	mockProvider := NewMockProvider(false)
	cap := New(nil, mockProvider, nil)

	req := capabilities.CapabilityRequest{
		RequirementID: "req-3",
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

	// Test Metrics
	metrics := cap.Metrics()
	if metrics.ExecutionCount != 1 || metrics.FailureCount != 1 {
		t.Errorf("Expected ExecutionCount=1 and FailureCount=1, got %d and %d", metrics.ExecutionCount, metrics.FailureCount)
	}
}

func TestExecuteProviderFailure(t *testing.T) {
	mockProvider := NewMockProvider(true) // Configured to fail
	cap := New(nil, mockProvider, nil)

	req := capabilities.CapabilityRequest{
		RequirementID: "req-4",
		Parameters: map[string]string{
			"operation": string(OperationSystemInfo),
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

	// Test Metrics
	metrics := cap.Metrics()
	if metrics.FailureCount != 1 {
		t.Errorf("Expected FailureCount=1, got %d", metrics.FailureCount)
	}
}
