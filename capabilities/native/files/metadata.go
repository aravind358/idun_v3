package files

import (
	"idun/capabilities"
)

// Metadata returns the immutable definition of the Native Files Capability.
func Metadata() capabilities.CapabilityMetadata {
	return capabilities.CapabilityMetadata{
		Name:                    "NativeFilesCapability",
		Category:                capabilities.CategoryFiles,
		Description:             "Provides mechanical access to native file system operations.",
		Version:                 "3.2.0",
		APIVersion:              "1.0.0",
		MinimumFrameworkVersion: "1.0.0",
		SupportedPlatforms:      []string{"windows", "linux", "darwin"},
		PermissionCategory:      "Files",
		ImplementationType:      "Native",
		Author:                  "IDUN Core",
		Tags:                    []string{"native", "files", "fs", "io"},
	}
}
