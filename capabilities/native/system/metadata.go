package system

import "idun/capabilities"

// Metadata returns the immutable definition of the Native System Capability.
func Metadata() capabilities.CapabilityMetadata {
	return capabilities.CapabilityMetadata{
		Name:                    "NativeSystemCapability",
		Category:                capabilities.CategorySystem,
		Description:             "Provides mechanical access to native OS information and power controls.",
		Version:                 "3.1.2",
		APIVersion:              "1.0.0",
		MinimumFrameworkVersion: "1.0.0",
		SupportedPlatforms:      []string{"windows", "linux", "darwin"},
		PermissionCategory:      "System",
		ImplementationType:      "Native",
		Author:                  "IDUN Core",
		Tags:                    []string{"native", "os", "system", "power", "info"},
	}
}
