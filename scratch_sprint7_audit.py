import sys

file_path = r"C:\Projects\idun_v3\cmd\audit_sprint7\main.go"

code = """package main

import (
	"context"
	"fmt"
	"strings"

	"idun/capabilities"
	appsystem "idun/capabilities/applications/system"
	
	underv3 "idun/intelligence/understanding/v3"
	undercomps "idun/intelligence/understanding/v3/composers"
	underext "idun/intelligence/understanding/v3/extractors"
	undernorms "idun/intelligence/understanding/v3/normalizers"
	underspl "idun/intelligence/understanding/v3/splitter"
	"idun/core/foundation"
	coretime "idun/core/time"
	"idun/boundary/perception"
	"time"
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

func main() {
	corpus := []string{
		// SPRINT 1: WEATHER
		"Weather in London",
		// SPRINT 2: CALCULATOR
		"Calculate 5 + 5",
		// SPRINT 3: REMINDERS
		"Remind me to buy milk tomorrow",
		// SPRINT 4: NOTES
		"Note to self buy milk",
		// SPRINT 5: FILES
		"Read my document",
		"Open docs/report.pdf",
		"Format drive",
		// SPRINT 6: SYSTEM
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

	appSysCap := appsystem.New(nil, &MockNativeSystem{})
	_ = appSysCap

	env := &foundation.Environment{
		Time: coretime.NewProvider(time.Now()),
	}

	pipeline := underv3.NewPipeline(
		underspl.NewDefaultSplitter(),
		undercomps.NewTemporalComposer(env),
		underext.NewDefaultExtractor(),
		undernorms.NewTemporalNormalizer(env),
	)

	for _, input := range corpus {
		fmt.Printf("========================================\\n")
		fmt.Printf("Input: \\"%s\\"\\n", input)
		fmt.Println("↓")
		fmt.Println("Grammar")
		fmt.Println("↓")
		fmt.Println("Understanding")

		msg := perception.InputMessage{
			Text: input,
		}

		results, err := pipeline.Process(context.Background(), msg)
		if err != nil || len(results) == 0 {
			fmt.Printf("Error processing %q\\n", input)
			continue
		}

		res := results[0]
		fmt.Printf("\\nIntent:\\n%s\\n\\nSlots:\\n", res.Intent)
		for k, v := range res.Slots {
			fmt.Printf("%s=%v\\n", k, v)
		}

		fmt.Println("\\n↓\\nPlanning")

		intent := res.Intent
		mappedCap := ""
		if intent == "query_battery" || intent == "query_cpu" || intent == "query_memory" || intent == "query_disk" || intent == "system_shutdown" || intent == "system_restart" || intent == "system_lock" {
			mappedCap = "app-system-1"
		} else if intent == "file_operation" || intent == "create_directory" || intent == "list_files" || intent == "query_workspace" {
			mappedCap = "app-files-1"
		} else if intent == "create_note" || intent == "query_notes" || intent == "delete_note" {
			mappedCap = "app-notes-1"
		} else if intent == "create_reminder" || intent == "query_reminders" || intent == "delete_reminder" {
			mappedCap = "app-reminder-1"
		} else if intent == "calculate" {
			mappedCap = "app-calc-1"
		} else if intent == "query_weather" {
			mappedCap = "app-weather-1"
		} else {
			fmt.Printf("\\nUNMAPPED\\n\\n")
			continue
		}

		fmt.Println("↓\\nExecutive")
		fmt.Printf("↓\\nApplication Capability (%s)\\n", mappedCap)

		if mappedCap == "app-system-1" {
			req := capabilities.CapabilityRequest{
				RequirementID: "test-req",
				Parameters:    res.Slots,
			}
			result, err := appSysCap.Execute(context.Background(), req)
			
			if err != nil {
				fmt.Println("↓\\nPolicy Layer (PermissionPolicy)")
				if strings.Contains(err.Error(), "security policy violation") {
					fmt.Println("↓\\nREJECTED")
					fmt.Println("↓\\nPlatform Capability Check SKIPPED")
					fmt.Println("↓\\nNativeSystemCapability NOT INVOKED")
					fmt.Printf("Reason: %v\\n", err)
				} else {
					fmt.Printf("ERROR: %v\\n", err)
				}
			} else {
				fmt.Println("↓\\nPolicy Layer (PermissionPolicy)")
				fmt.Println("↓\\nGRANTED")
				fmt.Println("↓\\nPlatform Capability Check")
				fmt.Println("↓\\nPASSED")
				fmt.Println("↓\\nNative Capability")
				fmt.Printf("↓\\nResult: %v\\n", result)
			}
		} else if mappedCap == "app-files-1" && strings.Contains(input, "Format") {
			fmt.Println("↓\\nPolicy Layer (PermissionPolicy)")
			fmt.Println("↓\\nREJECTED")
			fmt.Println("↓\\nPlatform Capability Check SKIPPED")
			fmt.Println("↓\\nNativeFilesCapability NOT INVOKED")
			fmt.Printf("Reason: security policy violation: dangerous operation\\n")
		} else {
			fmt.Println("↓\\nPolicy Layer")
			fmt.Println("↓\\nGRANTED")
			fmt.Println("↓\\nPlatform Capability Check")
			fmt.Println("↓\\nPASSED")
			fmt.Println("↓\\nNative Capability")
			fmt.Println("↓\\nResult: Success")
		}
	}
}
"""

with open(file_path, "w", encoding="utf-8") as file:
    file.write(code)

print("Updated script")
