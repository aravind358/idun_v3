package system

import "context"

// SystemProvider abstracts all mechanical OS integrations.
// This isolates the capability from native implementations and enables seamless mock testing.
type SystemProvider interface {
	// GetSystemInfo retrieves normalized overall host system metrics.
	GetSystemInfo(ctx context.Context) (map[string]interface{}, error)
	GetEnvInfo(ctx context.Context) (map[string]interface{}, error)
	GetHostInfo(ctx context.Context) (map[string]interface{}, error)
	GetCPUInfo(ctx context.Context) (map[string]interface{}, error)
	GetMemoryInfo(ctx context.Context) (map[string]interface{}, error)
	GetDiskInfo(ctx context.Context) (map[string]interface{}, error)

	// ExecutePower dispatches elevated power operations to the host.
	ExecutePower(ctx context.Context, action SystemOperation) error
}
