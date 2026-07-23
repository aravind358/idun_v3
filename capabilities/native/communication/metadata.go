package communication

import (
	"idun/capabilities"
)

// Metadata returns the immutable definition of the Native Communication Capability.
func Metadata() capabilities.CapabilityMetadata {
	return capabilities.CapabilityMetadata{
		Name:                    "NativeCommunicationCapability",
		Category:                capabilities.CategoryCommunication,
		Description:             "Provides mechanical message transport and communication operations.",
		Version:                 "3.3.0",
		APIVersion:              "1.0.0",
		MinimumFrameworkVersion: "1.0.0",
		SupportedPlatforms:      []string{"windows", "linux", "darwin"},
		PermissionCategory:      "Communication",
		ImplementationType:      "Native",
		Author:                  "IDUN Core",
		Tags:                    []string{"native", "communication", "transport", "message"},
	}
}
