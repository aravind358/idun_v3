# Native Files Capability

## Purpose
The Native Files Capability provides mechanical, platform-agnostic file system interaction for IDUN V3. It strictly operates within the bounds of the Phase 3A Capability Philosophy, ensuring isolation from the Intelligence layer and rigorous safety and permission verifications.

## Architecture
- **Router**: `execute.go` and `execute_*.go` separate operations securely.
- **Provider Interface**: Defines strict actions without linking directly to `os`.
- **Validation**: Enforces basic path traversal protections. Destructive operations (delete, copy, overwrite) strictly require explicit parameters.

## Tests
- Uses `MockProvider` ensuring tests do not hit the disk.
