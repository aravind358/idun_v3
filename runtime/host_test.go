package runtime

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// testRealizationModel returns the Ollama model to use in integration tests.
// It respects the IDUN_REALIZATION_MODEL env var so CI can override the model
// without code changes. Falls back to "llama3.1:8b" which is the model known
// to be installed locally.
func testRealizationModel() string {
	if m := os.Getenv("IDUN_REALIZATION_MODEL"); m != "" {
		return m
	}
	// Use mock by default in tests to avoid 180s hangs when Ollama is unavailable
	return "mock"
}

func TestRuntimeHost_Lifecycle(t *testing.T) {
	tempDir := filepath.Join(os.TempDir(), "idun_runtime_test_lifecycle")
	_ = os.RemoveAll(tempDir)
	defer os.RemoveAll(tempDir)

	cfg := DefaultConfiguration()
	cfg.StoragePath = tempDir
	cfg.EnableLogging = false

	h, err := NewHost(cfg)
	if err != nil {
		t.Fatalf("NewHost failed: %v", err)
	}

	if status := h.Status(); status != StatusStopped {
		t.Fatalf("Expected initial status STOPPED, got %s", status)
	}

	ctx := context.Background()
	if err := h.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if status := h.Status(); status != StatusRunning {
		t.Fatalf("Expected status RUNNING after boot, got %s", status)
	}

	if k := h.Kernel(); k == nil {
		t.Fatal("Expected active Kernel instance, got nil")
	}
	if ws := h.Workspace(); ws == nil {
		t.Fatal("Expected active Workspace instance, got nil")
	}

	manifest := h.Manifest()
	if manifest == nil {
		t.Fatal("Expected non-nil Manifest")
	}
	if manifest.RuntimeVersion != "2.0.0-FROZEN" {
		t.Fatalf("Expected runtime version 2.0.0-FROZEN, got %s", manifest.RuntimeVersion)
	}
	if manifest.ManifestFingerprint == "" {
		t.Fatal("Expected non-empty ManifestFingerprint")
	}

	report := h.Report()
	if report == nil {
		t.Fatal("Expected non-nil BootReport")
	}
	if !report.Success {
		t.Fatalf("Expected boot report success=true, warnings: %v", report.Warnings)
	}
	if len(report.StartedComponents) == 0 {
		t.Fatal("Expected started components in boot report, got 0")
	}

	// Verify stopping
	if err := h.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	if status := h.Status(); status != StatusStopped {
		t.Fatalf("Expected status STOPPED after stop, got %s", status)
	}
	if k := h.Kernel(); k != nil {
		t.Fatal("Expected nil Kernel after stop")
	}
}

func TestRuntimeHost_ConfigureInStoppedState(t *testing.T) {
	tempDir := filepath.Join(os.TempDir(), "idun_runtime_test_config")
	_ = os.RemoveAll(tempDir)
	defer os.RemoveAll(tempDir)

	cfg := DefaultConfiguration()
	cfg.StoragePath = tempDir
	cfg.EnableLogging = false

	h, err := NewHost(cfg)
	if err != nil {
		t.Fatalf("NewHost failed: %v", err)
	}

	newCfg := DefaultConfiguration()
	newCfg.StoragePath = tempDir
	newCfg.EnableLogging = false
	newCfg.InitialExecutiveBudget = 50000

	if err := h.Configure(newCfg); err != nil {
		t.Fatalf("Configure in STOPPED state failed: %v", err)
	}

	if err := h.Start(context.Background()); err != nil {
		t.Fatalf("Start after configure failed: %v", err)
	}

	// Attempting to configure or build while running must fail
	if err := h.Configure(newCfg); err == nil {
		t.Fatal("Expected error when calling Configure in RUNNING state, got nil")
	}
	if err := h.Build(); err == nil {
		t.Fatal("Expected error when calling Build in RUNNING state, got nil")
	}

	_ = h.Stop()
}

func TestRuntimeHost_OptionalSubsystemSkipping(t *testing.T) {
	tempDir := filepath.Join(os.TempDir(), "idun_runtime_test_skipping")
	_ = os.RemoveAll(tempDir)
	defer os.RemoveAll(tempDir)

	cfg := DefaultConfiguration()
	cfg.StoragePath = tempDir
	cfg.EnableLogging = false
	cfg.EnabledSubsystems["learning"] = false
	cfg.EnabledSubsystems["reflection"] = false

	h, err := NewHost(cfg)
	if err != nil {
		t.Fatalf("NewHost failed: %v", err)
	}

	if err := h.Start(context.Background()); err != nil {
		t.Fatalf("Start failed with skipped subsystems: %v", err)
	}
	defer h.Stop()

	report := h.Report()
	if report == nil || !report.Success {
		t.Fatalf("Expected successful boot report: %v", report)
	}

	var skippedLearning, skippedReflection bool
	for _, skipped := range report.SkippedComponents {
		if skipped == "Intelligence.learning" {
			skippedLearning = true
		}
		if skipped == "Intelligence.reflection" {
			skippedReflection = true
		}
	}
	if !skippedLearning || !skippedReflection {
		t.Fatalf("Expected both Intelligence.learning and Intelligence.reflection to be listed in SkippedComponents, got: %v", report.SkippedComponents)
	}
}

