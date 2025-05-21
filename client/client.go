// client/client.go
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log"

	"github.com/hibiken/asynq"
	"github.com/koddr/tutorial-go-asynq/tasks"
	"github.com/redis/go-redis/v9"
)

// Start initializes and starts the task client and scheduler with enhanced Redis handling
func Start(redisAddr string) {
	// Configure Redis client options with robust settings
	redisOpts := asynq.RedisClientOpt{
		Addr:         redisAddr,
		DialTimeout:  15 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		PoolSize:     20,
	}

	// Create initial client and scheduler
	client := asynq.NewClient(redisOpts)
	defer client.Close()

	// Prepare payload once
	userID := 42 + int(time.Now().UnixNano()%1000)
	payloadBytes, err := json.Marshal(tasks.EmailWelcomePayload{UserID: userID})
	if err != nil {
		log.Fatalf("[%s] Failed to marshal payload: %v", time.Now().Format(time.RFC3339), err)
	}

	// Scheduler with auto-restart
	go func() {
		var scheduler *asynq.Scheduler
		for {
			scheduler = asynq.NewScheduler(redisOpts, &asynq.SchedulerOpts{
				LogLevel: asynq.InfoLevel,
			})

			// Re-register task on each restart
			if _, err := scheduler.Register("@every 30s", asynq.NewTask(tasks.TypeEmailWelcome, payloadBytes)); err != nil {
				log.Printf("[%s] Scheduler registration failed: %v", time.Now().Format(time.RFC3339), err)
				time.Sleep(10 * time.Second)
				continue
			}

			log.Printf("[%s] Scheduler starting", time.Now().Format(time.RFC3339))
			if err := scheduler.Run(); err != nil {
				log.Printf("[%s] Scheduler error: %v", time.Now().Format(time.RFC3339), err)
				log.Printf("[%s] Restarting scheduler in 10s...", time.Now().Format(time.RFC3339))
				time.Sleep(10 * time.Second)
			} else {
				break // Clean exit
			}
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Task producer with connection handling
	go func() {
		var retryCount int
		for {
			// Health check
			if err := pingRedis(redisOpts); err != nil {
				log.Printf("[%s] Redis unreachable: %v", time.Now().Format(time.RFC3339), err)
				retryCount++
				time.Sleep(time.Duration(retryCount) * 5 * time.Second)
				continue
			}
			retryCount = 0

			// Enqueue tasks with resilience
			if err := enqueueTasks(client); err != nil {
				log.Printf("[%s] Task enqueue failed: %v", time.Now().Format(time.RFC3339), err)
				client.Close()
				client = asynq.NewClient(redisOpts)
			}

			time.Sleep(15 * time.Second)
		}
	}()

	<-stop
	log.Printf("[%s] Shutting down gracefully", time.Now().Format(time.RFC3339))
}

func pingRedis(opts asynq.RedisClientOpt) error {
	rdb := opts.MakeRedisClient().(*redis.Client)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return rdb.Ping(ctx).Err()
}

// enqueueTasks handles task creation and enqueueing with retries
func enqueueTasks(client *asynq.Client) error {
	// Email task
	emailTask, err := tasks.NewEmailDeliveryTask(123, "welcome_template", map[string]interface{}{
		"subject": "Welcome!",
		"body":    "Excited to have you!",
	})
	if err != nil {
		return fmt.Errorf("email task creation failed: %w", err)
	}
	if _, err := client.Enqueue(emailTask); err != nil {
		return fmt.Errorf("email enqueue failed: %w", err)
	}

	// Image task
	imageTask, err := tasks.NewImageResizeTask("profile.jpg", 800, 600, "jpeg", 123)
	if err != nil {
		return fmt.Errorf("image task creation failed: %w", err)
	}
	if _, err := client.Enqueue(imageTask, asynq.MaxRetry(5), asynq.Queue("critical")); err != nil {
		return fmt.Errorf("image enqueue failed: %w", err)
	}

	// Delayed task
	welcomeTask, err := tasks.NewEmailWelcomeTask(456)
	if err != nil {
		return fmt.Errorf("welcome task creation failed: %w", err)
	}
	if _, err := client.Enqueue(welcomeTask, asynq.ProcessIn(1*time.Minute)); err != nil {
		return fmt.Errorf("welcome enqueue failed: %w", err)
	}

	return nil
}
