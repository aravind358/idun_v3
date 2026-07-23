package template

import (
	"idun/capabilities"
)

// Metadata returns the immutable definition of the Template Capability.
func Metadata() capabilities.CapabilityMetadata {
	return capabilities.CapabilityMetadata{
		Name:                    "TemplateCapability", // TODO: Replace with Capability Name
		Category:                capabilities.CategorySystem, // TODO: Replace with proper CapabilityCategory
		Description:             "A reusable template for native capabilities.", // TODO: Replace description
		Version:                 "1.0.0", // TODO: Set initial version
		APIVersion:              "1.0.0",
		MinimumFrameworkVersion: "1.0.0",
		SupportedPlatforms:      []string{"windows", "linux", "darwin"}, // TODO: Adjust for capability constraints
		PermissionCategory:      "Template", // TODO: Map to actual category
		ImplementationType:      "Native",
		Author:                  "IDUN Core",
		Tags:                    []string{"template", "native"}, // TODO: Replace tags
	}
}