func TestRuntimeHost_ManifestDeterminism(t *testing.T) {
	tempDir := filepath.Join(os.TempDir(), "idun_runtime_test_manifest")
	_ = os.RemoveAll(tempDir)
	defer os.RemoveAll(tempDir)

	cfg := DefaultConfiguration()
	cfg.StoragePath = tempDir
	cfg.EnableLogging = false
	cfg.ReplayMode = true

	h1, _ := NewHost(cfg)
	_ = h1.Start(context.Background())
	m1 := h1.Manifest()
	fp1 := m1.ManifestFingerprint
	_ = h1.Stop()

	h2, _ := NewHost(cfg)
	_ = h2.Start(context.Background())
	m2 := h2.Manifest()
	fp2 := m2.ManifestFingerprint
	_ = h2.Stop()

	if fp1 != fp2 {
		t.Fatalf("Expected deterministic ManifestFingerprint across runs, got %s and %s", fp1, fp2)
	}
}

type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *syncBuffer) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Len()
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

func TestLayer1EndToEndRuntimeDemonstration(t *testing.T) {
	tempDir := filepath.Join(os.TempDir(), "idun_runtime_test_e2e")
	_ = os.RemoveAll(tempDir)
	defer os.RemoveAll(tempDir)

	cfg := DefaultConfiguration()
	cfg.StoragePath = tempDir
	cfg.EnableLogging = true

	input := bytes.NewReader([]byte("Hello IDUN\n"))
	outBuf := &syncBuffer{}

	h, err := NewHost(cfg, WithIOReaders(input, outBuf), WithRealizationModel(testRealizationModel()))
	if err != nil {
		t.Fatalf("NewHost failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := h.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer h.Stop()

	// Wait up to 180 seconds for the cognitive pipeline (Understanding -> Reasoning -> Planning -> Decision -> Executive -> World)
	// to process the input and emit the final output response through TextOutputAdapter.
	deadline := time.Now().Add(180 * time.Second)
	for time.Now().Before(deadline) {
		if outBuf.Len() > 0 && strings.TrimSpace(outBuf.String()) != "" {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	if outBuf.Len() == 0 || strings.TrimSpace(outBuf.String()) == "" {
		t.Fatalf("expected complete response through TextOutputAdapter from single text interaction, got empty output: %q", outBuf.String())
	}

	outputStr := outBuf.String()
	t.Logf("End-to-End Runtime Demonstration Output:\n%s", outputStr)
}

func TestLayer1ManualInteractionsSuite(t *testing.T) {
	tempDir := filepath.Join(os.TempDir(), "idun_runtime_test_manual_suite")
	_ = os.RemoveAll(tempDir)
	defer os.RemoveAll(tempDir)

	cfg := DefaultConfiguration()
	cfg.StoragePath = tempDir
	cfg.EnableLogging = true

	pipeReader, pipeWriter := io.Pipe()
	outBuf := &syncBuffer{}

	h, err := NewHost(cfg, WithIOReaders(pipeReader, outBuf), WithRealizationModel(testRealizationModel()))
	if err != nil {
		t.Fatalf("NewHost failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := h.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer h.Stop()

	interactions := []string{
		"Hello",
		"Hi",
		"Who are you?",
		"How are you?",
		"Goodbye",
	}

	for i, prompt := range interactions {
		expectedCount := i + 1
		_, err := pipeWriter.Write([]byte(prompt + "\n"))
		if err != nil {
			t.Fatalf("Write prompt %q failed: %v", prompt, err)
		}

		deadline := time.Now().Add(180 * time.Second)
		for time.Now().Before(deadline) {
			lines := strings.Split(strings.TrimSpace(outBuf.String()), "\n")
			var validLines int
			for _, l := range lines {
				if strings.TrimSpace(l) != "" {
					validLines++
				}
			}
			if validLines >= expectedCount {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}

		lines := strings.Split(strings.TrimSpace(outBuf.String()), "\n")
		var validLines []string
		for _, l := range lines {
			if strings.TrimSpace(l) != "" {
				validLines = append(validLines, strings.TrimSpace(l))
			}
		}
		if len(validLines) < expectedCount {
			t.Fatalf("interaction %d (%q) failed: expected %d responses, got %d:\n%v", i+1, prompt, expectedCount, len(validLines), validLines)
		}
		t.Logf("Interaction %d (%q) -> Response: %s", i+1, prompt, validLines[i])
	}

	// Verify "exit" cleanly shuts down through World loop.
	// Allow up to 15 seconds: the LLM inference for the final interaction may still be
	// in-flight when "exit" arrives. Shutdown cancels the context (near-instant), but
	// the HTTP stack needs a moment to propagate the cancellation before the goroutine
	// fully unwinds and h.doneCh is closed.
	_, _ = pipeWriter.Write([]byte("exit\n"))
	select {
	case <-h.Done():
		t.Log("Clean shutdown via 'exit' confirmed.")
	case <-time.After(15 * time.Second):
		t.Fatal("expected h.Done() to close when 'exit' was typed")
	}
}
