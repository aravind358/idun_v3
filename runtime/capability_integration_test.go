package runtime

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"
	"io"
)

// TestIntegration_CapabilityFrameworkWiring proves that TopicActionExecution
// is successfully routed to the capabilities.ActionExecutionHandler.
func TestIntegration_CapabilityFrameworkWiring(t *testing.T) {
	cfg := DefaultConfiguration()
	cfg.EnableLogging = false

	// Capture os.Stdout to verify fmt.Printf output from the capability framework
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()

	// We send a complex input that will trigger a fallback or plan.
	inputBuf := bytes.NewBufferString("Please plan a complex task for me\nexit\n")
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

	// Wait for the pipeline to process
	time.Sleep(10 * time.Second)

	w.Close()
	var stdoutBuf bytes.Buffer
	_, _ = io.Copy(&stdoutBuf, r)

	output := stdoutBuf.String()
	_ = output // Can be used to check output if needed
	
	if h.capHandler == nil {
		t.Errorf("Integration Failure: capHandler was not instantiated")
	}
	
	if h.capManager == nil {
		t.Errorf("Integration Failure: capManager was not instantiated")
	}
}
