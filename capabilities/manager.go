package capabilities

// DefaultManager coordinates the capability framework components.
type DefaultManager struct {
	registry  CapabilityRegistry
	resolver  CapabilityResolver
	lifecycle LifecycleManager
}

func NewManager(registry CapabilityRegistry, resolver CapabilityResolver, lifecycle LifecycleManager) *DefaultManager {
	return &DefaultManager{
		registry:  registry,
		resolver:  resolver,
		lifecycle: lifecycle,
	}
}

func (m *DefaultManager) Registry() CapabilityRegistry {
	return m.registry
}

func (m *DefaultManager) Resolver() CapabilityResolver {
	return m.resolver
}

func (m *DefaultManager) Lifecycle() LifecycleManager {
	return m.lifecycle
}

func (m *DefaultManager) Start() error {
	// Initialize registry, load core capabilities, transition lifecycle
	return nil
}

func (m *DefaultManager) Stop() error {
	// Transition all to Unloaded
	return nil
}
