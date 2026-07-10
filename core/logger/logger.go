// Package logger is the Logger Service — Core Service one of three.
//
// The Logger Service is the first Core Service implemented because it is
// needed by every other Core Service and by every pillar above Core Services.
//
// Responsibility: record structured, levelled system events from IDUN
// components to a configurable output sink.
//
// What this package intentionally does NOT do:
//   - Decide what is worth logging   (caller's policy)
//   - Alert or notify                (World Interface concern)
//   - Analyse log content            (Intelligence concern)
//   - Persist logs with query access (V2/V3 concern)
//   - Route through the Kernel Bus   (cross-cutting infrastructure; direct injection)
//   - Terminate the process          (Fatal/Panic are rejected)
//   - Register itself                (the Host owns all wiring)
//
// Architecture: ADR-CS-001 (frozen 2026-07-09)
//
// Bus exception: Logger write methods bypass the Kernel Bus entirely.
// Services receive a Writer interface at wiring time via direct dependency
// injection. No permission rule is required in the Permission Engine for
// Logger. This prevents circular dependencies and ensures Logger is
// available during startup before the Bus is fully operational.
//
// Kernel uses log.Println from the standard library and must not be changed
// to use this package. Logger Service is above the Kernel, not inside it.
package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// ============================================================
// Level
// ============================================================

// Level is an ordered severity type for log entries.
//
// Level is a named integer type rather than a plain int so the compiler
// catches assignments of arbitrary integers. The four constants are the
// complete and closed set of valid levels for Version 1.
//
// Ordering (lowest to highest):
//
//	Debug < Info < Warn < Error
//
// A Logger configured with minimum level L discards all entries whose
// level is strictly less than L. Only entries at or above L are written.
type Level int

const (
	// Debug is for low-level diagnostic events useful during development.
	// Suppressed in production (default minimum level is Info).
	Debug Level = iota

	// Info is for normal, expected operational events.
	// This is the default minimum level. Most logging calls use Info.
	Info

	// Warn is for anomalous but recoverable conditions.
	// The system continues operating; the event warrants attention.
	Warn

	// Error is for failures that require attention.
	// The system may continue but something went wrong.
	Error
)

// levelString returns the human-readable uppercase label for a Level.
// It is used only by the formatter — callers never need to convert.
func levelString(l Level) string {
	switch l {
	case Debug:
		return "DEBUG"
	case Info:
		return "INFO "
	case Warn:
		return "WARN "
	case Error:
		return "ERROR"
	default:
		return "UNKN "
	}
}

// ============================================================
// Field
// ============================================================

// Field is a structured key-value pair attached to a log entry.
//
// Fields provide machine-parseable context without embedding information
// inside the message string. The Logger formats Value internally —
// callers provide raw values; no manual conversion is required.
//
// Examples:
//
//	Field{Key: "service", Value: "StorageService"}
//	Field{Key: "key",     Value: "memory/record/abc"}
//	Field{Key: "error",   Value: err}
//	Field{Key: "bytes",   Value: 512}
//
// Value is typed as any (alias for interface{}) so callers can log
// errors, integers, booleans, durations, and structs without conversion.
// Typed field constructors (ErrorField, IntField, etc.) are deferred
// to V2 — no V1 caller requires them.
type Field struct {
	Key   string
	Value any
}

// ============================================================
// Writer — the capability interface
// ============================================================

// Writer is the capability interface that services depend on.
//
// Services that need logging receive a Writer, never a *Logger.
// This is the same architectural decision made for the Kernel's Locator
// interface: components depend on capabilities, not on concrete types.
//
// Benefits:
//   - Test doubles implement Writer without constructing a real Logger
//   - The Logger implementation can be replaced without changing callers
//   - Lower coupling between Core Services
//
// The Writer interface intentionally does not include:
//   - Name() — that is the kernel.Component interface, not logging
//   - SetLevel() — no V1 caller; deferred to V2
//   - With() — no V1 caller; deferred to V2
type Writer interface {
	Debug(message string, fields ...Field)
	Info(message string, fields ...Field)
	Warn(message string, fields ...Field)
	Error(message string, fields ...Field)
}

