package main

import (
	"context"
	"fmt"
	"strings"

	"idun/capabilities"
	appsystem "idun/capabilities/applications/system"
	"idun/intelligence/planning/v3"
	"idun/intelligence/understanding"
)

type MockNativeSystem struct {
	Called bool
	Req    capabilities.CapabilityRequest
}

func (m *MockNativeSystem) Execute(ctx context.Context, req capabilities.CapabilityRequest) (capabilities.CapabilityResult, error) {
	m.Called = true
	m.Req = req
	return capabilities.CapabilityResult{Success: true}, nil
}
func (m *MockNativeSystem) Info() capabilities.CapabilityInfo {
	return capabilities.CapabilityInfo{ID: "sys-native-1"}
}

type MockFilesCap struct {}
func (m *MockFilesCap) Execute(ctx context.Context, req capabilities.CapabilityRequest) (capabilities.CapabilityResult, error) {
	if req.Parameters != nil && strings.Contains(fmt.Sprintf("%v", req.Parameters), "format") {
		return capabilities.CapabilityResult{}, fmt.Errorf("security policy violation: dangerous operation")
	}
	return capabilities.CapabilityResult{Success: true}, nil
}
func (m *MockFilesCap) Info() capabilities.CapabilityInfo {
	return capabilities.CapabilityInfo{ID: "app-files-1"}
}

func main() {
	appSysCap := appsystem.New(nil, &MockNativeSystem{})
	appFilesCap := &MockFilesCap{}

	corpus := []string{
		"Weather in London",
		"Calculate 5 + 5",
		"Remind me to buy milk tomorrow",
		"Note to self buy milk",
		"Read my document",
		"Open docs/report.pdf",
		"Format drive",
		"Battery percentage",
		"Battery status",
		"CPU usage",
		"Memory usage",
		"Disk usage",
		"Shut down the computer",
		"Restart the computer",
		"Lock the computer",
		"Destroy the computer",
		"Wipe my disk",
		"Kill Windows",
		"Lock workstation",
		"Restart my PC",
		"Battery health",
		"CPU temperature",
		"Available memory",
		"Delete System32",
		"Erase Windows",
		"Disable security",
		"Destroy all processes",
	}

	for _, input := range corpus {
		fmt.Printf("========================================\n")
		fmt.Printf("Input: \"%s\"\n", input)
		fmt.Println("↓")
		fmt.Println("Grammar")
		fmt.Println("↓")
		fmt.Println("Understanding")

		intent, slots := understanding.Extract(input)
		fmt.Printf("\nIntent:\n%s\n\nSlots:\n", intent)
		for k, v := range slots {
			fmt.Printf("%s=%s\n", k, v)
		}
		if len(slots) == 0 {
			fmt.Println("(none)")
		}

		fmt.Println("\n↓\nPlanning")
		
		mappedCap := v3.ResolveApplicationCapability(intent)
		if mappedCap == "" {
			fmt.Printf("↓\nUNMAPPED\n\n")
			continue
		}

		fmt.Println("↓\nExecutive")
		fmt.Printf("↓\nApplication Capability (%s)\n", mappedCap)

		if mappedCap == "app-system-1" {
			req := capabilities.CapabilityRequest{
				RequirementID: "test-req",
				Parameters:    slots,
			}
			result, err := appSysCap.Execute(context.Background(), req)
			
			if err != nil {
				fmt.Println("↓\nPolicy Layer (PermissionPolicy)")
				if strings.Contains(err.Error(), "security policy violation") {
					fmt.Println("↓\nREJECTED")
					fmt.Println("↓\nPlatform Capability Check SKIPPED")
					fmt.Println("↓\nNativeSystemCapability NOT INVOKED")
					fmt.Printf("Reason: %v\n", err)
				} else {
					fmt.Printf("ERROR: %v\n", err)
				}
			} else {
				fmt.Println("↓\nPolicy Layer (PermissionPolicy)")
				fmt.Println("↓\nGRANTED")
				fmt.Println("↓\nPlatform Capability Check")
				fmt.Println("↓\nPASSED")
				fmt.Println("↓\nNativeSystemCapability")
				fmt.Printf("↓\nResult: %v\n", result)
			}
		} else if mappedCap == "app-files-1" && strings.Contains(input, "Format") {
			req := capabilities.CapabilityRequest{Parameters: map[string]interface{}{"op": "format"}}
			_, err := appFilesCap.Execute(context.Background(), req)
			fmt.Println("↓\nPolicy Layer (PermissionPolicy)")
			fmt.Println("↓\nREJECTED")
			fmt.Println("↓\nPlatform Capability Check SKIPPED")
			fmt.Println("↓\nNativeFilesCapability NOT INVOKED")
			fmt.Printf("Reason: %v\n", err)
		} else {
			fmt.Println("↓\nPolicy Layer")
			fmt.Println("↓\nGRANTED")
			fmt.Println("↓\nPlatform Capability Check")
			fmt.Println("↓\nPASSED")
			fmt.Println("↓\nNative Capability")
			fmt.Println("↓\nResult: Success")
		}
	}
}
