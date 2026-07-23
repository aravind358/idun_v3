package media

import (
	"idun/capabilities"
)

// Metadata returns the immutable definition of the Native Media Capability.
func Metadata() capabilities.CapabilityMetadata {
	return capabilities.CapabilityMetadata{
		Name:                    "NativeMediaCapability",
		Category:                capabilities.CategoryMedia,
		Description:             "Provides mechanical access to OS multimedia resources.",
		Version:                 "3.5.0",
		APIVersion:              "1.0.0",
		MinimumFrameworkVersion: "1.0.0",
		SupportedPlatforms:      []string{"windows", "linux", "darwin"},
		PermissionCategory:      "Media",
		ImplementationType:      "Native",
		Author:                  "IDUN Core",
		Tags:                    []string{"native", "media", "audio", "video", "image", "camera"},
	}
}
