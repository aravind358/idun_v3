package reflection

import (
	"errors"
	"fmt"
)

// Config defines operational configuration parameters for IDUN Reflection.
type Config struct {
	Enabled                     bool    `json:"enabled"`
	MaxConcurrentReflections    int     `json:"max_concurrent_reflections"`
	DefaultReflectionConfidence float64 `json:"default_reflection_confidence"`
	PeriodicIntervalSeconds     int     `json:"periodic_interval_seconds"`
}

// DefaultConfig returns the standard default configuration for Reflection.
func DefaultConfig() Config {
	return Config{
		Enabled:                     true,
		MaxConcurrentReflections:    4,
		DefaultReflectionConfidence: 0.85,
		PeriodicIntervalSeconds:     3600,
	}
}

// Validate verifies that operational configuration parameters are valid.
func (c Config) Validate() error {
	if c.MaxConcurrentReflections <= 0 {
		return errors.New("reflection: max concurrent reflections must be positive")
	}
	if c.DefaultReflectionConfidence < 0.0 || c.DefaultReflectionConfidence > 1.0 {
		return fmt.Errorf("%w: default reflection confidence %f", ErrInvalidConfidence, c.DefaultReflectionConfidence)
	}
	if c.PeriodicIntervalSeconds < 0 {
		return errors.New("reflection: periodic interval seconds cannot be negative")
	}
	return nil
}

// Option defines a functional option for modifying Reflection Config.
type Option func(*Config)

// WithEnabled configures whether Reflection is enabled.
func WithEnabled(enabled bool) Option {
	return func(c *Config) {
		c.Enabled = enabled
	}
}

// WithMaxConcurrentReflections sets the maximum number of concurrent reflection tasks.
func WithMaxConcurrentReflections(max int) Option {
	return func(c *Config) {
		c.MaxConcurrentReflections = max
	}
}

// WithDefaultReflectionConfidence sets the default reflection confidence baseline.
func WithDefaultReflectionConfidence(conf float64) Option {
	return func(c *Config) {
		c.DefaultReflectionConfidence = conf
	}
}

// WithPeriodicIntervalSeconds sets the periodic reflection schedule interval.
func WithPeriodicIntervalSeconds(seconds int) Option {
	return func(c *Config) {
		c.PeriodicIntervalSeconds = seconds
	}
}
