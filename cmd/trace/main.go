package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"idun/intelligence/communication"
	"idun/runtime"
)

type syncBuffer struct {
	buf bytes.Buffer
}
func (b *syncBuffer) Read(p []byte) (n int, err error) { return b.buf.Read(p) }
func (b *syncBuffer) Write(p []byte) (n int, err error) { return b.buf.Write(p) }

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: trace <input string>")
		return
	}
	inputStr := os.Args[1]

	cfg := runtime.DefaultConfiguration()
	cfg.EnableLogging = false

	inputBuf := bytes.NewBufferString(inputStr + "\nexit\n")
	outBuf := &syncBuffer{}

	h, err := runtime.NewHost(cfg, runtime.WithIOReaders(inputBuf, outBuf))
	if err != nil {
		panic(err)
	}

	if err := h.Build(); err != nil {
		panic(err)
	}

	// We need access to the memory provider to fetch payloads.
	// We'll just read from storage directly since the default Memory provider writes to it.
	store := h.Storage()

	ws := h.Workspace()
	
	printPayload := func(topic communication.TopicID, env communication.Envelope) {
		fmt.Printf("\n========== STAGE: %s ==========\n", topic)
		fmt.Printf("Envelope ID: %s\n", env.ID)
		
		if env.PayloadRef == "" {
			fmt.Printf("Empty PayloadRef.\n")
			return
		}
		
		data, err := store.Read(env.PayloadRef)
		if err != nil {
			fmt.Printf("Failed to read payload: %v\n", err)
			return
		}
		
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, data, "", "  "); err != nil {
			fmt.Printf("Raw Payload:\n%s\n", string(data))
		} else {
			fmt.Printf("Parsed Payload:\n%s\n", prettyJSON.String())
		}
	}

	doneCh := make(chan struct{})
	timeout := time.After(15 * time.Second)

	// Subscribe to all relevant topics
	ws.Subscribe(communication.TopicUserIntent, "trace", func(ctx context.Context, env communication.Envelope) error {
		printPayload(communication.TopicUserIntent, env)
		close(doneCh) // Signal completion for Understanding audit
		return nil
	})
	ws.Subscribe(communication.TopicActiveGoals, "trace", func(ctx context.Context, env communication.Envelope) error {
		printPayload(communication.TopicActiveGoals, env)
		return nil
	})
	ws.Subscribe(communication.TopicCandidatePlans, "trace", func(ctx context.Context, env communication.Envelope) error {
		printPayload(communication.TopicCandidatePlans, env)
		return nil
	})
	ws.Subscribe(communication.TopicEvaluatedOptions, "trace", func(ctx context.Context, env communication.Envelope) error {
		printPayload(communication.TopicEvaluatedOptions, env)
		return nil
	})

	ws.Subscribe(communication.TopicActionExecution, "trace", func(ctx context.Context, env communication.Envelope) error {
		printPayload(communication.TopicActionExecution, env)
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := h.Start(ctx); err != nil {
		panic(err)
	}

	select {
	case <-doneCh:
		// Give it a tiny bit of time to flush prints
		time.Sleep(100 * time.Millisecond)
	case <-timeout:
		fmt.Println("Trace timed out after 15 seconds")
	}

	h.Stop()
}
