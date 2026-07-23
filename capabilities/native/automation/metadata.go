package automation

import (
	"idun/capabilities"
)

// Metadata returns the immutable definition of the Native Automation Capability.
func Metadata() capabilities.CapabilityMetadata {
	return capabilities.CapabilityMetadata{
		Name:                    "NativeAutomationCapability",
		Category:                capabilities.CategoryAutomation,
		Description:             "Provides mechanical access to OS input, clipboard, screen, and window primitives.",
		Version:                 "3.7.0",
		APIVersion:              "1.0.0",
		MinimumFrameworkVersion: "1.0.0",
		SupportedPlatforms:      []string{"windows", "linux", "darwin"},
		PermissionCategory:      "Automation",
		ImplementationType:      "Native",
		Author:                  "IDUN Core",
		Tags:                    []string{"native", "automation", "mouse", "keyboard", "clipboard", "screen", "windows", "processes"},
	}
}
