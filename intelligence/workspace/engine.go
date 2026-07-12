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
	mu           sync.RWMutex
	closed       bool
	bufferLimit  int
	subs         map[SubscriptionID]*subscription
	topicSubs    map[communication.TopicID]map[SubscriptionID]*subscription
	envelopes    map[string]communication.Envelope
	topicBuffers map[communication.TopicID][]communication.Envelope
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
		bufferLimit:  defaultBufferLimit,
		subs:         make(map[SubscriptionID]*subscription),
		topicSubs:    make(map[communication.TopicID]map[SubscriptionID]*subscription),
		envelopes:    make(map[string]communication.Envelope),
		topicBuffers: make(map[communication.TopicID][]communication.Envelope),
	}
	for _, topic := range communication.AllTopics() {
		e.topicSubs[topic] = make(map[SubscriptionID]*subscription)
		e.topicBuffers[topic] = make([]communication.Envelope, 0, 64)
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

// Ensure Engine implements Workspace.
var _ Workspace = (*Engine)(nil)
