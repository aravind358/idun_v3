package applications

import (
	"idun/capabilities"
	"idun/capabilities/applications/calculator"
	"idun/capabilities/applications/core"
	"idun/capabilities/applications/notes"
	"idun/capabilities/applications/reminder"
	"idun/capabilities/applications/weather"
)

// LoadApplicationCapabilities instantiates and registers the baseline application capabilities.
// It requires an AppCapabilityDependencies container which provides access to the NativeCapabilityResolver
// and any temporary technical debt dependencies (like Storage/Scheduler).
func LoadApplicationCapabilities(registry capabilities.CapabilityRegistry, deps core.AppCapabilityDependencies) error {
	caps := []capabilities.Capability{
		calculator.New(deps),
		reminder.New(deps),
		notes.New(deps),
		weather.New(deps),
	}

	for _, c := range caps {
		if err := registry.Register(c); err != nil {
			return err
		}
	}

	return nil
}
