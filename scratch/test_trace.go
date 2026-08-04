package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"idun/runtime"
)

func main() {
	cfg := runtime.DefaultConfiguration()
	cfg.EnableLogging = false
	cfg.StoragePath = ".idun/storage"
	os.Setenv("IDUN_TEST_MODE", "1")

	inR, inW := io.Pipe()
	outR, outW := io.Pipe()

	host, err := runtime.NewHost(cfg, runtime.WithIOReaders(inR, outW))
	if err != nil {
		fmt.Printf("Init err: %v\n", err)
		return
	}

	host.Build()
	host.Wire()
	host.Register()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	host.Start(ctx)

	time.Sleep(1 * time.Second)

	queries := []string{"Hello", "Who are you?", "How are you?", "Bye"}
	for _, q := range queries {
		fmt.Printf("\nQuery: %s\n", q)
		
		go func() {
			inW.Write([]byte(q + "\n"))
		}()

		buf := make([]byte, 1024)
		n, _ := outR.Read(buf)
		fmt.Printf("Output: %s\n", string(buf[:n]))
	}

	host.Stop()
}
