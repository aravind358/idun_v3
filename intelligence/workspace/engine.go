package workspace

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"

	"idun/intelligence/communication"
)

const defaultBufferLimit = 500

// Engine is the concrete, thread-safe implementation of the Global Workspace & Leveled Blackboard.
type Engine struct {
	mu             sync.RWMutex
	closed         bool
	bufferLimit    int
	subs           map[SubscriptionID]*subscription
	topicSubs      map[communication.TopicID]map[SubscriptionID]*subscription
	envelopes      map[string]communication.Envelope
	topicBuffers   map[communication.TopicID][]communication.Envelope
	pending        map[communication.TopicID][]PendingCandidate
	depGraph       map[string][]string
	resolvedDeps   map[string]bool
	hierarchyGraph map[string][]string
}

// Option configures functional options for Engine construction.
type Option func(*Engine)

// WithBufferLimit configures the per-topic retention capacity limit.
func WithBufferLimit(limit int) Option {
	return func(e *Engine) {
		if limit > 0 {
			e.bufferLimit = limit
		}
	}
}

// NewEngine constructs a new Global Workspace & Leveled Blackboard engine.
func NewEngine(opts ...Option) *Engine {
	e := &Engine{
		bufferLimit:    defaultBufferLimit,
		subs:           make(map[SubscriptionID]*subscription),
		topicSubs:      make(map[communication.TopicID]map[SubscriptionID]*subscription),
		envelopes:      make(map[string]communication.Envelope),
		topicBuffers:   make(map[communication.TopicID][]communication.Envelope),
		pending:        make(map[communication.TopicID][]PendingCandidate),
		depGraph:       make(map[string][]string),
		resolvedDeps:   make(map[string]bool),
		hierarchyGraph: make(map[string][]string),
	}
	for _, topic := range communication.AllTopics() {
		e.topicSubs[topic] = make(map[SubscriptionID]*subscription)
		e.topicBuffers[topic] = make([]communication.Envelope, 0, 64)
		e.pending[topic] = make([]PendingCandidate, 0, 16)
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Name returns the canonical Kernel component name.
func (e *Engine) Name() string {
	return "Intelligence.Workspace"
}

// Start boots the Workspace engine.
func (e *Engine) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return ErrWorkspaceClosed
	}
	return nil
}

// Close gracefully shuts down the Workspace and cancels all active subscriptions.
func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return nil
	}
	e.closed = true

	for _, sub := range e.subs {
		sub.markClosed()
	}
	e.subs = make(map[SubscriptionID]*subscription)
	for topic := range e.topicSubs {
		e.topicSubs[topic] = make(map[SubscriptionID]*subscription)
	}
	e.pending = make(map[communication.TopicID][]PendingCandidate)
	return nil
}

// Publish validates and posts an Envelope to its leveled topic channel.
func (e *Engine) Publish(ctx context.Context, env communication.Envelope, opts ...PublishOption) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := env.Validate(); err != nil {
		return ErrInvalidEnvelope
	}

	cfg := &publishConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	e.mu.Lock()
	if e.closed {
		e.mu.Unlock()
		return ErrWorkspaceClosed
	}

	// Buffer envelope
	e.envelopes[env.ID] = env
	buf := e.topicBuffers[env.Topic]
	buf = append(buf, env)
	if len(buf) > e.bufferLimit {
		// Evict oldest envelope from topic buffer
		buf = buf[len(buf)-e.bufferLimit:]
	}
	e.topicBuffers[env.Topic] = buf

	// Collect snapshot of subscriber handlers under lock
	var handlers []EnvelopeHandler
	if cfg.broadcastGlobal {
		for _, topicMap := range e.topicSubs {
			for _, sub := range topicMap {
				handlers = append(handlers, sub.handler)
			}
		}
	} else {
		if topicMap, ok := e.topicSubs[env.Topic]; ok {
			for _, sub := range topicMap {
				handlers = append(handlers, sub.handler)
			}
		}
	}
	e.mu.Unlock()

	// Dispatch to subscribers without holding lock
	for _, h := range handlers {
		_ = h(ctx, env)
	}

	return nil
}

// Subscribe subscribes a cognitive ability to a specific leveled topic channel.
func (e *Engine) Subscribe(topic communication.TopicID, subscriberID string, handler EnvelopeHandler) (Subscription, error) {
	if !topic.IsValid() {
		return nil, ErrInvalidTopic
	}
	if subscriberID == "" {
		return nil, ErrSubscriberIDMissing
	}
	if handler == nil {
		return nil, ErrHandlerMissing
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return nil, ErrWorkspaceClosed
	}

	id := SubscriptionID("sub-" + generateHexID(8))
	sub := &subscription{
		id:         id,
		topic:      topic,
		subscriber: subscriberID,
		handler:    handler,
		engine:     e,
	}

	e.subs[id] = sub
	if e.topicSubs[topic] == nil {
		e.topicSubs[topic] = make(map[SubscriptionID]*subscription)
	}
	e.topicSubs[topic][id] = sub

	return sub, nil
}

// SubscribeAll subscribes a monitor to all canonical topics.
func (e *Engine) SubscribeAll(subscriberID string, handler EnvelopeHandler) ([]Subscription, error) {
	if subscriberID == "" {
		return nil, ErrSubscriberIDMissing
	}
	if handler == nil {
		return nil, ErrHandlerMissing
	}

	topics := communication.AllTopics()
	subs := make([]Subscription, 0, len(topics))
	for _, topic := range topics {
		sub, err := e.Subscribe(topic, subscriberID, handler)
		if err != nil {
			return subs, err
		}
		subs = append(subs, sub)
	}
	return subs, nil
}

