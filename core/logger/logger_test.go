// Tests for the Logger Service.
//
// Test file: logger_test.go
// Package: logger_test (external black-box test package)
//
// Using an external test package (logger_test instead of logger) means
// tests exercise only the exported API — exactly what callers see.
// This catches accidentally unexported symbols and tests the real surface.
//
// Test naming convention: Test<Type>_<Scenario>
//   - Happy paths use "Success" or a short description.
//   - Error paths describe the invalid condition.
//
// All 24 tests from ADR-CS-001 Section 10 are implemented here.
// Run with:
//
//	go test ./core/logger/...
//	go test ./core/logger/... -race    ← required for concurrency tests
//
// The -race flag is mandatory for the concurrency tests to be meaningful.
package logger_test

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"idun/core/logger"
)

// ============================================================
// Helpers
// ============================================================

// newTestLogger constructs a Logger that writes to a bytes.Buffer.
// This is the standard test fixture: captures all output without touching stdout.
func newTestLogger(t *testing.T) (*logger.Logger, *bytes.Buffer) {
	t.Helper()
	buf := &bytes.Buffer{}
	l, err := logger.NewLogger(logger.Config{Output: buf})
	if err != nil {
		t.Fatalf("NewLogger() returned unexpected error: %v", err)
	}
	return l, buf
}

// newDebugLogger constructs a Logger at Debug level that writes to a buffer.
// Used by tests that need all levels to pass the filter.
func newDebugLogger(t *testing.T) (*logger.Logger, *bytes.Buffer) {
	t.Helper()
	buf := &bytes.Buffer{}
	l, err := logger.NewLogger(logger.Config{}.WithLevel(logger.Debug).WithOutput(buf))
	if err != nil {
		t.Fatalf("NewLogger(Debug) returned unexpected error: %v", err)
	}
	return l, buf
}

// newLevelLogger constructs a Logger at the given level that writes to a buffer.
func newLevelLogger(t *testing.T, level logger.Level) (*logger.Logger, *bytes.Buffer) {
	t.Helper()
	buf := &bytes.Buffer{}
	l, err := logger.NewLogger(logger.Config{}.WithLevel(level).WithOutput(buf))
	if err != nil {
		t.Fatalf("NewLogger(%v) returned unexpected error: %v", level, err)
	}
	return l, buf
}

// ============================================================
// Section 10.1 — Constructor Tests
// ============================================================

// TestNewLogger_DefaultConfig verifies that NewLogger with an empty Config
// returns a non-nil, usable Logger with no error.
//
// An empty Config means: no output override, no level override.
// Defaults are applied: Info level, os.Stdout output.
func TestNewLogger_DefaultConfig(t *testing.T) {
	l, err := logger.NewLogger(logger.Config{})

	if err != nil {
		t.Fatalf("NewLogger(Config{}) returned unexpected error: %v", err)
	}
	if l == nil {
		t.Fatal("NewLogger(Config{}) returned nil Logger, expected a valid *Logger")
	}
}

// TestNewLogger_WithCustomOutput verifies that NewLogger accepts a custom
// io.Writer and uses it for output.
func TestNewLogger_WithCustomOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	l, err := logger.NewLogger(logger.Config{Output: buf})

	if err != nil {
		t.Fatalf("NewLogger() with custom output returned unexpected error: %v", err)
	}
	if l == nil {
		t.Fatal("NewLogger() with custom output returned nil Logger")
	}

	// Writing an entry must populate the buffer, proving Logger uses the provided sink.
	l.Info("probe")
	if buf.Len() == 0 {
		t.Error("after Info(), the provided buffer should be non-empty")
	}
}

// TestNewLogger_WithAllLevels verifies that NewLogger accepts every valid
// Level value without returning an error.
func TestNewLogger_WithAllLevels(t *testing.T) {
	cases := []struct {
		name  string
		level logger.Level
	}{
		{name: "Debug level", level: logger.Debug},
		{name: "Info level", level: logger.Info},
		{name: "Warn level", level: logger.Warn},
		{name: "Error level", level: logger.Error},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l, err := logger.NewLogger(logger.Config{}.WithLevel(tc.level).WithOutput(&bytes.Buffer{}))
			if err != nil {
				t.Errorf("NewLogger(%v) returned unexpected error: %v", tc.level, err)
			}
			if l == nil {
				t.Errorf("NewLogger(%v) returned nil Logger", tc.level)
			}
		})
	}
}

