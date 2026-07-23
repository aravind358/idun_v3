package automation

import "context"

// AutomationProvider abstracts native automation operations.
// This isolates the capability from host APIs and enables seamless mock testing.
type AutomationProvider interface {
	// Mouse
	MouseMove(ctx context.Context, x, y int) error
	MouseClick(ctx context.Context, button string, clicks int) error
	MouseScroll(ctx context.Context, deltaX, deltaY int) error

	// Keyboard
	KeyboardPress(ctx context.Context, key string) error
	KeyboardRelease(ctx context.Context, key string) error
	KeyboardType(ctx context.Context, text string) error

	// Clipboard
	ClipboardRead(ctx context.Context) (string, error)
	ClipboardWrite(ctx context.Context, text string) error

	// Screen
	CaptureScreen(ctx context.Context, region map[string]int) ([]byte, error)

	// Windows
	ListWindows(ctx context.Context) ([]map[string]interface{}, error)
	GetWindow(ctx context.Context, handle string) (map[string]interface{}, error)
	FocusWindow(ctx context.Context, handle string) error
	MinimizeWindow(ctx context.Context, handle string) error
	MaximizeWindow(ctx context.Context, handle string) error
	RestoreWindow(ctx context.Context, handle string) error

	// Processes
	ListProcesses(ctx context.Context) ([]map[string]interface{}, error)
	GetProcess(ctx context.Context, pid int) (map[string]interface{}, error)
}
