package files

import (
	"context"
	"testing"

	"idun/capabilities"
)

func TestMetadata(t *testing.T) {
	meta := Metadata()
	if meta.Name != "NativeFilesCapability" {
		t.Errorf("Expected NativeFilesCapability, got %s", meta.Name)
	}
	if meta.Category != capabilities.CategoryFiles {
		t.Errorf("Expected CategoryFiles, got %s", meta.Category)
	}
}

func TestExecuteReadFile(t *testing.T) {
	mockProvider := NewMockProvider(false)
	cap := New(nil, mockProvider)

	req := capabilities.CapabilityRequest{
		RequirementID: "req-read",
		Parameters: map[string]string{
			"operation": string(OperationReadText),
			"path":      "/fake/path.txt",
		},
	}

	res, err := cap.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected nil err, got %v", err)
	}
	if !res.Success {
		t.Fatalf("Expected Success=true, got error: %v", res.Error)
	}
	
	if res.Data["text"] != "mock data" {
		t.Errorf("Expected mock data, got %v", res.Data["text"])
	}
	
	metrics := cap.Metrics()
	if metrics.ExecutionCount != 1 || metrics.SuccessCount != 1 {
		t.Errorf("Metrics mismatch")
	}
}

func TestExecuteWriteFile(t *testing.T) {
	mockProvider := NewMockProvider(false)
	cap := New(nil, mockProvider)

	req := capabilities.CapabilityRequest{
		RequirementID: "req-write",
		Parameters: map[string]string{
			"operation": string(OperationWriteFile),
			"path":      "/fake/path.txt",
			"data_text": "hello",
		},
	}

	res, err := cap.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected nil err, got %v", err)
	}
	if !res.Success {
		t.Fatalf("Expected Success=true")
	}
}

func TestExecuteInvalidOperation(t *testing.T) {
	mockProvider := NewMockProvider(false)
	cap := New(nil, mockProvider)

	req := capabilities.CapabilityRequest{
		RequirementID: "req-invalid",
		Parameters: map[string]string{
			"operation": "delete_everything",
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
	mockProvider := NewMockProvider(true) // fails
	cap := New(nil, mockProvider)

	req := capabilities.CapabilityRequest{
		RequirementID: "req-fail",
		Parameters: map[string]string{
			"operation": string(OperationReadFile),
			"path":      "/fake/path.txt",
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
