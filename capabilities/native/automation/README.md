# Native Automation Capability

## Purpose
The Native Automation Capability provides mechanical execution boundaries for OS automation primitives natively exposed by the operating system for IDUN V3. Following the Phase 3A Capability Philosophy, it handles simulated mechanical input (mouse/keyboard), clipboard control, screen capture, and window/process interaction. It strictly enforces a pure-execution boundary, avoiding any cognitive AI interpretation, OCR, or browser element understanding.

## Scope Integration
This represents the final implementation phase (Phase 3B.7) of the Native Capability Layer. The Automation Capability performs blind mechanical manipulation of input and presentation boundaries.

## Architecture
- **Router**: Handlers split across `execute_mouse.go`, `execute_keyboard.go`, `execute_clipboard.go`, `execute_screen.go`, `execute_windows.go`, and `execute_process.go`.
- **Provider Interface**: The `AutomationProvider` isolates host OS integration.
- **Implementations**: `NativeProvider` acts as the execution facade (stubbed for future low-level GUI bindings). `MockProvider` perfectly simulates automation interactions offline for continuous integration testing.

## Modules
- **Mouse**: Cursor manipulation and clicking.
- **Keyboard**: Typomatic simulation and key events.
- **Clipboard**: Direct read/write access.
- **Screen**: Raw bitmapped frame buffer captures (No vision processing).
- **Windows**: Handle-based manipulation of desktop components.
- **Processes**: External runtime enumeration.
