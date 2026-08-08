package output

import (
	"fmt"
	"sync"
	"idun/world"
)

// DefaultPluginRegistry provides a thread-safe implementation of PluginRegistry.
type DefaultPluginRegistry struct {
	mu      sync.RWMutex
	plugins map[world.Modality]OutputPlugin
}

// NewDefaultPluginRegistry creates a new empty PluginRegistry.
func NewDefaultPluginRegistry() *DefaultPluginRegistry {
	return &DefaultPluginRegistry{
		plugins: make(map[world.Modality]OutputPlugin),
	}
}

// Register adds a new OutputPlugin to the registry.
// It maps the plugin to all its supported modalities.
func (r *DefaultPluginRegistry) Register(plugin OutputPlugin) error {
	if plugin == nil {
		return fmt.Errorf("world/output: cannot register nil plugin")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, mod := range plugin.SupportedModalities() {
		// A more advanced implementation might handle priority and multiple plugins per modality.
		// For now, we simply register the primary plugin for the modality.
		r.plugins[mod] = plugin
	}

	return nil
}

// Get returns the appropriate plugin for a given modality.
func (r *DefaultPluginRegistry) Get(modality world.Modality) (OutputPlugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, ok := r.plugins[modality]
	if !ok {
		return nil, fmt.Errorf("world/output: no plugin registered for modality '%s'", modality)
	}

	return plugin, nil
}

// Active returns all currently registered plugins.
func (r *DefaultPluginRegistry) Active() []OutputPlugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Use a map to deduplicate plugins that support multiple modalities
	unique := make(map[string]OutputPlugin)
	for _, p := range r.plugins {
		unique[p.Name()] = p
	}

	var active []OutputPlugin
	for _, p := range unique {
		active = append(active, p)
	}

	return active
}