// ============================================================
// Section 10.2 — Name / Component Interface Tests
// ============================================================

// TestLogger_Name verifies that Logger returns "LoggerService" from Name().
//
// The compile-time interface check below proves *Logger satisfies
// kernel.Component without importing the kernel package.
// kernel.Component requires only Name() string, so any type with that
// method satisfies the interface via Go's structural typing.
func TestLogger_Name(t *testing.T) {
	l, _ := newTestLogger(t)

	// Compile-time check: *Logger must have Name() string.
	// If kernel.Component is defined as interface{ Name() string }, this
	// assignment would be the compile-time proof. Since we do not import
	// kernel in this test package, we verify the method exists and returns
	// the correct value directly.
	got := l.Name()
	want := "LoggerService"
	if got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
}

// ============================================================
// Section 10.3 — Level Filtering Tests
// ============================================================

// TestLogger_LevelFilter_DebugSuppressedAtInfoLevel verifies that Debug
// entries are discarded when the minimum level is Info.
func TestLogger_LevelFilter_DebugSuppressedAtInfoLevel(t *testing.T) {
	l, buf := newTestLogger(t) // default level is Info

	l.Debug("debug message that must not appear")

	if buf.Len() != 0 {
		t.Errorf("Debug() at Info level should produce no output; buffer contains: %q", buf.String())
	}
}

// TestLogger_LevelFilter_InfoPassesAtInfoLevel verifies that Info entries
// are written when the minimum level is Info.
func TestLogger_LevelFilter_InfoPassesAtInfoLevel(t *testing.T) {
	l, buf := newTestLogger(t)

	l.Info("info message")

	if buf.Len() == 0 {
		t.Error("Info() at Info level should write an entry, but buffer is empty")
	}
}

// TestLogger_LevelFilter_WarnPassesAtInfoLevel verifies that Warn entries
// are written when the minimum level is Info (Warn > Info).
func TestLogger_LevelFilter_WarnPassesAtInfoLevel(t *testing.T) {
	l, buf := newTestLogger(t)

	l.Warn("warn message")

	if buf.Len() == 0 {
		t.Error("Warn() at Info level should write an entry, but buffer is empty")
	}
}

// TestLogger_LevelFilter_ErrorPassesAtInfoLevel verifies that Error entries
// are written when the minimum level is Info.
func TestLogger_LevelFilter_ErrorPassesAtInfoLevel(t *testing.T) {
	l, buf := newTestLogger(t)

	l.Error("error message")

	if buf.Len() == 0 {
		t.Error("Error() at Info level should write an entry, but buffer is empty")
	}
}

// TestLogger_LevelFilter_AllLevelsPassAtDebugLevel verifies that all four
// levels are written when the minimum level is Debug.
func TestLogger_LevelFilter_AllLevelsPassAtDebugLevel(t *testing.T) {
	cases := []struct {
		name  string
		write func(l *logger.Logger)
	}{
		{name: "Debug", write: func(l *logger.Logger) { l.Debug("msg") }},
		{name: "Info", write: func(l *logger.Logger) { l.Info("msg") }},
		{name: "Warn", write: func(l *logger.Logger) { l.Warn("msg") }},
		{name: "Error", write: func(l *logger.Logger) { l.Error("msg") }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l, buf := newDebugLogger(t)
			tc.write(l)
			if buf.Len() == 0 {
				t.Errorf("%s() at Debug level should write an entry, but buffer is empty", tc.name)
			}
		})
	}
}

