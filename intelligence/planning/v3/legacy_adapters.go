package v3

import (
	"context"
	"idun/capabilities"
)

// LegacyCapabilityAdapter wraps the legacy core capability manager to implement the V3 CapabilityRegistry.
type LegacyCapabilityAdapter struct {
	manager capabilities.CapabilityManager
}

// NewLegacyCapabilityAdapter creates a new legacy capability adapter for Planning V3.
func NewLegacyCapabilityAdapter(manager capabilities.CapabilityManager) CapabilityRegistry {
	return &LegacyCapabilityAdapter{manager: manager}
}

// Discover maps a resolved intent (goal) to a list of available capabilities.
func (a *LegacyCapabilityAdapter) Discover(ctx context.Context, goal string) ([]CapabilityDescriptor, error) {
	if a.manager == nil || a.manager.Registry() == nil {
		// Mock capability if no manager is wired
		return []CapabilityDescriptor{
			NewCapabilityDescriptor(CapabilityID("sys-communicative-1"), "Fallback communicative capability", []string{"target"}),
		}, nil
	}

	var results []CapabilityDescriptor
	allCaps := a.manager.Registry().List()

	for _, c := range allCaps {
		meta := c.Metadata()
		
		match := false
		switch goal {
		case "query_weather":
			if c.ID() == "app-weather-1" {
				desc := CapabilityDescriptor{
					ID:          CapabilityID(c.ID()),
					Description: meta.Description,
					Params:      []string{"location", "operation"}, 
				}
				results = append(results, desc)
			}
			continue
		case "query_time", "query_date":
			match = c.ID() == "sys-time-1"
		case "calculate", "math":
			if c.ID() == "app-calc-1" {
				desc := CapabilityDescriptor{
					ID:          CapabilityID(c.ID()),
					Description: meta.Description,
					Params:      []string{"operand1", "operand2", "operator", "expression"}, 
				}
				results = append(results, desc)
			}
			continue
		case "manage_notes":
			if c.ID() == "app-notes-1" {
				desc := CapabilityDescriptor{
					ID:          CapabilityID(c.ID()),
					Description: meta.Description,
					Params:      []string{"operation", "title", "content"}, 
				}
				results = append(results, desc)
			}
			continue
		case "manage_reminders", "set_alarm", "create_reminder":
			if c.ID() == "app-rem-1" {
				desc := CapabilityDescriptor{
					ID:          CapabilityID(c.ID()),
					Description: meta.Description,
					Params:      []string{"operation", "person", "time", "task", "duration", "date"}, 
				}
				results = append(results, desc)
			}
			continue
		case "file_operation", "create_directory", "list_files":
			if c.ID() == "app-files-1" {
				desc := CapabilityDescriptor{
					ID:          CapabilityID(c.ID()),
					Description: meta.Description,
					Params:      []string{"operation", "filename", "source", "destination", "directory", "path", "data_text"},
				}
				results = append(results, desc)
			}
			continue
		case "query_battery", "query_cpu", "query_memory", "query_disk", "system_shutdown", "system_restart", "system_lock":
			if c.ID() == "app-system-1" {
				desc := CapabilityDescriptor{
					ID:          CapabilityID(c.ID()),
					Description: meta.Description,
					Params:      []string{"operation"},
				}
				results = append(results, desc)
			}
			continue
		case "system_command", "query_status", "cancel_action":
			match = c.ID() == "sys-native-1"
		}

		if match {
			desc := CapabilityDescriptor{
				ID:          CapabilityID(c.ID()),
				Description: meta.Description,
				Params:      []string{}, 
			}
			results = append(results, desc)
		}
	}

	// We return empty if no match, allowing orchestrator.go to inject sys-communicative-1 for communicative intents.

	return results, nil
}
