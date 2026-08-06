package system

import (
	"context"
	"fmt"
	"strings"
	"time"

	"idun/capabilities"
	"idun/capabilities/applications/core"
	nativesystem "idun/capabilities/native/system"
)

// Metadata returns the capability definition.
func Metadata() capabilities.CapabilityMetadata {
	return capabilities.CapabilityMetadata{
		Name:                    "SystemManager",
		Category:                capabilities.CategorySystem,
		Description:             "Manages high-level system queries and power operations with safety policies.",
		Version:                 "1.0.0",
		APIVersion:              "1.0.0",
		MinimumFrameworkVersion: "1.0.0",
		Author:                  "IDUN Core",
		Tags:                    []string{"system", "app"},
	}
}

// Capability provides system queries and control operations with application-level security.
type Capability struct {
	core.AppCapability
}

// New creates a new System application capability.
func New(resolver core.NativeCapabilityResolver) *Capability {
	return &Capability{
		AppCapability: core.NewAppCapability("app-system-1", Metadata(), resolver),
	}
}

// PermissionPolicy evaluates the security policy on a system operation.
type PermissionPolicy struct{}

// IsAllowed enforces the System Policy Rule.
func (p *PermissionPolicy) IsAllowed(intent string, operation string, rawInput string) error {
	intentLower := strings.ToLower(intent)
	
	// Fast-path read-only queries
	switch intentLower {
	case "query_battery", "query_cpu", "query_memory", "query_disk":
		return nil
	case "system_shutdown", "system_restart", "system_lock":
		// Control operations require exact semantic mapping and block dangerous wildcards.
		// If the original input contains dangerous ambiguous language, block it.
		dangerousPhrases := []string{
			"shut everything down", "destroy", "kill", "wipe", "format drive", "erase disk",
		}
		rawLower := strings.ToLower(rawInput)
		opLower := strings.ToLower(operation)
		for _, phrase := range dangerousPhrases {
			if strings.Contains(rawLower, phrase) || strings.Contains(opLower, phrase) {
				return fmt.Errorf("security policy violation: dangerous operation %q is blocked", phrase)
			}
		}
		return nil
	default:
		// Attempt to block unrecognized operations. 
		// Even if semantic grammar matched it as something else, we enforce bounds here.
		return fmt.Errorf("security policy violation: unrecognized or unsupported system intent %q", intent)
	}
}

// Execute fulfills the capability execution request.
func (c *Capability) Execute(ctx context.Context, req capabilities.CapabilityRequest) (capabilities.CapabilityResult, error) {
	start := time.Now()

	// 1. Parameter extraction
	operation := req.Parameters["operation"]
	intent := req.Parameters["intent"]
	
	rawInput := req.Parameters["raw_input"] 
	if rawInput == "" {
		rawInput = operation 
	}

	// 2. Permission Policy Check
	policy := &PermissionPolicy{}
	if err := policy.IsAllowed(intent, operation, rawInput); err != nil {
		return capabilities.CapabilityResult{
			RequirementID: req.RequirementID,
			Success:       false,
			Error: &capabilities.CapabilityError{
				Code:    "SecurityViolation",
				Message: err.Error(),
			},
			Duration: time.Since(start),
		}, nil
	}

	// 3. Map Cognitive Semantics to Native Operations
	var nativeOp nativesystem.SystemOperation
	switch strings.ToLower(intent) {
	case "query_battery":
		nativeOp = nativesystem.OperationBattery
	case "query_cpu":
		nativeOp = nativesystem.OperationCPU
	case "query_memory":
		nativeOp = nativesystem.OperationMemory
	case "query_disk":
		nativeOp = nativesystem.OperationDisk
	case "system_shutdown":
		nativeOp = nativesystem.OperationShutdown
	case "system_restart":
		nativeOp = nativesystem.OperationRestart
	case "system_lock":
		nativeOp = nativesystem.OperationLock
	default:
		// Fallback check if it maps directly to native op
		nativeOp = nativesystem.SystemOperation(intent)
		if !nativeOp.IsValid() {
			return capabilities.CapabilityResult{}, fmt.Errorf("unsupported system semantics: %s", intent)
		}
	}

	nativeParams := map[string]string{
		"operation": string(nativeOp),
	}

	// 4. Native System Capability Invocation
	sysCap, err := c.Resolver.Resolve(ctx, req.RequirementID, "NativeSystemCapability", nativeParams)
	if err != nil {
		return capabilities.CapabilityResult{}, fmt.Errorf("failed to resolve native system capability: %w", err)
	}

	nativeReq := capabilities.CapabilityRequest{
		RequirementID: req.RequirementID,
		Parameters:    nativeParams,
	}

	res, err := sysCap.Execute(ctx, nativeReq)
	if err != nil {
		return res, err
	}

	// Intercept the native result to enrich it with presentation routing and semantic operation.
	// We do NOT modify res.Data to preserve semantic purity.
	res.Realization = capabilities.Deterministic
	res.ResponseType = "system"
	res.Operation = intent
	
	return res, nil
}
