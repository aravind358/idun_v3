package native

import (
	"context"
	"time"

	"idun/capabilities"
	nativefiles "idun/capabilities/native/files"
	nativenetwork "idun/capabilities/native/network"
	nativesystem "idun/capabilities/native/system"
	nativetime "idun/capabilities/native/time"
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
		Realization:   capabilities.Generative,
		Data:          map[string]interface{}{"status": "placeholder_executed", "category": p.Metadata().Category},
		Duration:      1 * time.Millisecond,
	}, nil
}

// LoadNativeCapabilities instantiates and registers the baseline V1 categories.
func LoadNativeCapabilities(registry capabilities.CapabilityRegistry, deps NativeCapabilityDependencies) error {
	caps := []capabilities.Capability{
		nativetime.New(deps.Time),
		nativesystem.New(nil, nativesystem.NewNativeProvider(), deps.Scheduler),
		nativefiles.New(nil, nativefiles.NewNativeProvider()),
		nativenetwork.New(nil, nativenetwork.NewNativeProvider()),
	}

	for _, c := range caps {
		if err := registry.Register(c); err != nil {
			return err
		}
	}
	return nil
}
