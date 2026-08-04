package main

import (
	"context"
	"fmt"
	"time"

	"idun/boundary/perception"
	"idun/capabilities"
	appsystem "idun/capabilities/applications/system"
	"idun/core/foundation"
	coretime "idun/core/time"
	
	underv3 "idun/intelligence/understanding/v3"
	undercomps "idun/intelligence/understanding/v3/composers"
	underext "idun/intelligence/understanding/v3/extractors"
	undernorms "idun/intelligence/understanding/v3/normalizers"
	underspl "idun/intelligence/understanding/v3/splitter"
)

// Mock NativeSystemCapability that never actually hits the OS
type MockNativeSystem struct {
	Called bool
	Req    capabilities.CapabilityRequest
}

func (m *MockNativeSystem) Execute(ctx context.Context, req capabilities.CapabilityRequest) (capabilities.CapabilityResult, error) {
	m.Called = true
	m.Req = req
	return capabilities.CapabilityResult{Success: true}, nil
}

func (m *MockNativeSystem) ID() string { return "sys-native-1" }
func (m *MockNativeSystem) Metadata() capabilities.CapabilityMetadata {
	return capabilities.CapabilityMetadata{Name: "NativeSystemCapability"}
}
func (m *MockNativeSystem) State() capabilities.CapabilityState {
	return capabilities.CapabilityState{Lifecycle: capabilities.LifecycleHealthy, Operational: capabilities.StatusHealthy}
}

type MockResolver struct {
	Native *MockNativeSystem
}

func (r *MockResolver) Resolve(ctx context.Context, reqID string, name string, args map[string]string) (capabilities.Capability, error) {
	if name == "NativeSystemCapability" {
		return r.Native, nil
	}
	return nil, fmt.Errorf("not found: %s", name)
}

func main() {
	v3Grammar := underv3.NewDefaultGrammarSpecialist()
	exts := underext.NewDeterministicExtractors()
	timeSvc := coretime.NewTimeService(time.Local)
	tempNorm := undernorms.NewDeterministicTemporalNormalizer(timeSvc)
	norms := undernorms.NewDeterministicNormalizers(tempNorm)
	comps := undercomps.NewDeterministicTemporalComposer()
	orch := underv3.NewOrchestrator(v3Grammar, nil, nil, exts, norms, comps, underspl.NewDeterministicSplitter())

	corpus := []string{
		// Information
		"Battery percentage",
		"Battery status",
		"CPU usage",
		"Memory usage",
		"Disk usage",
		// System Actions
		"Shut down the computer",
		"Restart the computer",
		"Lock the computer",
		// Invalid / Unsupported
		"Destroy the computer",
		"Wipe my disk",
		"Kill Windows",
		"Format drive",
		// Additional Sprint 6 Review
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

	infoCount := 0
	controlCount := 0
	blockedCount := 0
	nativeAllowed := 0
	nativeBlocked := 0

	for _, text := range corpus {
		fmt.Printf("========================================\n")
		fmt.Printf("Input: %q\n", text)
		fmt.Printf("↓\n")

		// 1. Understanding
		envID, _ := foundation.NewUUID()
		artID, _ := foundation.NewUUID()
		env, _ := perception.NewBuilder().
			ArtifactID(artID).
			EnvelopeID(envID).
			RawInput(text).
			Version("3.0").
			Timestamp(time.Now()).
			Build()

		ctx := context.Background()
		batch, err := orch.Analyze(ctx, env)
		interps := batch.Interpretations()
		if err != nil || len(interps) == 0 {
			fmt.Printf("Grammar\n↓\nUnderstanding: Failed to parse\n\n")
			continue
		}

		interp := interps[0]
		intent := interp.PrimaryIntent()
		
		fmt.Printf("Grammar\n↓\nUnderstanding\n\nIntent:\n%s\n\nSlots:\n", intent)
		params := make(map[string]string)
		for _, s := range interp.PrimaryHypothesis().Slots() {
			fmt.Printf("%s=%s\n", s.Name(), s.Value())
			params[s.Name()] = s.Value()
		}
		
		params["intent"] = intent
		params["raw_input"] = text

		// Map intent to AppCapability via Planning
		mappedCap := ""
		if intent == "query_battery" || intent == "query_cpu" || intent == "query_memory" || intent == "query_disk" {
			mappedCap = "app-system-1"
			infoCount++
		} else if intent == "system_shutdown" || intent == "system_restart" || intent == "system_lock" {
			mappedCap = "app-system-1"
			controlCount++
		} else {
			fmt.Printf("\nPlanning\n↓\nUNMAPPED\n\n")
			continue
		}

		fmt.Printf("\n↓\nPlanning\n↓\nExecutive\n↓\n%s\n↓\nPermissionPolicy\n↓\n", mappedCap)

		// Create Capability and Execute
		mockNative := &MockNativeSystem{}
		resolver := &MockResolver{Native: mockNative}
		appCap := appsystem.New(resolver)

		req := capabilities.CapabilityRequest{
			RequirementID: string(envID),
			Parameters:    params,
		}

		res, _ := appCap.Execute(ctx, req)
		
		if !res.Success && res.Error != nil && res.Error.Code == "SecurityViolation" {
			fmt.Printf("REJECTED\n↓\nNativeSystemCapability NOT INVOKED\n")
			fmt.Printf("Reason: %v\n\n", res.Error.Message)
			blockedCount++
		} else if !res.Success {
			fmt.Printf("ERROR: %v\n\n", res.Error.Message)
		} else {
			if mockNative.Called {
				fmt.Printf("GRANTED\n↓\nNativeSystemCapability\n")
				fmt.Printf("    • Executed Op: %s\n\n", mockNative.Req.Parameters["operation"])
				nativeAllowed++
			} else {
				fmt.Printf("ERROR: Success but NativeSystemCapability NOT CALLED\n\n")
			}
		}
	}
	
	fmt.Printf("========================================\n")
	fmt.Printf("Metrics:\n")
	fmt.Printf("Information Queries: %d\n", infoCount)
	fmt.Printf("Control Operations: %d\n", controlCount)
	fmt.Printf("Blocked System Operations: %d\n", blockedCount)
	fmt.Printf("Native System Invocations\n")
	fmt.Printf("    Allowed: %d\n", nativeAllowed)
	fmt.Printf("    Blocked: %d\n", nativeBlocked)
	fmt.Printf("========================================\n")
}
