package runtime

import (
	"time"
)

// RuntimeBootReport is a purely diagnostic summary generated after startup attempts.
// It details exactly which components started, which were skipped due to configuration,
// durations, and any non-fatal warnings discovered during wiring.
type RuntimeBootReport struct {
	// BootDuration records the total time elapsed during topological startup.
	BootDuration time.Duration `json:"boot_duration"`

	// StartedComponents lists canonical names of all components successfully started.
	StartedComponents []string `json:"started_components"`

	// SkippedComponents lists canonical names of optional subsystems that were skipped.
	SkippedComponents []string `json:"skipped_components"`

	// Manifest holds the immutable runtime version and fingerprint metadata.
	Manifest *RuntimeManifest `json:"manifest"`

	// Warnings contains diagnostic messages for non-fatal configuration issues.
	Warnings []string `json:"warnings"`

	// Success indicates whether the boot sequence completed without terminal errors.
	Success bool `json:"success"`
}