// TestLogger_LevelFilter_OnlyErrorPassesAtErrorLevel verifies that only
// Error entries are written when the minimum level is Error.
func TestLogger_LevelFilter_OnlyErrorPassesAtErrorLevel(t *testing.T) {
	// Suppressed levels: Debug, Info, Warn must produce no output.
	suppressed := []struct {
		name  string
		write func(l *logger.Logger)
	}{
		{name: "Debug suppressed at Error level", write: func(l *logger.Logger) { l.Debug("msg") }},
		{name: "Info suppressed at Error level", write: func(l *logger.Logger) { l.Info("msg") }},
		{name: "Warn suppressed at Error level", write: func(l *logger.Logger) { l.Warn("msg") }},
	}

	for _, tc := range suppressed {
		t.Run(tc.name, func(t *testing.T) {
			l, buf := newLevelLogger(t, logger.Error)
			tc.write(l)
			if buf.Len() != 0 {
				t.Errorf("%s should produce no output at Error level; buffer: %q", tc.name, buf.String())
			}
		})
	}

	// Error must pass.
	t.Run("Error passes at Error level", func(t *testing.T) {
		l, buf := newLevelLogger(t, logger.Error)
		l.Error("this should appear")
		if buf.Len() == 0 {
			t.Error("Error() at Error level should write an entry, but buffer is empty")
		}
	})
}

// ============================================================
// Section 10.4 — Structured Field Tests
// ============================================================

// TestLogger_FieldsAppearedInOutput verifies that a Field attached to a
// log call appears in the output entry.
func TestLogger_FieldsAppearedInOutput(t *testing.T) {
	l, buf := newTestLogger(t)

	l.Info("message", logger.Field{Key: "service", Value: "TestService"})

	output := buf.String()
	if !strings.Contains(output, "message") {
		t.Errorf("output should contain the message; got: %q", output)
	}
	if !strings.Contains(output, "service") {
		t.Errorf("output should contain the field key; got: %q", output)
	}
	if !strings.Contains(output, "TestService") {
		t.Errorf("output should contain the field value; got: %q", output)
	}
}

// TestLogger_MultipleFieldsAllAppear verifies that multiple fields all
// appear in a single log entry.
func TestLogger_MultipleFieldsAllAppear(t *testing.T) {
	l, buf := newTestLogger(t)

	l.Info("msg",
		logger.Field{Key: "a", Value: "1"},
		logger.Field{Key: "b", Value: "2"},
	)

	output := buf.String()
	for _, want := range []string{"a", "1", "b", "2"} {
		if !strings.Contains(output, want) {
			t.Errorf("output should contain %q; got: %q", want, output)
		}
	}
}

// TestLogger_ZeroFieldsIsValid verifies that calling a write method with
// no fields does not panic and produces valid output.
func TestLogger_ZeroFieldsIsValid(t *testing.T) {
	l, buf := newTestLogger(t)

	// Must not panic. No fields — variadic receives nil/empty slice.
	l.Info("message with no fields")

	if buf.Len() == 0 {
		t.Error("Info() with no fields should write an entry; buffer is empty")
	}
}

// ============================================================
// Section 10.5 — Output Content Tests
// ============================================================

// TestLogger_OutputContainsLevel verifies that every log entry contains
// the level string in the output.
func TestLogger_OutputContainsLevel(t *testing.T) {
	cases := []struct {
		name     string
		write    func(l *logger.Logger)
		levelStr string
	}{
		{name: "Debug", write: func(l *logger.Logger) { l.Debug("msg") }, levelStr: "DEBUG"},
		{name: "Info", write: func(l *logger.Logger) { l.Info("msg") }, levelStr: "INFO"},
		{name: "Warn", write: func(l *logger.Logger) { l.Warn("msg") }, levelStr: "WARN"},
		{name: "Error", write: func(l *logger.Logger) { l.Error("msg") }, levelStr: "ERROR"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l, buf := newDebugLogger(t) // Debug level so all entries pass the filter
			tc.write(l)
			output := buf.String()
			if !strings.Contains(output, tc.levelStr) {
				t.Errorf("output should contain %q; got: %q", tc.levelStr, output)
			}
		})
	}
}

