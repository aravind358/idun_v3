package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"
	"strings"

	"idun/runtime"
)

func main() {
	cfg := runtime.DefaultConfiguration()
	cfg.EnableLogging = false
	cfg.StoragePath = "data/runtime"
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

	queries := []string{
		"hello", 
		"who are you", 
		"how are you", 
		"bye",
		"hello",
		"hello",
		"hello",
		"who are you",
		"current time",
		"today's date",
		"2+2",
		"weather",
		"shutdown",
	}
	for _, q := range queries {
		fmt.Printf("\n==================================\n")
		fmt.Printf("Query: %s\n", q)
		fmt.Printf("==================================\n")
		
		go func() {
			inW.Write([]byte(q + "\n"))
		}()

		buf := make([]byte, 4096)
		
		timeout := time.After(10 * time.Second)
		readChan := make(chan string)
		
		go func() {
			n, _ := outR.Read(buf)
			readChan <- string(buf[:n])
		}()

		select {
		case out := <-readChan:
			fmt.Printf("Output: %s\n", strings.TrimSpace(out))
		case <-timeout:
			fmt.Printf("Output: [TIMED OUT]\n")
		}
	}

	host.Stop()
}