// Unsubscribe cancels an active subscription by ID.
func (e *Engine) Unsubscribe(id SubscriptionID) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	sub, ok := e.subs[id]
	if !ok {
		return ErrSubscriptionClosed
	}

	delete(e.subs, id)
	if topicMap, ok := e.topicSubs[sub.topic]; ok {
		delete(topicMap, id)
	}
	sub.markClosed()
	return nil
}

// GetEnvelope retrieves a recently published envelope by ID.
func (e *Engine) GetEnvelope(id string) (communication.Envelope, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	env, ok := e.envelopes[id]
	return env, ok
}

// ListTopicEnvelopes retrieves up to limit recent envelopes buffered on a specific topic.
func (e *Engine) ListTopicEnvelopes(topic communication.TopicID, limit int) []communication.Envelope {
	e.mu.RLock()
	defer e.mu.RUnlock()

	buf, ok := e.topicBuffers[topic]
	if !ok || len(buf) == 0 {
		return nil
	}

	n := len(buf)
	if limit <= 0 || limit > n {
		limit = n
	}

	out := make([]communication.Envelope, limit)
	start := n - limit
	copy(out, buf[start:])
	return out
}

// StorePendingCandidate stores a candidate bid envelope pending competition/arbitration on a topic.
func (e *Engine) StorePendingCandidate(ctx context.Context, topic communication.TopicID, candidate PendingCandidate) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if !topic.IsValid() {
		return ErrInvalidTopic
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return ErrWorkspaceClosed
	}

	list := e.pending[topic]
	list = append(list, candidate)
	e.pending[topic] = list
	return nil
}

// GetPendingCandidates retrieves all currently pending candidate bids on a specific topic channel.
func (e *Engine) GetPendingCandidates(topic communication.TopicID) []PendingCandidate {
	e.mu.RLock()
	defer e.mu.RUnlock()

	list, ok := e.pending[topic]
	if !ok || len(list) == 0 {
		return nil
	}
	out := make([]PendingCandidate, len(list))
	copy(out, list)
	return out
}

// RemovePendingCandidate removes a specific candidate bid by Envelope ID from pending state after arbitration.
func (e *Engine) RemovePendingCandidate(topic communication.TopicID, envelopeID string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	list, ok := e.pending[topic]
	if !ok || len(list) == 0 {
		return false
	}

	for i, item := range list {
		if item.Envelope.ID == envelopeID {
			e.pending[topic] = append(list[:i], list[i+1:]...)
			return true
		}
	}
	return false
}


// subscription represents a concrete subscription handle.
type subscription struct {
	mu         sync.Mutex
	id         SubscriptionID
	topic      communication.TopicID
	subscriber string
	handler    EnvelopeHandler
	engine     *Engine
	closed     bool
}

func (s *subscription) ID() SubscriptionID {
	return s.id
}

func (s *subscription) Topic() communication.TopicID {
	return s.topic
}

func (s *subscription) Subscriber() string {
	return s.subscriber
}

func (s *subscription) Cancel() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	s.mu.Unlock()
	return s.engine.Unsubscribe(s.id)
}

func (s *subscription) markClosed() {
	s.mu.Lock()
	s.closed = true
	s.mu.Unlock()
}

func generateHexID(n int) string {
	buf := make([]byte, n)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

func (e *Engine) RegisterEpisodeDependencies(ctx context.Context, epID string, dependsOn []string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return ErrWorkspaceClosed
	}
	e.depGraph[epID] = append(e.depGraph[epID], dependsOn...)
	return nil
}

func (e *Engine) IsEpisodeReady(ctx context.Context, epID string) (bool, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.closed {
		return false, ErrWorkspaceClosed
	}
	deps, exists := e.depGraph[epID]
	if !exists || len(deps) == 0 {
		return true, nil
	}
	for _, dep := range deps {
		if !e.resolvedDeps[dep] {
			// Also check if dep is an envelope currently present in envelopes
			if _, ok := e.envelopes[dep]; !ok {
				return false, nil
			}
		}
	}
	return true, nil
}

func (e *Engine) ResolveDependencies(ctx context.Context, epID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return ErrWorkspaceClosed
	}
	deps, exists := e.depGraph[epID]
	if !exists {
		return nil
	}
	for _, dep := range deps {
		if _, ok := e.envelopes[dep]; ok {
			e.resolvedDeps[dep] = true
		}
	}
	return nil
}

func (e *Engine) NotifyDependencyComplete(ctx context.Context, depID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return ErrWorkspaceClosed
	}
	e.resolvedDeps[depID] = true
	return nil
}

func (e *Engine) RegisterEpisodeChild(ctx context.Context, parentID string, childID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return ErrWorkspaceClosed
	}
	e.hierarchyGraph[parentID] = append(e.hierarchyGraph[parentID], childID)
	return nil
}

func (e *Engine) GetEpisodeChildren(ctx context.Context, parentID string) ([]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.closed {
		return nil, ErrWorkspaceClosed
	}
	children := e.hierarchyGraph[parentID]
	out := make([]string, len(children))
	copy(out, children)
	return out, nil
}

// Ensure Engine implements Workspace.
var _ Workspace = (*Engine)(nil)