// ============================================================
// Config
// ============================================================

// Config carries the optional parameters the Host may supply to NewLogger.
//
// Both fields are optional. When left at their zero value, NewLogger
// applies the documented defaults before constructing the Logger.
//
// Why a Config struct instead of function parameters?
//   - Future fields (formatter, caller depth) can be added here without
//     changing the NewLogger signature.
//   - The caller's wiring code stays readable: Config{Level: Debug}.
//
// Default values:
//   - Level:  Info  (correct production default; suppresses debug noise)
//   - Output: os.Stdout (correct development default; no config required)
type Config struct {
	// Level is the minimum severity. Entries below this level are discarded.
	// Zero value (Debug) would log everything — too verbose for production.
	// NewLogger replaces the zero value with Info, the intended default.
	// Callers that genuinely want Debug must set Level: Debug explicitly.
	Level Level

	// Output is the sink for formatted log entries.
	// Any io.Writer is accepted: os.Stdout, os.Stderr, a *bytes.Buffer (tests),
	// or a *os.File (log rotation in V2+).
	// When nil, NewLogger defaults to os.Stdout.
	Output io.Writer

	// levelSet is an unexported sentinel that distinguishes an explicitly
	// set Level: Debug from an unset Level field (both are zero values).
	// Without this, there is no way to know whether Level: 0 means
	// "the caller wants Debug" or "the caller left it unset".
	levelSet bool
}

// WithLevel returns a copy of cfg with Level set to l, recording that
// it was explicitly assigned. This is the correct way to request Debug
// level without ambiguity.
//
// Usage:
//
//	cfg := logger.Config{}.WithLevel(logger.Debug)
func (c Config) WithLevel(l Level) Config {
	c.Level = l
	c.levelSet = true
	return c
}

// WithOutput returns a copy of cfg with Output set to w.
// This allows fluent chaining with WithLevel in test helpers and
// in Host wiring code.
//
// Usage:
//
//	cfg := logger.Config{}.WithLevel(logger.Debug).WithOutput(buf)
func (c Config) WithOutput(w io.Writer) Config {
	c.Output = w
	return c
}

// ============================================================
// Logger
// ============================================================

// Logger is the concrete structured event recorder.
//
// Logger satisfies two interfaces:
//   - logger.Writer   (Debug/Info/Warn/Error) — for service callers
//   - kernel.Component (Name())               — for the Service Registry
//
// Do not construct Logger with a struct literal. Always use NewLogger so
// that defaults are applied and the output sink is initialised.
//
// Logger is safe for concurrent use. Multiple goroutines may call write
// methods simultaneously without producing interleaved or corrupted output.
// The mutex serialises each complete entry write; it does not hold locks
// during formatting.
type Logger struct {
	level  Level
	output io.Writer
	mu     sync.Mutex
}

// NewLogger constructs a Logger with the given configuration.
//
// Defaults applied when not explicitly set:
//   - cfg.Output nil → os.Stdout
//   - cfg.Level not set via WithLevel → Info
//
// Returns an error if the configured Level is outside the valid range.
// This keeps NewLogger consistent with NewBus and kernel.Boot, which
// both return errors rather than panicking on invalid input.
func NewLogger(cfg Config) (*Logger, error) {
	// Apply output default.
	output := cfg.Output
	if output == nil {
		output = os.Stdout
	}

	// Apply level default.
	// If the caller used WithLevel, levelSet is true and we honour their choice.
	// If levelSet is false, the zero value of Level is Debug, but the intended
	// default is Info — apply it.
	level := cfg.Level
	if !cfg.levelSet {
		level = Info
	}

	// Validate level. The Level type is an int, so nothing prevents a caller
	// from passing an out-of-range value like Level(99). Reject it explicitly.
	if level < Debug || level > Error {
		return nil, fmt.Errorf("logger: NewLogger failed — Level %d is not valid", int(level))
	}

	return &Logger{
		level:  level,
		output: output,
	}, nil
}

// ============================================================
// kernel.Component interface
// ============================================================

