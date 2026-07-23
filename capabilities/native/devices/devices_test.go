package devices

import (
	"context"
	"testing"

	"idun/capabilities"
)

func TestMetadata(t *testing.T) {
	meta := Metadata()
	if meta.Name != "NativeDevicesCapability" {
		t.Errorf("Expected NativeDevicesCapability, got %s", meta.Name)
	}
	if meta.Category != capabilities.CategoryDevicesSensors {
		t.Errorf("Expected CategoryDevicesSensors, got %s", meta.Category)
	}
}

func TestExecuteListUSBDevices(t *testing.T) {
	mockProvider := NewMockProvider(false)
	cap := New(nil, mockProvider)

	req := capabilities.CapabilityRequest{
		RequirementID: "req-usb-list",
		Parameters: map[string]string{
			"operation": string(OperationListUSBDevices),
		},
	}

	res, err := cap.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected nil err, got %v", err)
	}
	if !res.Success {
		t.Fatalf("Expected Success=true, got error: %v", res.Error)
	}
	
	devices := res.Data["devices"].([]map[string]interface{})
	if len(devices) == 0 || devices[0]["id"] != "usb-1" {
		t.Errorf("Expected mock usb device")
	}
	
	metrics := cap.Metrics()
	if metrics.ExecutionCount != 1 || metrics.SuccessCount != 1 {
		t.Errorf("Metrics mismatch")
	}
}

func TestExecuteBatteryStatus(t *testing.T) {
	mockProvider := NewMockProvider(false)
	cap := New(nil, mockProvider)

	req := capabilities.CapabilityRequest{
		RequirementID: "req-battery",
		Parameters: map[string]string{
			"operation": string(OperationBatteryStatus),
		},
	}

	res, err := cap.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected nil err return, got %v", err)
	}
	if !res.Success {
		t.Fatalf("Expected Success=true")
	}
	
	status := res.Data["status"].(map[string]interface{})
	if status["percentage"].(int) != 85 {
		t.Errorf("Expected battery percentage 85")
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
			"operation": string(OperationGetGPS),
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
