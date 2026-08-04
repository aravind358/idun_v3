package notes

import (
	"idun/capabilities"
)

// Metadata returns the immutable definition of the Notes Capability.
func Metadata() capabilities.CapabilityMetadata {
	return capabilities.CapabilityMetadata{
		Name:                    "NotesApp",
		Category:                capabilities.CategoryFiles,
		Description:             "Application capability for managing textual data stores.",
		Version:                 "1.0.0",
		APIVersion:              "1.0.0",
		MinimumFrameworkVersion: "1.0.0",
		SupportedPlatforms:      []string{"windows", "linux", "darwin"},
		PermissionCategory:      "Notes",
		ImplementationType:      "Application",
		Author:                  "IDUN Core",
		Tags:                    []string{"application", "notes", "files"},
	}
}
