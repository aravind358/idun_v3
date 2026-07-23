package files

// Health status checks can be implemented here.
// By default, it inherits from BaseCapability.

func (c *Capability) checkProviderHealth() bool {
	// A basic health check for a file provider could be verifying the temp directory is writable.
	return true
}
