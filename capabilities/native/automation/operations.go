package automation

// AutomationOperation defines strongly-typed constants for permitted operations.
type AutomationOperation string

const (
	// Mouse
	OperationMouseMove   AutomationOperation = "mouse_move"
	OperationMouseClick  AutomationOperation = "mouse_click"
	OperationMouseScroll AutomationOperation = "mouse_scroll"

	// Keyboard
	OperationKeyboardPress   AutomationOperation = "keyboard_press"
	OperationKeyboardRelease AutomationOperation = "keyboard_release"
	OperationKeyboardType    AutomationOperation = "keyboard_type"

	// Clipboard
	OperationClipboardRead  AutomationOperation = "clipboard_read"
	OperationClipboardWrite AutomationOperation = "clipboard_write"

	// Screen
	OperationCaptureScreen AutomationOperation = "capture_screen"

	// Windows
	OperationListWindows    AutomationOperation = "list_windows"
	OperationGetWindow      AutomationOperation = "get_window"
	OperationFocusWindow    AutomationOperation = "focus_window"
	OperationMinimizeWindow AutomationOperation = "minimize_window"
	OperationMaximizeWindow AutomationOperation = "maximize_window"
	OperationRestoreWindow  AutomationOperation = "restore_window"

	// Processes
	OperationListProcesses AutomationOperation = "list_processes"
	OperationGetProcess    AutomationOperation = "get_process"
)

// IsValid validates if a string matches a known AutomationOperation.
func (o AutomationOperation) IsValid() bool {
	switch o {
	case OperationMouseMove, OperationMouseClick, OperationMouseScroll,
		OperationKeyboardPress, OperationKeyboardRelease, OperationKeyboardType,
		OperationClipboardRead, OperationClipboardWrite,
		OperationCaptureScreen,
		OperationListWindows, OperationGetWindow, OperationFocusWindow,
		OperationMinimizeWindow, OperationMaximizeWindow, OperationRestoreWindow,
		OperationListProcesses, OperationGetProcess:
		return true
	}
	return false
}
