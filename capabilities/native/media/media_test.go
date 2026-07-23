package media

import (
	"context"
	"testing"

	"idun/capabilities"
)

func TestMetadata(t *testing.T) {
	meta := Metadata()
	if meta.Name != "NativeMediaCapability" {
		t.Errorf("Expected NativeMediaCapability, got %s", meta.Name)
	}
	if meta.Category != capabilities.CategoryMedia {
		t.Errorf("Expected CategoryMedia, got %s", meta.Category)
	}
}

func TestExecutePlayAudio(t *testing.T) {
	mockProvider := NewMockProvider(false)
	cap := New(nil, mockProvider)

	req := capabilities.CapabilityRequest{
		RequirementID: "req-audio-play",
		Parameters: map[string]string{
			"operation": string(OperationPlayAudio),
			"path":      "/path/to/audio.mp3",
		},
	}

	res, err := cap.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected nil err, got %v", err)
	}
	if !res.Success {
		t.Fatalf("Expected Success=true, got error: %v", res.Error)
	}
	
	if res.Data["session_id"] != "audio-session-1" {
		t.Errorf("Expected audio-session-1, got %v", res.Data["session_id"])
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
			"operation": "synthesize_speech",
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
			"operation": string(OperationGetMetadata),
			"path":      "/path/to/video.mp4",
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
