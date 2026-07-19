package realization

import "idun/intelligence/infrastructure/inference"

// Builder implements a functional builder pattern for clean service initialization.
type Builder struct {
	ws     WorkspaceSubscriberPublisher
	inf    inference.InferenceService
	storer PayloadStorer
	cfg    Config
}

// NewServiceBuilder initializes a new builder with default configuration.
func NewServiceBuilder() *Builder {
	return &Builder{cfg: DefaultConfig()}
}

func (b *Builder) WithWorkspace(ws WorkspaceSubscriberPublisher) *Builder {
	b.ws = ws
	return b
}

func (b *Builder) WithInference(inf inference.InferenceService) *Builder {
	b.inf = inf
	return b
}

func (b *Builder) WithStorage(storer PayloadStorer) *Builder {
	b.storer = storer
	return b
}

func (b *Builder) WithConfig(cfg Config) *Builder {
	b.cfg = cfg
	return b
}

func (b *Builder) Build() (*Service, error) {
	return NewService(b.ws, b.inf, b.storer, b.cfg)
}
