package network

import (
	"context"
	"testing"

	"idun/capabilities"
)

func TestMetadata(t *testing.T) {
	meta := Metadata()
	if meta.Name != "NativeNetworkCapability" {
		t.Errorf("Expected NativeNetworkCapability, got %s", meta.Name)
	}
	if meta.Category != capabilities.CategoryNetwork {
		t.Errorf("Expected CategoryNetwork, got %s", meta.Category)
	}
}

func TestExecuteResolveDNS(t *testing.T) {
	mockProvider := NewMockProvider(false)
	cap := New(nil, mockProvider)

	req := capabilities.CapabilityRequest{
		RequirementID: "req-dns",
		Parameters: map[string]string{
			"operation": string(OperationResolveDNS),
			"hostname":  "example.com",
		},
	}

	res, err := cap.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected nil err, got %v", err)
	}
	if !res.Success {
		t.Fatalf("Expected Success=true, got error: %v", res.Error)
	}
	
	ips := res.Data["ips"].([]string)
	if len(ips) == 0 || ips[0] != "192.168.1.1" {
		t.Errorf("Expected mocked IP 192.168.1.1")
	}
	
	metrics := cap.Metrics()
	if metrics.ExecutionCount != 1 || metrics.SuccessCount != 1 {
		t.Errorf("Metrics mismatch")
	}
}

func TestExecuteHTTPGet(t *testing.T) {
	mockProvider := NewMockProvider(false)
	cap := New(nil, mockProvider)

	req := capabilities.CapabilityRequest{
		RequirementID: "req-http",
		Parameters: map[string]string{
			"operation": string(OperationHTTPGet),
			"url":       "http://example.com",
		},
	}

	res, err := cap.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected nil err return, got %v", err)
	}
	if !res.Success {
		t.Fatalf("Expected Success=true")
	}
	if res.Data["status_code"].(int) != 200 {
		t.Errorf("Expected status 200")
	}
}

func TestProviderFailure(t *testing.T) {
	mockProvider := NewMockProvider(true) // Configured to fail
	cap := New(nil, mockProvider)

	req := capabilities.CapabilityRequest{
		RequirementID: "req-fail",
		Parameters: map[string]string{
			"operation": string(OperationDownload),
			"url":       "http://example.com/file",
			"destination": "/tmp/out",
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

func TestInvalidOperation(t *testing.T) {
	mockProvider := NewMockProvider(false)
	cap := New(nil, mockProvider)

	req := capabilities.CapabilityRequest{
		RequirementID: "req-invalid",
		Parameters: map[string]string{
			"operation": "ftp_upload",
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
