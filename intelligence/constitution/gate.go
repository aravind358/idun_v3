package constitution

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"idun/intelligence/communication"
	"idun/intelligence/workspace"
)

// Gate is the concrete, thread-safe implementation of the Constitutional Action Gate.
type Gate struct {
	mu        sync.RWMutex
	closed    bool
	rules     map[string]Rule
	ruleOrder []string
	hmacKey   []byte
}

// Option configures functional options for Gate construction.
type Option func(*Gate)

// WithHMACKey configures a custom cryptographic key for signing approved action tokens.
func WithHMACKey(key []byte) Option {
	return func(g *Gate) {
		if len(key) > 0 {
			g.hmacKey = append([]byte(nil), key...)
		}
	}
}

// NewGate constructs a new Constitutional Action Gate.
func NewGate(opts ...Option) *Gate {
	key := make([]byte, 32)
	_, _ = rand.Read(key)
	g := &Gate{
		rules:   make(map[string]Rule),
		hmacKey: key,
	}
	for _, opt := range opts {
		opt(g)
	}
	// Register foundational safety metadata check by default
	_ = g.RegisterRule(NewSafetyMetadataRule())
	return g
}

// Name returns the canonical Kernel component name.
func (g *Gate) Name() string {
	return "Intelligence.Constitution"
}

// Start boots the Constitutional Action Gate.
func (g *Gate) Start() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return ErrGateClosed
	}
	return nil
}

// Close gracefully shuts down the Action Gate.
func (g *Gate) Close() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil
	}
	g.closed = true
	g.rules = make(map[string]Rule)
	g.ruleOrder = nil
	return nil
}

// RegisterRule registers a new invariant constitutional rule.
func (g *Gate) RegisterRule(rule Rule) error {
	if rule == nil {
		return ErrNilRule
	}
	id := rule.ID()
	if id == "" {
		return ErrRuleIDMissing
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return ErrGateClosed
	}
	if _, exists := g.rules[id]; exists {
		return ErrDuplicateRule
	}

	g.rules[id] = rule
	g.ruleOrder = append(g.ruleOrder, id)
	return nil
}

// ListRules returns all currently registered rule IDs in order of registration.
func (g *Gate) ListRules() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := make([]string, len(g.ruleOrder))
	copy(out, g.ruleOrder)
	return out
}

// EvaluateAction evaluates all registered rules against a candidate Envelope.
func (g *Gate) EvaluateAction(ctx context.Context, env communication.Envelope) (EvaluationResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return EvaluationResult{}, err
	}

	g.mu.RLock()
	if g.closed {
		g.mu.RUnlock()
		return EvaluationResult{}, ErrGateClosed
	}
	rulesSnapshot := make([]Rule, len(g.ruleOrder))
	for i, id := range g.ruleOrder {
		rulesSnapshot[i] = g.rules[id]
	}
	keyCopy := append([]byte(nil), g.hmacKey...)
	g.mu.RUnlock()

	var escalateResult *EvaluationResult

	for _, rule := range rulesSnapshot {
		verdict, reason, err := rule.Evaluate(ctx, env)
		if err != nil {
			return EvaluationResult{}, err
		}
		switch verdict {
		case VerdictVetoed:
			// Veto is immediate terminal reject
			return EvaluationResult{
				EnvelopeID:   env.ID,
				Verdict:      VerdictVetoed,
				Reason:       reason,
				RuleViolated: rule.ID(),
				EvaluatedAt:  time.Now().UTC(),
			}, nil
		case VerdictEscalateToUser:
			if escalateResult == nil {
				escalateResult = &EvaluationResult{
					EnvelopeID:   env.ID,
					Verdict:      VerdictEscalateToUser,
					Reason:       reason,
					RuleViolated: rule.ID(),
					EvaluatedAt:  time.Now().UTC(),
				}
			}
		}
	}

	if escalateResult != nil {
		return *escalateResult, nil
	}

	sig := signEnvelopeToken(keyCopy, env)
	return EvaluationResult{
		EnvelopeID:  env.ID,
		Verdict:     VerdictApproved,
		Signature:   sig,
		EvaluatedAt: time.Now().UTC(),
	}, nil
}

// InterceptAndPublish evaluates and either publishes an approved action to TopicActionExecution
// or diverts/vetoes unconstitutional actions.
func (g *Gate) InterceptAndPublish(ctx context.Context, env communication.Envelope, ws workspace.Workspace) (EvaluationResult, error) {
	if ws == nil {
		return EvaluationResult{}, ErrNilWorkspace
	}

	res, err := g.EvaluateAction(ctx, env)
	if err != nil {
		return EvaluationResult{}, err
	}

	switch res.Verdict {
	case VerdictApproved:
		// Ensure topic is set to ActionExecution
		env.Topic = communication.TopicActionExecution
		if err := ws.Publish(ctx, env); err != nil {
			return res, err
		}
		return res, nil

	case VerdictVetoed:
		// Block action and emit high-urgency alert to TopicValueFlags
		alertEnv, _ := communication.NewEnvelopeBuilder().
			WithSource("Intelligence.Constitution").
			WithTopic(communication.TopicValueFlags).
			WithParentRef(env.ID).
			WithPayloadRef("storage://constitution/veto/" + env.ID).
			WithModality("veto-event").
			WithConfidence(1.0).
			WithUrgency(100).
			Build()
		_ = ws.Publish(ctx, alertEnv, workspace.WithGlobalBroadcast(true))
		return res, ErrActionVetoed

	case VerdictEscalateToUser:
		// Block action and emit inquiry to TopicUserIntent
		inquiryEnv, _ := communication.NewEnvelopeBuilder().
			WithSource("Intelligence.Constitution").
			WithTopic(communication.TopicUserIntent).
			WithParentRef(env.ID).
			WithPayloadRef("storage://constitution/escalate/" + env.ID).
			WithModality("user-escalation").
			WithConfidence(1.0).
			WithUrgency(env.Urgency).
			Build()
		_ = ws.Publish(ctx, inquiryEnv)
		return res, ErrActionEscalation

	default:
		return res, ErrInvalidEnvelope
	}
}

func signEnvelopeToken(key []byte, env communication.Envelope) string {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(env.ID))
	h.Write([]byte(env.Source))
	h.Write([]byte(env.PayloadRef))
	return hex.EncodeToString(h.Sum(nil))
}

// Ensure Gate implements ActionGate.
var _ ActionGate = (*Gate)(nil)
