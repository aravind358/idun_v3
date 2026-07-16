package runtime

import (
	"context"
	"idun/kernel"
)

// ComponentWrapper bridges concrete cognitive subsystem instances to the Kernel
// lifecycle and phased boot infrastructure without modifying frozen interfaces.
type ComponentWrapper struct {
	name      string
	phase     kernel.Phase
	startFunc func(ctx context.Context) error
	closeFunc func() error
	instance  interface{}
}

// NewWrapper constructs a ComponentWrapper for any subsystem or service.
func NewWrapper(name string, phase kernel.Phase, instance interface{}, startFunc func(ctx context.Context) error, closeFunc func() error) *ComponentWrapper {
	return &ComponentWrapper{
		name:      name,
		phase:     phase,
		startFunc: startFunc,
		closeFunc: closeFunc,
		instance:  instance,
	}
}

// Name implements kernel.Component.
func (w *ComponentWrapper) Name() string {
	return w.name
}

// BootPhase implements kernel.Phased.
func (w *ComponentWrapper) BootPhase() kernel.Phase {
	return w.phase
}

// Start implements kernel.Lifecycle.
func (w *ComponentWrapper) Start(ctx context.Context) error {
	if w.startFunc != nil {
		return w.startFunc(ctx)
	}
	// Fall back to checking if instance implements Start(context.Context) error or Start() error
	if lcCtx, ok := w.instance.(interface{ Start(context.Context) error }); ok {
		return lcCtx.Start(ctx)
	}
	if lcNoCtx, ok := w.instance.(interface{ Start() error }); ok {
		return lcNoCtx.Start()
	}
	return nil
}

// Close implements kernel.Lifecycle.
func (w *ComponentWrapper) Close() error {
	if w.closeFunc != nil {
		return w.closeFunc()
	}
	if lc, ok := w.instance.(interface{ Close() error }); ok {
		return lc.Close()
	}
	return nil
}

// Instance returns the underlying wrapped subsystem instance.
func (w *ComponentWrapper) Instance() interface{} {
	return w.instance
}
