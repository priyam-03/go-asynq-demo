// worker/worker.go
package worker

import (
	"fmt"
	"log"

	"context" // Add this at the top

	"github.com/hibiken/asynq"
	"github.com/koddr/tutorial-go-asynq/tasks"
)

// Start initializes and starts the task worker
func Start(redisAddr string) {
	// Create a new Redis connection for the Asynq server
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			// Specify how many concurrent workers to use
			Concurrency: 10,
			// Specify multiple queues with different priorities
			Queues: map[string]int{
				"critical": 6, // Higher priority
				"default":  3, // Normal priority
				"low":      1, // Lower priority
			},
			// Optionally enable the Asynq Web UI
			// Note: This would require additional setup
			// UseRedisStreams: true,
			// LogLevel: asynq.DebugLevel,
		},
	)

	// Create a new mux server to register task handlers
	mux := asynq.NewServeMux()

	// Register task handlers
	mux.HandleFunc(tasks.TypeEmailDelivery, tasks.HandleEmailDeliveryTask)
	mux.HandleFunc(tasks.TypeEmailWelcome, tasks.HandleEmailWelcomeTask)
	mux.HandleFunc(tasks.TypeImageResize, tasks.HandleImageResizeTask)

	// Additional handler with a wildcard to catch unregistered tasks

	mux.Handle("*", asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		fmt.Printf("Received an unknown task: %s\n", t.Type())
		return nil
	}))

	fmt.Println("Starting worker server...")
	if err := srv.Run(mux); err != nil {
		log.Fatal(err)
	}
}
