package calculator

import (
	"idun/capabilities"
)

// Metadata returns the immutable definition of the Calculator Capability.
func Metadata() capabilities.CapabilityMetadata {
	return capabilities.CapabilityMetadata{
		Name:                    "CalculatorApp",
		Category:                capabilities.CategorySystem, // Deterministic logic, closest match
		Description:             "Application capability for pure deterministic math operations.",
		Version:                 "1.0.0",
		APIVersion:              "1.0.0",
		MinimumFrameworkVersion: "1.0.0",
		SupportedPlatforms:      []string{"windows", "linux", "darwin"},
		PermissionCategory:      "Calculator",
		ImplementationType:      "Application",
		Author:                  "IDUN Core",
		Tags:                    []string{"application", "math", "calculator"},
	}
}