// Name returns "LoggerService".
//
// This method makes *Logger satisfy the kernel.Component interface,
// which is required for the Host to register Logger in the Service Registry.
// The name is fixed — Logger is a singular, well-defined service.
//
// Satisfying kernel.Component does not require importing the kernel package.
// Go's structural typing means any type with a Name() string method implicitly
// satisfies kernel.Component. No import cycle is introduced.
func (l *Logger) Name() string {
	return "LoggerService"
}

// ============================================================
// Writer interface — write methods
// ============================================================

// Debug records a Debug-level event.
//
// Discarded if the Logger's minimum level is above Debug.
// Use for low-level diagnostic events during development.
func (l *Logger) Debug(message string, fields ...Field) {
	l.write(Debug, message, fields)
}

// Info records an Info-level event.
//
// The most frequently used level. Use for normal operational events:
// service startup, successful operations, lifecycle transitions.
func (l *Logger) Info(message string, fields ...Field) {
	l.write(Info, message, fields)
}

// Warn records a Warn-level event.
//
// Use for anomalous but recoverable conditions. The system continues
// operating; the event warrants attention but not an error response.
func (l *Logger) Warn(message string, fields ...Field) {
	l.write(Warn, message, fields)
}

// Error records an Error-level event.
//
// Use for failures that require attention. The operation failed;
// the system may continue operating but something went wrong.
func (l *Logger) Error(message string, fields ...Field) {
	l.write(Error, message, fields)
}

// ============================================================
// Internal write path
// ============================================================

// write is the single internal entry point for all log output.
//
// Steps:
//  1. Level filter — discard entries below the minimum level.
//     This happens before formatting to avoid work for suppressed entries.
//  2. Format — build the complete entry string from timestamp, level,
//     message, and fields. Formatting happens before the lock is acquired
//     so the critical section is as short as possible.
//  3. Write — acquire the mutex and write the formatted entry to the sink.
//     The mutex ensures each entry is written atomically with no interleaving.
//
// Error handling in the write path:
// If the sink write fails, Logger attempts to write to os.Stderr as a fallback.
// If the stderr write also fails, the entry is discarded silently.
// Logger never panics and never propagates sink errors to the caller —
// the caller cannot meaningfully handle a logging failure.
func (l *Logger) write(level Level, message string, fields []Field) {
	// Step 1 — level filter (no lock needed; level is immutable after construction).
	if level < l.level {
		return
	}

	// Step 2 — format the complete entry before acquiring the lock.
	entry := format(level, message, fields)

	// Step 3 — write atomically.
	l.mu.Lock()
	_, err := fmt.Fprintln(l.output, entry)
	l.mu.Unlock()

	// Fallback: if the primary sink fails, try stderr.
	// This handles the case where stdout is closed or redirected to /dev/null
	// and the caller is debugging a silent system.
	if err != nil && l.output != os.Stderr {
		//nolint:errcheck // Intentional: if stderr also fails, discard silently.
		fmt.Fprintln(os.Stderr, entry)
	}
}

// format produces a single human-readable log entry string.
//
// Format (Version 1):
//
//	2026-07-09T07:45:00Z [INFO ] StorageService — Write completed  key=memory/record/abc  bytes=512
//
// Components:
//   - RFC3339 UTC timestamp (second precision is sufficient for V1)
//   - Level label in fixed-width brackets (5 chars including trailing space for alignment)
//   - Message
//   - Structured fields as key=value pairs separated by two spaces
//
// JSON format is deferred to V2. No V1 caller requires machine-parseable
// JSON output, and JSON adds encoding/json import overhead.
func format(level Level, message string, fields []Field) string {
	ts := time.Now().UTC().Format(time.RFC3339)
	label := levelString(level)

	if len(fields) == 0 {
		return fmt.Sprintf("%s [%s] %s", ts, label, message)
	}

	// Build field string: key=value  key=value  ...
	// Two spaces between fields improve readability in a terminal.
	fieldStr := ""
	for i, f := range fields {
		if i == 0 {
			fieldStr = fmt.Sprintf("  %s=%v", f.Key, f.Value)
		} else {
			fieldStr += fmt.Sprintf("  %s=%v", f.Key, f.Value)
		}
	}

	return fmt.Sprintf("%s [%s] %s%s", ts, label, message, fieldStr)
}
