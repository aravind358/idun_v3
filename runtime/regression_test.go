package runtime

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"idun/intelligence/communication"
)

// TestRegression_OneExecutionPath proves that a single user input
// produces exactly one execution path through the Planning -> Decision -> Executive pipeline.
func TestRegression_OneExecutionPath(t *testing.T) {
	cfg := DefaultConfiguration()
	cfg.EnableLogging = false
	inputBuf := bytes.NewBufferString("Hello\nexit\n")
	outBuf := &syncBuffer{}

	h, err := NewHost(cfg, WithIOReaders(inputBuf, outBuf))
	if err != nil {
		t.Fatalf("NewHost failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := h.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer h.Stop()

	var candidatePlanCount int
	sub, err := h.Workspace().Subscribe(communication.TopicCandidatePlans, "TestSubscriber", func(_ context.Context, env communication.Envelope) error {
		candidatePlanCount++
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	defer sub.Cancel()

	time.Sleep(10 * time.Second)

	if candidatePlanCount != 1 {
		t.Errorf("Regression Failure: expected exactly 1 TopicCandidatePlans envelope for one input, got %d", candidatePlanCount)
	}
}

// TestRegression_DeliberativeWorkerWiring proves that the Deliberative Worker
// is properly wired into the runtime and is invoked when a low-confidence input is received.
func TestRegression_DeliberativeWorkerWiring(t *testing.T) {
	cfg := DefaultConfiguration()
	cfg.EnableLogging = false

	// Capture os.Stdout to verify devLog output for deliberative-parser
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()

	inputBuf := bytes.NewBufferString("utterance that no local specialist recognizes\nexit\n")
	outBuf := &syncBuffer{}

	h, err := NewHost(cfg, WithIOReaders(inputBuf, outBuf))
	if err != nil {
		t.Fatalf("NewHost failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := h.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer h.Stop()

	// Wait for the pipeline to process the complex input
	time.Sleep(10 * time.Second)

	w.Close()
	var stdoutBuf bytes.Buffer
	_, _ = io.Copy(&stdoutBuf, r)

	if !strings.Contains(stdoutBuf.String(), "deliberative-parser") {
		t.Errorf("Regression Failure: expected Deliberative Worker to be invoked and request deliberative-parser")
	}
}

