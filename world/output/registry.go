package output

import (
	"idun/world"
)

// PluginRegistry provides dynamic discovery and routing to output plugins based on modality.
type PluginRegistry interface {
	// Register adds a new OutputPlugin to the registry.
	Register(plugin OutputPlugin) error
	
	// Get returns the appropriate plugin for a given modality.
	Get(modality world.Modality) (OutputPlugin, error)
	
	// Active returns all currently registered plugins.
	Active() []OutputPlugin
}
