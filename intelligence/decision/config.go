package decision

// DecisionConfig defines the declarative runtime configuration for the Decision Subsystem.
type DecisionConfig struct {
	StrategyVersion             string `json:"strategy_version"`
	MaxReflexiveLatencyUs       uint32 `json:"max_reflexive_latency_us"`
	EnableParetoFrontier        bool   `json:"enable_pareto_frontier"`
	EnableConfidenceCalibration bool   `json:"enable_confidence_calibration"`
}

// DefaultDecisionConfig returns canonical default settings for Decision Subsystem.
func DefaultDecisionConfig() DecisionConfig {
	return DecisionConfig{
		StrategyVersion:             "v2.0.0-FROZEN",
		MaxReflexiveLatencyUs:       2000,
		EnableParetoFrontier:        true,
		EnableConfidenceCalibration: true,
	}
}
