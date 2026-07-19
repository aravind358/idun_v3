package realization

import (
	"os"
	"time"
)

// Config defines the lightweight execution settings for the Language Realization Layer.
type Config struct {
	// ModelID is the logical model identifier resolved via ModelRegistry (e.g., "local-realizer").
	ModelID string `json:"model_id"`
	// Temperature controls phrasing variance (must be low: 0.1 to 0.2 to prevent hallucination).
	Temperature float64 `json:"temperature"`
	// Timeout is the maximum SLA duration allowed for surface phrasing.
	Timeout time.Duration `json:"timeout"`
	// MaxOutputLength restricts the token or character length of the generated response.
	MaxOutputLength int `json:"max_output_length"`
	// BypassCache indicates whether to bypass exact content-addressed memoization for surface realization.
	BypassCache bool `json:"bypass_cache"`
}

// DefaultConfig returns production-ready defaults for surface realization.
func DefaultConfig() Config {
	return Config{
		ModelID:         "local-realizer",
		Temperature:     0.15,
		Timeout:         180 * time.Second,
		MaxOutputLength: 512,
		BypassCache:     os.Getenv("IDUN_REALIZATION_BYPASS_CACHE") == "true",
	}
}
