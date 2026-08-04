package time_test

import (
	"context"
	"testing"
	"time"

	"idun/capabilities"
	nativetime "idun/capabilities/native/time"
	coretime "idun/core/time"
)

func TestTimeCapability_Execute(t *testing.T) {
	timeSvc := coretime.NewTimeService(time.UTC)
	cap := nativetime.New(timeSvc)

	if cap.ID() != "sys-time-1" {
		t.Errorf("Expected ID 'sys-time-1', got %q", cap.ID())
	}

	req := capabilities.CapabilityRequest{
		RequirementID: "req-1",
	}

	res, err := cap.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !res.Success {
		t.Fatalf("Expected success, got false. Error: %v", res.Error)
	}

	data := res.Data
	if data == nil {
		t.Fatalf("Expected Data to be populated")
	}

	if _, exists := data["CurrentTime"]; !exists {
		t.Errorf("Expected CurrentTime in data")
	}

	if tz, exists := data["Timezone"]; !exists || tz != "UTC" {
		t.Errorf("Expected Timezone 'UTC', got %v", tz)
	}
}

func TestTimeCapability_NoTimeService(t *testing.T) {
	cap := nativetime.New(nil)

	req := capabilities.CapabilityRequest{
		RequirementID: "req-2",
	}

	res, err := cap.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if res.Success {
		t.Fatalf("Expected failure due to missing time service")
	}

	if res.Error == nil {
		t.Errorf("Expected error detailing missing time service")
	}
}
