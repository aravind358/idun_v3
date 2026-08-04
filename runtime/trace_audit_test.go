package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"idun/intelligence/communication"
)

func TestTraceAudit(t *testing.T) {
	inputs := []string{
		"hi",
		"hello",
		"hi bro",
		"bro",
		"what time is it",
		"3+3",
		"tell me a joke",
		"create a reminder for tomorrow",
		"take a note saying buy milk",
		"what is the weather today",
	}

	cfg := DefaultConfiguration()
	cfg.EnableLogging = false

	for _, inputStr := range inputs {
		fmt.Printf("\n\n## Trace: %q\n", inputStr)
		
		inputBuf := bytes.NewBufferString(inputStr + "\nexit\n")
		outBuf := &syncBuffer{}

		h, err := NewHost(cfg, WithIOReaders(inputBuf, outBuf))
		if err != nil {
			t.Fatalf("Failed to create host: %v", err)
		}

		if err := h.Build(); err != nil {
			t.Fatalf("Failed to build host: %v", err)
		}

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
		
		ws.Subscribe(communication.TopicUserIntent, "trace", func(ctx context.Context, env communication.Envelope) error {
			printPayload(communication.TopicUserIntent, env)
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
		var once sync.Once
		ws.Subscribe(communication.TopicActionExecution, "trace", func(ctx context.Context, env communication.Envelope) error {
			printPayload(communication.TopicActionExecution, env)
			once.Do(func() {
				close(doneCh)
			})
			return nil
		})

		ctx, cancel := context.WithCancel(context.Background())

		if err := h.Start(ctx); err != nil {
			t.Fatalf("Failed to start host: %v", err)
		}

		timeout := time.After(15 * time.Second)
		select {
		case <-doneCh:
			time.Sleep(100 * time.Millisecond) // flush
		case <-timeout:
			fmt.Printf("\n========== Trace timed out after 15 seconds ==========\n")
		}

		h.Stop()
		cancel()
	}
}
