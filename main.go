// main.go
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/koddr/tutorial-go-asynq/client"
	"github.com/koddr/tutorial-go-asynq/worker"
)

var (
	redisAddr string
	mode      string
)

func init() {
	flag.StringVar(&redisAddr, "redis", "127.0.0.1:6379", "redis address")
	flag.StringVar(&mode, "mode", "both", "mode to run the service in (worker, client, or both)")
	flag.Parse()
}

func main() {
	// Create a channel to receive OS signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Start worker and/or client based on mode
	switch mode {
	case "worker":
		go worker.Start(redisAddr)
	case "client":
		go client.Start(redisAddr)
	case "both":
		go worker.Start(redisAddr)
		go client.Start(redisAddr)
	default:
		fmt.Println("Invalid mode. Use 'worker', 'client', or 'both'")
		os.Exit(1)
	}

	// Wait for OS signal
	<-sigs
	fmt.Println("Shutting down...")
}
