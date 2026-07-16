package world

import (
	"errors"
	"time"
)

// Config defines operational configuration parameters for the World service.
// Per-interaction policy is governed by WorldPolicyProfile; Config covers
// service-level infrastructure concerns only.
type Config struct {
	// EnableTracing determines whether WorldTrace artifacts are recorded per interaction.
	EnableTracing bool `json:"enable_tracing"`

	// EnableSummary determines whether WorldSummary statistics are maintained.
	EnableSummary bool `json:"enable_summary"`

	// MaxPendingInteractions limits the number of in-flight interactions awaiting responses.
	// Zero means unlimited (bounded only by available memory).
	MaxPendingInteractions int `json:"max_pending_interactions"`

	// DefaultSessionID is used when an adapter does not supply a session identifier.
	DefaultSessionID string `json:"default_session_id"`

	// ShutdownDrainTimeout is the maximum time to wait for in-flight interactions
	// to resolve before forceful shutdown.
	ShutdownDrainTimeout time.Duration `json:"shutdown_drain_timeout"`
}

// DefaultConfig returns the standard, production-safe configuration for the World service.
func DefaultConfig() Config {
	return Config{
		EnableTracing:          true,
		EnableSummary:          true,
		MaxPendingInteractions: 1024,
		DefaultSessionID:       "default-session",
		ShutdownDrainTimeout:   10 * time.Second,
	}
}

// Validate verifies that operational configuration parameters are structurally valid.
func (c Config) Validate() error {
	if c.MaxPendingInteractions < 0 {
		return errors.New("world: MaxPendingInteractions cannot be negative")
	}
	if c.ShutdownDrainTimeout < 0 {
		return errors.New("world: ShutdownDrainTimeout cannot be negative")
	}
	if c.DefaultSessionID == "" {
		return errors.New("world: DefaultSessionID cannot be empty")
	}
	return nil
}

// Option defines a functional option for modifying World service Config.
type Option func(*Config)

// WithTracing configures whether WorldTrace artifacts are recorded.
func WithTracing(enabled bool) Option {
	return func(c *Config) {
		c.EnableTracing = enabled
	}
}

// WithSummary configures whether WorldSummary statistics are maintained.
func WithSummary(enabled bool) Option {
	return func(c *Config) {
		c.EnableSummary = enabled
	}
}

// WithMaxPendingInteractions sets the maximum number of concurrent in-flight interactions.
func WithMaxPendingInteractions(max int) Option {
	return func(c *Config) {
		c.MaxPendingInteractions = max
	}
}

// WithDefaultSessionID configures the default session ID for interactions without an explicit session.
func WithDefaultSessionID(id string) Option {
	return func(c *Config) {
		c.DefaultSessionID = id
	}
}

// WithShutdownDrainTimeout sets the maximum drain time during graceful shutdown.
func WithShutdownDrainTimeout(d time.Duration) Option {
	return func(c *Config) {
		c.ShutdownDrainTimeout = d
	}
}
