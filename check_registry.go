package main

import (
	"bytes"
	"context"
	"fmt"
	"idun/runtime"
	"time"
)

func main() {
	cfg := runtime.DefaultConfiguration()
	cfg.EnableLogging = false
	
	input := "what time is it\nbattery status\ncreate a folder called test\nexit\n"
	inBuf := bytes.NewBufferString(input)
	
	h, _ := runtime.NewHost(cfg, runtime.WithIOReaders(inBuf, nil))
	h.Build()
	
	ctx, cancel := context.WithCancel(context.Background())
	h.Start(ctx)
	
	time.Sleep(2 * time.Second)
	h.Stop()
	cancel()
	fmt.Println("Done")
}
