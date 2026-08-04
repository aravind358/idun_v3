package reminder

import (
	"idun/capabilities"
)

// Metadata returns the immutable definition of the Reminder Capability.
func Metadata() capabilities.CapabilityMetadata {
	return capabilities.CapabilityMetadata{
		Name:                    "ReminderApp",
		Category:                capabilities.CategorySystem,
		Description:             "Application capability for setting and managing reminders.",
		Version:                 "1.0.0",
		APIVersion:              "1.0.0",
		MinimumFrameworkVersion: "1.0.0",
		SupportedPlatforms:      []string{"windows", "linux", "darwin"},
		PermissionCategory:      "Reminder",
		ImplementationType:      "Application",
		Author:                  "IDUN Core",
		Tags:                    []string{"application", "reminder", "time"},
	}
}
