package files

import "context"

// FileProvider abstracts native file system operations.
// This isolates the capability from host APIs and enables seamless mock testing.
type FileProvider interface {
	// Read Operations
	ReadFile(ctx context.Context, path string) ([]byte, error)
	ReadText(ctx context.Context, path string) (string, error)
	FileExists(ctx context.Context, path string) (bool, error)
	GetMetadata(ctx context.Context, path string) (map[string]interface{}, error)
	
	// Write Operations
	WriteFile(ctx context.Context, path string, data []byte, append bool) error
	CopyFile(ctx context.Context, src, dest string) error
	MoveFile(ctx context.Context, src, dest string) error
	DeleteFile(ctx context.Context, path string) error
	
	// Directory Operations
	ListDirectory(ctx context.Context, path string, recursive bool) ([]map[string]interface{}, error)
	CreateDirectory(ctx context.Context, path string) error
	DeleteDirectory(ctx context.Context, path string) error
	
	// Search Operations
	SearchFiles(ctx context.Context, root, pattern string, recursive, caseSensitive bool) ([]string, error)
	
	// Hash Operations
	CalculateHash(ctx context.Context, path, algorithm string) (string, error)
	
	// Temporary Operations
	CreateTemporaryFile(ctx context.Context, prefix, suffix string) (string, error)
	CreateTemporaryDirectory(ctx context.Context, prefix string) (string, error)
}
