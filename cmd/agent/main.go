package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/ospiem/mcollector/internal/agent"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	agent.Run(ctx, &wg)
	handleSignals(cancel)

	wg.Wait()
	fmt.Println("Graceful shutdown complete")
}

func handleSignals(cancel context.CancelFunc) {

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	sig := <-sigCh
	fmt.Printf("Recieved signal: %v\n", sig)

	cancel()
}
