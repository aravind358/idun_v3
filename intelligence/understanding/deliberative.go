package understanding

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"idun/intelligence/communication"
	"idun/intelligence/infrastructure/inference"
	"idun/intelligence/workspace"
)

var (
	ErrInferenceTimeout           = errors.New("understanding: deliberative inference timed out")
	ErrMalformedInferenceResponse = errors.New("understanding: malformed structured response from inference")
	ErrInferenceUnavailable       = errors.New("understanding: inference service unavailable")
	ErrDeliberativeCancelled      = errors.New("understanding: deliberative worker cancelled")
)

const DefaultInferenceTimeout = 5 * time.Second

// DeliberativeWorker implements Phase 4 LLM-assisted semantic interpretation
// as an asynchronous escalation layer invoked only when deterministic and neural
// specialists fail or fall below admission threshold tau.
type DeliberativeWorker struct {
	mu        sync.RWMutex
	inference inference.InferenceService
	ws        workspace.Workspace
	timeout   time.Duration
	sub       workspace.Subscription
	closed    bool
}

// NewDeliberativeWorker constructs an asynchronous Deliberative Understanding Worker.
func NewDeliberativeWorker(infService inference.InferenceService, ws workspace.Workspace, timeout time.Duration) *DeliberativeWorker {
	if timeout <= 0 {
		timeout = DefaultInferenceTimeout
	}
	return &DeliberativeWorker{
		inference: infService,
		ws:        ws,
		timeout:   timeout,
	}
}

// Start boots the Deliberative Worker and subscribes to TopicImpasses if workspace is configured.
func (d *DeliberativeWorker) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return ErrServiceClosed
	}

	if d.ws != nil {
		sub, err := d.ws.Subscribe(communication.TopicImpasses, "Understanding.DeliberativeWorker", d.handleImpasse)
		if err != nil {
			return err
		}
		d.sub = sub
	}
	return nil
}

// Close gracefully shuts down the Deliberative Worker and cancels workspace subscription.
func (d *DeliberativeWorker) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.closed = true
	if d.sub != nil {
		_ = d.sub.Cancel()
		d.sub = nil
	}
	return nil
}

// InterpretDeliberative executes LLM-assisted semantic interpretation via the shared InferenceService.
func (d *DeliberativeWorker) InterpretDeliberative(ctx context.Context, envelopeID, rawText, prior string) (SemanticFrame, error) {
	d.mu.RLock()
	if d.closed {
		d.mu.RUnlock()
		return SemanticFrame{}, ErrServiceClosed
	}
	if d.inference == nil {
		d.mu.RUnlock()
		return SemanticFrame{}, ErrInferenceUnavailable
	}
	timeout := d.timeout
	d.mu.RUnlock()

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := timeoutCtx.Err(); err != nil {
		if errors.Is(err, context.Canceled) {
			return SemanticFrame{}, ErrDeliberativeCancelled
		}
		return SemanticFrame{}, ErrInferenceTimeout
	}

	req := inference.InferenceRequest{
		ModelID:  "deliberative-parser",
		InputRef: rawText,
		Modality: inference.ModalityStructured,
		Budget:   "DELIBERATIVE",
		CallerID: "Understanding.DeliberativeWorker",
	}
	fmt.Printf("DEBUG: InterpretDeliberative calling Execute with req: %+v\n", req)

	res, err := d.inference.Execute(timeoutCtx, req)
	if err != nil {
		fmt.Printf("DEBUG: InterpretDeliberative Execute returned error: %v\n", err)
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, inference.ErrInferenceTimeout) {
			return SemanticFrame{}, ErrInferenceTimeout
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, inference.ErrInferenceCancelled) {
			return SemanticFrame{}, ErrDeliberativeCancelled
		}
		return SemanticFrame{}, fmt.Errorf("%w: %v", ErrInferenceUnavailable, err)
	}

	// Parse structured JSON response from LLM into canonical SemanticFrame using strict decoding
	dec := json.NewDecoder(strings.NewReader(res.OutputRef))
	dec.DisallowUnknownFields()

	var decodedFrame SemanticFrame
	if err := dec.Decode(&decodedFrame); err != nil {
		return SemanticFrame{}, fmt.Errorf("%w: %v", ErrMalformedInferenceResponse, err)
	}
	if dec.More() {
		return SemanticFrame{}, fmt.Errorf("%w: multiple JSON objects or trailing tokens in inference output", ErrMalformedInferenceResponse)
	}

	decodedFrame.FrameVersion = FrameVersion
	if envelopeID != "" {
		decodedFrame.EnvelopeID = envelopeID
	}
	decodedFrame.Status = StatusUnambiguous
	decodedFrame.PrimaryHypothesis.SourceLayer = LayerDeliberativeLLM
	if decodedFrame.PrimaryHypothesis.CalibratedConfidence == 0 {
		decodedFrame.PrimaryHypothesis.CalibratedConfidence = 0.85
	}

	// Strictly pass through Phase 1 canonical Validate() before any publication
	if err := decodedFrame.Validate(); err != nil {
		return SemanticFrame{}, fmt.Errorf("%w: %v", ErrMalformedInferenceResponse, err)
	}

	return decodedFrame, nil
}

func (d *DeliberativeWorker) handleImpasse(ctx context.Context, env communication.Envelope) error {
	rawText := env.PayloadRef
	if rawText == "" {
		rawText = env.ID
	}
	frame, err := d.InterpretDeliberative(ctx, env.ID, rawText, "")
	if err != nil {
		return err
	}

	d.mu.RLock()
	ws := d.ws
	d.mu.RUnlock()

	if ws != nil {
		payloadBytes, _ := json.Marshal(frame)
		pubEnv := communication.Envelope{
			ID:            fmt.Sprintf("delib-%s", env.ID),
			Source:        "Intelligence.Understanding.DeliberativeWorker",
			Topic:         communication.TopicUserIntent,
			RawConfidence: frame.PrimaryHypothesis.CalibratedConfidence,
			PayloadRef:    string(payloadBytes),
			CreatedAt:     time.Now().UTC(),
		}
		return ws.Publish(ctx, pubEnv)
	}
	return nil
}
