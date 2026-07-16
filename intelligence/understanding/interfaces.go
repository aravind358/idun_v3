package understanding

import (
	"context"
	"fmt"

	"idun/intelligence/communication"
	"idun/intelligence/executive"
)

// Config holds runtime configuration options for the Understanding subsystem.
type Config struct {
	// MaxBeamWidth limits total preserved hypotheses (default 3).
	MaxBeamWidth int

	// AmbiguityDeltaThreshold sets the maximum P_eff difference for runner-up survival (default 0.15).
	AmbiguityDeltaThreshold float64

	// AdmissionThreshold sets the minimum P_eff confidence required for broadcast (default 0.70).
	AdmissionThreshold float64
}

// Validate verifies runtime configuration parameters.
func (c Config) Validate() error {
	if c.MaxBeamWidth <= 0 {
		return fmt.Errorf("understanding: invalid MaxBeamWidth %d", c.MaxBeamWidth)
	}
	if c.AmbiguityDeltaThreshold < 0.0 || c.AmbiguityDeltaThreshold > 1.0 {
		return fmt.Errorf("understanding: invalid AmbiguityDeltaThreshold %f", c.AmbiguityDeltaThreshold)
	}
	if c.AdmissionThreshold < 0.0 || c.AdmissionThreshold > 1.0 {
		return fmt.Errorf("understanding: invalid AdmissionThreshold %f", c.AdmissionThreshold)
	}
	return nil
}

// Option configures functional overrides for Config.
type Option func(*Config)

// PayloadStorer defines the interface required to persist and retrieve payloads to/from CAS storage.
type PayloadStorer interface {
	Store(ctx context.Context, data []byte) (string, error)
	Retrieve(ctx context.Context, key string) ([]byte, error)
}

// WithConfigOptions applies functional options to a default Config.
func WithConfigOptions(opts ...Option) Config {
	c := Config{
		MaxBeamWidth:            MaxBeamWidth,
		AmbiguityDeltaThreshold: DefaultAmbiguityDelta,
		AdmissionThreshold:      0.70,
	}
	for _, opt := range opts {
		opt(&c)
	}
	return c
}

// UnderstandingService defines the primary capability interface for Phase 1
// and future implementation phases of CognitiveAbility.Understanding.
type UnderstandingService interface {
	executive.UnderstandingAbility // Implements Executive AbilityDriver contract

	// InterpretEnvelope transforms a perception Envelope into a canonical SemanticFrame.
	InterpretEnvelope(ctx context.Context, perceptionEnv communication.Envelope) (SemanticFrame, error)

	// InterpretWithPrior transforms a perception Envelope conditioned on a top-down dialogue prior.
	InterpretWithPrior(ctx context.Context, perceptionEnv communication.Envelope, prior string) (SemanticFrame, error)

	// Name returns the canonical Kernel component name ("Intelligence.Understanding").
	Name() string

	// Start boots the Understanding service lifecycle.
	Start() error

	// Close gracefully shuts down the Understanding service.
	Close() error
}