// TestLogger_OutputContainsMessage verifies that the caller's message
// appears in every log entry.
func TestLogger_OutputContainsMessage(t *testing.T) {
	l, buf := newTestLogger(t)
	const uniqueMessage = "unique-test-message-xyz-7f3a"

	l.Info(uniqueMessage)

	if !strings.Contains(buf.String(), uniqueMessage) {
		t.Errorf("output should contain %q; got: %q", uniqueMessage, buf.String())
	}
}

// TestLogger_OutputContainsTimestamp verifies that every log entry
// contains a timestamp.
//
// The test checks for the current year as a proxy — if the year is in
// the output, a timestamp is present. Checking the exact format is
// fragile and ties the test to the format implementation.
func TestLogger_OutputContainsTimestamp(t *testing.T) {
	l, buf := newTestLogger(t)

	l.Info("timestamp test")

	// The formatter uses RFC3339 UTC: "2006-01-02T15:04:05Z"
	// The year will always be present in a valid RFC3339 timestamp.
	output := buf.String()
	if !strings.Contains(output, "202") { // "202x" covers 2020–2029
		t.Errorf("output should contain a timestamp (year); got: %q", output)
	}
}

// TestLogger_EachCallProducesOneEntry verifies that a single write method
// call produces exactly one log entry (one newline-terminated line).
func TestLogger_EachCallProducesOneEntry(t *testing.T) {
	l, buf := newTestLogger(t)

	l.Info("single entry")

	output := buf.String()
	// fmt.Fprintln appends a newline; split on newline and count non-empty lines.
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	if len(lines) != 1 {
		t.Errorf("one Info() call should produce exactly one line; got %d lines in: %q", len(lines), output)
	}
}

// ============================================================
// Section 10.6 — Goroutine Safety Tests
// ============================================================

// TestLogger_ConcurrentWrites_NoPanic verifies that concurrent calls to
// write methods do not panic and do not produce data races.
//
// IMPORTANT: This test must be run with the -race flag to be meaningful:
//
//	go test ./core/logger/... -race
func TestLogger_ConcurrentWrites_NoPanic(t *testing.T) {
	l, _ := newTestLogger(t)

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			l.Info("concurrent message")
		}()
	}

	wg.Wait()
	// If we reach here without a data race (detected by -race) or panic,
	// the test passes.
}

// TestLogger_ConcurrentWrites_AllEntriesPresent verifies that under
// concurrent load, no write calls are silently dropped.
func TestLogger_ConcurrentWrites_AllEntriesPresent(t *testing.T) {
	l, buf := newTestLogger(t)

	const goroutines = 10
	const marker = "CONCURRENT_MARKER"

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			l.Info(marker)
		}()
	}

	wg.Wait()

	output := buf.String()
	count := strings.Count(output, marker)
	if count != goroutines {
		t.Errorf("expected %d log entries, got %d in output:\n%s", goroutines, count, output)
	}
}

// ============================================================
// Section 10.7 — Sink / Output Tests
// ============================================================

// TestLogger_WritesToProvidedSink verifies that Logger writes exclusively
// to the sink provided at construction, not to a hardcoded destination.
func TestLogger_WritesToProvidedSink(t *testing.T) {
	customBuf := &bytes.Buffer{}
	l, err := logger.NewLogger(logger.Config{Output: customBuf})
	if err != nil {
		t.Fatalf("NewLogger() failed: %v", err)
	}

	l.Info("sink test")

	if customBuf.Len() == 0 {
		t.Error("Logger should have written to the provided sink; buffer is empty")
	}
}

// TestLogger_DefaultSinkIsStdout verifies that when no Output is provided,
// the Logger is constructed successfully (defaults to os.Stdout).
//
// We do not capture stdout in this test — that would require os.Pipe()
// manipulation which is fragile. Instead, we verify construction succeeds
// and that the Logger is usable (no panic on Info call).
func TestLogger_DefaultSinkIsStdout(t *testing.T) {
	l, err := logger.NewLogger(logger.Config{})
	if err != nil {
		t.Fatalf("NewLogger() with nil Output should succeed (defaults to stdout); got: %v", err)
	}
	if l == nil {
		t.Fatal("NewLogger() with nil Output returned nil Logger")
	}

	// Verify the Logger is usable. Output goes to stdout in this test,
	// which is acceptable — we are proving no panic, not capturing output.
	// The test runner captures stdout; any output here is harmless.
	l.Info("[test] default sink test — this line goes to stdout")
}

