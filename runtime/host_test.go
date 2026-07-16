package runtime

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

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
