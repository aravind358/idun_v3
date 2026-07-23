package template

import "context"

// TemplateProvider abstracts native host operations.
// This isolates the capability from host APIs and enables seamless mock testing.
type TemplateProvider interface {
	// ExecuteExample provides a scaffold structure.
	// TODO: Replace with actual capability provider methods.
	ExecuteExample(ctx context.Context) (map[string]interface{}, error)
}
