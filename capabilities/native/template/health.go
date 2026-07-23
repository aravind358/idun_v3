package template

// Health status checks can be implemented here.
// By default, it inherits from BaseCapability, but specific metrics or provider health checks can be integrated.

func (c *Capability) checkProviderHealth() bool {
	// TODO: Add capability-specific health checking logic here.
	return true
}
