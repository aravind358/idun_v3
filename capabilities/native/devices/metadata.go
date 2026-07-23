package devices

import (
	"idun/capabilities"
)

// Metadata returns the immutable definition of the Native Devices Capability.
func Metadata() capabilities.CapabilityMetadata {
	return capabilities.CapabilityMetadata{
		Name:                    "NativeDevicesCapability",
		Category:                capabilities.CategoryDevicesSensors,
		Description:             "Provides mechanical access to physical devices and hardware sensors.",
		Version:                 "3.6.0",
		APIVersion:              "1.0.0",
		MinimumFrameworkVersion: "1.0.0",
		SupportedPlatforms:      []string{"windows", "linux", "darwin"},
		PermissionCategory:      "Devices",
		ImplementationType:      "Native",
		Author:                  "IDUN Core",
		Tags:                    []string{"native", "devices", "sensors", "usb", "bluetooth", "battery", "power", "location", "hid"},
	}
}
