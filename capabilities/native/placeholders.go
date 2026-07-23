package native

import (
	"context"
	"time"

	"idun/capabilities"
)

// Define a simple helper to quickly generate placeholder capabilities.
func buildPlaceholder(id string, name string, category capabilities.CapabilityCategory) capabilities.Capability {
	meta := capabilities.CapabilityMetadata{
		Name:        name,
		Category:    category,
		Description: "V1 Placeholder for " + name,
		Version:     "1.0.0",
		Author:      "IDUN Core",
		Tags:        []string{"native", "placeholder"},
	}
	base := capabilities.NewBaseCapability(id, meta)
	return &placeholderCapability{BaseCapability: base}
}

type placeholderCapability struct {
	capabilities.BaseCapability
}

func (p *placeholderCapability) Execute(ctx context.Context, req capabilities.CapabilityRequest) (capabilities.CapabilityResult, error) {
	// A placeholder capability returns immediately with success.
	return capabilities.CapabilityResult{
		RequirementID: req.RequirementID,
		Success:       true,
		Data:          map[string]interface{}{"status": "placeholder_executed", "category": p.Metadata().Category},
		Duration:      1 * time.Millisecond,
	}, nil
}

// LoadNativeCapabilities instantiates and registers the baseline V1 categories.
func LoadNativeCapabilities(registry capabilities.CapabilityRegistry) error {
	caps := []capabilities.Capability{
		buildPlaceholder("sys-core", "SystemProcess", capabilities.CategorySystem),
		buildPlaceholder("fs-core", "FileSystemOps", capabilities.CategoryFiles),
		buildPlaceholder("hw-sensors", "HardwareSensors", capabilities.CategoryDevicesSensors),

		buildPlaceholder("comm-user", "UserNotifier", capabilities.CategoryCommunication),
		buildPlaceholder("media-mgr", "MediaManager", capabilities.CategoryMedia),
		buildPlaceholder("net-tcp", "NetworkSockets", capabilities.CategoryNetwork),
		buildPlaceholder("ext-api", "ExternalIntegrations", capabilities.CategoryExternalServices),
		buildPlaceholder("auto-webhook", "WebhookTriggers", capabilities.CategoryAutomation), // No conditional logic
	}

	for _, c := range caps {
		if err := registry.Register(c); err != nil {
			return err
		}
	}
	return nil
}