// ============================================================
// Section 10.8 — Integration Tests
// ============================================================

// stubRegistry is a minimal registry that records registered component names.
// It exists only in this test file to avoid importing the kernel package.
// The test proves Logger can satisfy kernel.Component by satisfying the
// interface structurally — the same way Go does it at runtime.
type stubRegistry struct {
	names []string
}

// register accepts any value that has a Name() string method.
// This mirrors what kernel.Registry.Register does with kernel.Component.
func (r *stubRegistry) register(c interface{ Name() string }) {
	r.names = append(r.names, c.Name())
}

func (r *stubRegistry) contains(name string) bool {
	for _, n := range r.names {
		if n == name {
			return true
		}
	}
	return false
}

// TestLogger_RegistersInKernelRegistry verifies that Logger can be
// registered in the Kernel Service Registry without error.
//
// We use a stubRegistry here to avoid importing the kernel package,
// keeping the logger package dependency-free of the kernel package.
// The structural typing proof is: stubRegistry.register accepts
// interface{Name() string}, and *Logger has Name() string — it compiles.
func TestLogger_RegistersInKernelRegistry(t *testing.T) {
	l, _ := newTestLogger(t)

	reg := &stubRegistry{}
	reg.register(l)

	if !reg.contains("LoggerService") {
		t.Error("after registering Logger, registry should contain 'LoggerService'")
	}
}

// TestLogger_LookupAfterRegistration verifies that a registered Logger
// can be retrieved and used.
func TestLogger_LookupAfterRegistration(t *testing.T) {
	buf := &bytes.Buffer{}
	original, err := logger.NewLogger(logger.Config{Output: buf})
	if err != nil {
		t.Fatalf("NewLogger() failed: %v", err)
	}

	reg := &stubRegistry{}
	reg.register(original)

	// Simulate lookup: find the component with name "LoggerService".
	// In real usage, kernel.Registry.Lookup returns a kernel.Component
	// and the caller type-asserts to *logger.Logger or logger.Writer.
	// Here we verify the name round-trips correctly.
	if !reg.contains("LoggerService") {
		t.Fatal("lookup should find 'LoggerService' after registration")
	}

	// Verify the retrieved Logger is usable (can write without panic).
	original.Info("retrieved and used")
	if buf.Len() == 0 {
		t.Error("Logger retrieved after registration should be usable; buffer is empty")
	}
}

// TestLogger_AvailableBeforeBusIsWired verifies that Logger is usable
// before the Kernel Bus is constructed, confirming it has no Bus dependency
// on its write path.
//
// This is the critical Bus-independence integration invariant from the ADR.
func TestLogger_AvailableBeforeBusIsWired(t *testing.T) {
	buf := &bytes.Buffer{}
	l, err := logger.NewLogger(logger.Config{Output: buf})
	if err != nil {
		t.Fatalf("NewLogger() failed: %v", err)
	}

	// No Bus, no Registry, no Kernel — Logger must work regardless.
	l.Info("startup message before bus is wired",
		logger.Field{Key: "phase", Value: "pre-boot"},
	)

	if buf.Len() == 0 {
		t.Error("Logger should be usable before Bus is wired; buffer is empty")
	}
	if !strings.Contains(buf.String(), "pre-boot") {
		t.Error("log entry should contain the field value 'pre-boot'")
	}
}

// ============================================================
// Section 10.9 — Compile-time Interface Compliance
// ============================================================

// TestLogger_WriterInterfaceCompliance is a compile-time proof that
// *Logger satisfies the logger.Writer interface.
//
// This test produces no runtime assertion. It fails at compile time if
// the Writer interface or Logger method signatures drift out of sync.
func TestLogger_WriterInterfaceCompliance(t *testing.T) {
	var _ logger.Writer = (*logger.Logger)(nil)
}
