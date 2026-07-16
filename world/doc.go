// Package world implements IDUN's World subsystem — the boundary between IDUN and the external world.
//
// Architecture Version: 2.0.0-FROZEN
//
// # Responsibility
//
// World owns all user-facing I/O at the system boundary. It accepts input from external
// adapters, normalizes it, constructs Interaction artifacts, publishes them to the Global
// Workspace, and eventually presents responses back through output adapters.
//
// # World Never Owns
//
// World must never:
//   - Reason, plan, decide, or learn
//   - Evaluate salience or attention
//   - Own memory or workspace state
//   - Modify policies or capabilities
//   - Interpret content or intent (that belongs to Understanding)
//
// # Interaction Flow (Event-Driven)
//
//	TextInputAdapter.Receive()
//	  → creates Interaction (validated, fingerprinted)
//	  → Service.HandleInteraction() publishes Envelope → TopicPerception
//	  → returns immediately (non-blocking)
//	  → Workspace delivers TopicActionExecution envelope asynchronously
//	  → Service.responseHandler() builds Response
//	  → TextOutputAdapter.Send(Response)
//
// World never blocks waiting for Executive. All interaction flows are asynchronous.
package world
