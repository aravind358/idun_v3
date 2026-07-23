package network

import (
	"idun/capabilities"
)

// Metadata returns the immutable definition of the Native Network Capability.
func Metadata() capabilities.CapabilityMetadata {
	return capabilities.CapabilityMetadata{
		Name:                    "NativeNetworkCapability",
		Category:                capabilities.CategoryNetwork,
		Description:             "Provides mechanical access to raw OS networking primitives.",
		Version:                 "3.4.0",
		APIVersion:              "1.0.0",
		MinimumFrameworkVersion: "1.0.0",
		SupportedPlatforms:      []string{"windows", "linux", "darwin"},
		PermissionCategory:      "Network",
		ImplementationType:      "Native",
		Author:                  "IDUN Core",
		Tags:                    []string{"native", "network", "tcp", "udp", "http", "dns", "socket"},
	}
}
