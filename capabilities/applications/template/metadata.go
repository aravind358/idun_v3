package template

import (
	"idun/capabilities"
)

// Metadata returns the immutable definition of the Template Capability.
func Metadata() capabilities.CapabilityMetadata {
	return capabilities.CapabilityMetadata{
		Name:                    "AppTemplateCapability",
		Category:                capabilities.CategorySystem,
		Description:             "A reusable template for application capabilities.",
		Version:                 "1.0.0",
		APIVersion:              "1.0.0",
		MinimumFrameworkVersion: "1.0.0",
		SupportedPlatforms:      []string{"windows", "linux", "darwin"},
		PermissionCategory:      "Template",
		ImplementationType:      "Application",
		Author:                  "IDUN Core",
		Tags:                    []string{"template", "application"},
	}
}
