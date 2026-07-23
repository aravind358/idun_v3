# Native Communication Capability

## Purpose
The Native Communication Capability provides mechanical message transport operations for IDUN V3. Following the Phase 3A Capability Philosophy, it strictly separates the transport mechanism from any cognitive interpretation, planning, or decision-making.

## Architecture
- **Router**: Uses `execute.go` and `execute_*.go` handlers mapped to strongly typed enum operations.
- **Provider Interface**: The `CommunicationProvider` acts as the execution boundary.
- **Implementations**: `NativeProvider` for OS/IPC interactions and `MockProvider` for stateless testing without dependencies.

## Operations
- **Transport**: `send_message`, `receive_message`
- **Management**: `get_history`, `delete_message`, `mark_read`, `mark_unread`, `get_status`
