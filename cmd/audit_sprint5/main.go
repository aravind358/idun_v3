package main

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"idun/boundary/perception"
	"idun/capabilities"
	appfiles "idun/capabilities/applications/files"
	"idun/core/foundation"
	coretime "idun/core/time"
	
	underv3 "idun/intelligence/understanding/v3"
	undercomps "idun/intelligence/understanding/v3/composers"
	underext "idun/intelligence/understanding/v3/extractors"
	undernorms "idun/intelligence/understanding/v3/normalizers"
	underspl "idun/intelligence/understanding/v3/splitter"
)

type MockNativeFiles struct {
	Called bool
	Req    capabilities.CapabilityRequest
}

func (m *MockNativeFiles) Execute(ctx context.Context, req capabilities.CapabilityRequest) (capabilities.CapabilityResult, error) {
	m.Called = true
	m.Req = req
	return capabilities.CapabilityResult{Success: true}, nil
}

func (m *MockNativeFiles) ID() string { return "files-native-1" }
func (m *MockNativeFiles) Metadata() capabilities.CapabilityMetadata {
	return capabilities.CapabilityMetadata{Name: "NativeFilesCapability"}
}
func (m *MockNativeFiles) State() capabilities.CapabilityState {
	return capabilities.CapabilityState{Lifecycle: capabilities.LifecycleHealthy, Operational: capabilities.StatusHealthy}
}

type MockResolver struct {
	Native *MockNativeFiles
}

func (r *MockResolver) Resolve(ctx context.Context, reqID string, name string, args map[string]string) (capabilities.Capability, error) {
	if name == "NativeFilesCapability" {
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
		// Allowed Operations
		"Create directory Docs/Projects",
		"List files in Docs",
		"Open Docs/report.txt",
		"Open docs/../report.pdf",
		"Move images/logo.png to archive",
		"Open ./notes.txt",
		"Open images/./logo.png",
		// Rejected Operations
		"Open C:\\Windows\\System32",
		"Delete D:\\Games",
		"Copy ../../secret.txt",
		"Delete C:\\Windows\\System32",
		"Delete ../../config",
		"Move report.pdf to C:\\Windows",
		"Delete everything",
		"Delete ../../../secret.txt",
		"Open ..\\..\\Windows\\System32",
		"Open C:\\Windows\\explorer.exe",
		"Delete C:\\Users",
		"List D:\\",
		"Delete all files",
		"Remove all folders",
		"Format drive",
		"Erase disk",
	}

	workspaceRoot := "C:\\Projects\\idun_v3\\sandbox"
	
	allowedCount := 0
	blockedCount := 0

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
		
		// Map intent to AppCapability via Planning
		mappedCap := ""
		if intent == "file_operation" || intent == "create_directory" || intent == "list_files" {
			mappedCap = "app-files-1"
		} else {
			fmt.Printf("\nPlanning\n↓\nUNMAPPED\n\n")
			continue
		}

		fmt.Printf("\n↓\nPlanning\n↓\nExecutive\n↓\n%s\n↓\n", mappedCap)

		// Create Capability and Execute
		mockNative := &MockNativeFiles{}
		resolver := &MockResolver{Native: mockNative}
		appCap := appfiles.New(resolver, workspaceRoot)

		req := capabilities.CapabilityRequest{
			RequirementID: string(envID),
			Parameters:    params,
		}

		// Re-implementing the app-files-1 internal steps here just to print the trace 
		
		operation := params["operation"]
		filename := params["filename"]
		directory := params["directory"]
		source := params["source"]
		destination := params["destination"]
	
		targetPath := ""
		if filename != "" {
			targetPath = filename
		} else if directory != "" {
			targetPath = directory
		} else if source != "" {
			targetPath = source
		}

		// WorkspaceResolver Trace
		fmt.Printf("WorkspaceResolver\n")
		fmt.Printf("    • Original Path: %s\n", targetPath)
		
		cleanPath := filepath.Clean(targetPath)
		resolvedPath := ""
		if filepath.IsAbs(cleanPath) {
			resolvedPath = filepath.Clean(cleanPath)
		} else if targetPath != "" {
			resolvedPath = filepath.Join(workspaceRoot, cleanPath)
		} else {
			resolvedPath = workspaceRoot
		}
		
		fmt.Printf("    • Resolved Path: %s\n", resolvedPath)
		fmt.Printf("    • Canonical Path: %s\n", resolvedPath)
		
		if destination != "" {
			fmt.Printf("    • Original Dest: %s\n", destination)
			cleanDest := filepath.Clean(destination)
			if filepath.IsAbs(cleanDest) {
				fmt.Printf("    • Canonical Dest: %s\n", filepath.Clean(cleanDest))
			} else {
				fmt.Printf("    • Canonical Dest: %s\n", filepath.Join(workspaceRoot, cleanDest))
			}
		}

		fmt.Printf("↓\nPermissionPolicy\n↓\n")

		res, _ := appCap.Execute(ctx, req)
		
		if !res.Success && res.Error != nil && res.Error.Code == "SecurityViolation" {
			fmt.Printf("REJECTED\n↓\nNativeFilesCapability NOT INVOKED\n")
			fmt.Printf("Reason: %v\n\n", res.Error.Message)
			blockedCount++
		} else if !res.Success {
			fmt.Printf("ERROR: %v\n\n", res.Error.Message)
		} else {
			if mockNative.Called {
				fmt.Printf("GRANTED\n↓\nNativeFilesCapability\n")
				fmt.Printf("    • Executed Op: %s\n", mockNative.Req.Parameters["operation"])
				fmt.Printf("    • Final Path: %s\n", mockNative.Req.Parameters["path"])
				if mockNative.Req.Parameters["destination"] != "" {
					fmt.Printf("    • Final Dest: %s\n", mockNative.Req.Parameters["destination"])
				}
				fmt.Printf("\n")
				allowedCount++
			} else {
				fmt.Printf("ERROR: Success but NativeFilesCapability NOT CALLED\n\n")
			}
		}
		
		_ = operation // to avoid unused variable error
	}
	
	fmt.Printf("========================================\n")
	fmt.Printf("Metrics:\n")
	fmt.Printf("Native Capability Invocations\n")
	fmt.Printf("    Allowed Requests: %d\n", allowedCount)
	fmt.Printf("    Blocked Requests: %d\n", blockedCount)
	fmt.Printf("========================================\n")
}
