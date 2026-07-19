package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"idun/runtime"
)

func main() {
	cfg := runtime.DefaultConfiguration()
	h, err := runtime.NewHost(cfg)
	if err != nil {
		log.Fatalf("failed to initialize runtime host: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := h.Start(ctx); err != nil {
		log.Fatalf("failed to start runtime host: %v", err)
	}

	fmt.Println("=====================================")
	fmt.Println("IDUN V3")
	fmt.Println("Architecture: 2.0.0-FROZEN")
	fmt.Println()
	fmt.Println("Runtime Ready")
	fmt.Println()
	fmt.Println("Type \"exit\" to quit.")
	fmt.Println("=====================================")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sigCh:
	case <-h.Done():
	}

	_ = h.Stop()
}
