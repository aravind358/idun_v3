package reasoning

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidConfig = errors.New("reasoning: invalid configuration")
)

// Config holds operational parameters for the Reasoning subsystem.
type Config struct {
	DefaultTimeout      time.Duration `json:"default_timeout"`
	MaxBeamWidth        int           `json:"max_beam_width"`
	EscalationThreshold float64       `json:"escalation_threshold"`
	MaxGraphNodes       int           `json:"max_graph_nodes"`
	MaxGraphEdges       int           `json:"max_graph_edges"`
	MaxGraphDepth       int           `json:"max_graph_depth"`
}

// Validate verifies configuration invariants.
func (c Config) Validate() error {
	if c.DefaultTimeout <= 0 {
		return fmt.Errorf("%w: default timeout must be positive", ErrInvalidConfig)
	}
	if c.MaxBeamWidth <= 0 || c.MaxBeamWidth > MaxBeamWidth {
		return fmt.Errorf("%w: max beam width %d must be in range [1, %d]", ErrInvalidConfig, c.MaxBeamWidth, MaxBeamWidth)
	}
	if c.EscalationThreshold < 0.0 || c.EscalationThreshold > 1.0 {
		return fmt.Errorf("%w: escalation threshold %f out of bounds [0.0, 1.0]", ErrInvalidConfig, c.EscalationThreshold)
	}
	if c.MaxGraphNodes <= 0 || c.MaxGraphEdges <= 0 || c.MaxGraphDepth <= 0 {
		return fmt.Errorf("%w: max graph nodes/edges/depth must be positive", ErrInvalidConfig)
	}
	return nil
}

// DefaultConfig returns the canonical production configuration for Reasoning.
func DefaultConfig() Config {
	return Config{
		DefaultTimeout:      5 * time.Second,
		MaxBeamWidth:        MaxBeamWidth,
		EscalationThreshold: 0.45,
		MaxGraphNodes:       500,
		MaxGraphEdges:       2000,
		MaxGraphDepth:       3,
	}
}

// ConfigOption mutates a Config instance.
type ConfigOption func(*Config)

// WithTimeout overrides default operation timeout.
func WithTimeout(timeout time.Duration) ConfigOption {
	return func(c *Config) {
		c.DefaultTimeout = timeout
	}
}

// WithEscalationThreshold overrides minimum confidence threshold before LLM escalation.
func WithEscalationThreshold(threshold float64) ConfigOption {
	return func(c *Config) {
		c.EscalationThreshold = threshold
	}
}

// WithGraphLimits overrides session-scoped graph bounds.
func WithGraphLimits(maxNodes, maxEdges, maxDepth int) ConfigOption {
	return func(c *Config) {
		c.MaxGraphNodes = maxNodes
		c.MaxGraphEdges = maxEdges
		c.MaxGraphDepth = maxDepth
	}
}
