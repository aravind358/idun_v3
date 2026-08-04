package core

// AppCapabilityDependencies encapsulates dependencies that are injected into
// Application Capabilities during instantiation.
type AppCapabilityDependencies struct {
	// Resolver is used to resolve and execute Native Capabilities.
	// It is the ONLY mechanism Application capabilities have to interact with the system.
	Resolver NativeCapabilityResolver
}
