// client/client.go
package client

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	// "json"
	"log"
	"time"

	"github.com/hibiken/asynq"
	"github.com/koddr/tutorial-go-asynq/tasks"
)

// Start initializes and starts the task client and scheduler
func Start(redisAddr string) {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	defer client.Close()

	scheduler := asynq.NewScheduler(
		asynq.RedisClientOpt{Addr: redisAddr},
		&asynq.SchedulerOpts{},
	)

	// Prepare randomized user ID
	userID := 42 + int(time.Now().UnixNano()%1000)

	// Correctly marshal the payload using struct
	payloadBytes, err := json.Marshal(tasks.EmailWelcomePayload{UserID: userID})
	if err != nil {
		log.Fatalf("failed to marshal welcome payload: %v", err)
	}

	_, err = scheduler.Register("@every 30s", asynq.NewTask(tasks.TypeEmailWelcome, payloadBytes))
	if err != nil {
		log.Fatalf("failed to register scheduled task: %v", err)
	}

	// Start the scheduler in a goroutine
	go func() {
		if err := scheduler.Run(); err != nil {
			log.Fatalf("scheduler error: %v", err)
		}
	}()

	// Graceful shutdown support
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Demo task enqueue loop
	go func() {
		for {
			// --- Email tasks ---
			emailTask, err := tasks.NewEmailDeliveryTask(123, "welcome_template", map[string]interface{}{
				"subject": "Welcome to our platform!",
				"body":    "We are excited to have you onboard.",
			})
			if err != nil {
				log.Fatal(err)
			}

			info, err := client.Enqueue(emailTask)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Enqueued email task: id=%s queue=%s\n", info.ID, info.Queue)

			// --- Image tasks ---
			imageTask, err := tasks.NewImageResizeTask("profile_pic.jpg", 800, 600, "jpeg", 123)
			if err != nil {
				log.Fatal(err)
			}

			info, err = client.Enqueue(imageTask,
				asynq.MaxRetry(5),
				asynq.Timeout(1*time.Minute),
				asynq.Queue("critical"),
			)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Enqueued image task: id=%s queue=%s\n", info.ID, info.Queue)

			// --- Delayed welcome task ---
			welcomeTask, err := tasks.NewEmailWelcomeTask(456)
			if err != nil {
				log.Fatal(err)
			}

			info, err = client.Enqueue(welcomeTask, asynq.ProcessIn(1*time.Minute))
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Enqueued delayed welcome task: id=%s queue=%s\n", info.ID, info.Queue)

			time.Sleep(15 * time.Second)
		}
	}()

	// Wait for shutdown signal
	<-stop
	fmt.Println("Shutting down client and scheduler.")
}
