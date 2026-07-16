// Package text implements the TextInputAdapter and TextOutputAdapter for the World subsystem.
//
// Architecture Version: 2.0.0-FROZEN
//
// These adapters are the only Phase 1 concrete implementations of the InputAdapter
// and OutputAdapter contracts. They operate on io.Reader and io.Writer respectively,
// making them fully testable via any reader/writer pair (bufio, bytes.Buffer, os.Stdin, etc.).
//
// Adapter Identity:
// Both adapters carry immutable identity (Name, AdapterVersion, AdapterFingerprint) to
// enable exact replay provenance when adapter implementations evolve over time (Refinement 10).
package text
