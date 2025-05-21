// worker/worker.go
package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/koddr/tutorial-go-asynq/tasks"
	"github.com/redis/go-redis/v9"
)

// Start initializes and starts the task worker with robust Redis handling
func Start(redisAddr string) {
    redisOpts := asynq.RedisClientOpt{
        Addr:         redisAddr,
        DialTimeout:  15 * time.Second,
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 10 * time.Second,
        PoolSize:     20,
    }

    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

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

            // Create Asynq server
            srv := asynq.NewServer(
                redisOpts,
                asynq.Config{
                    Concurrency: 10,
                    Queues: map[string]int{
                        "critical": 6,
                        "default":  3,
                        "low":      1,
                    },
                },
            )

            // Register handlers
            mux := asynq.NewServeMux()
            mux.HandleFunc(tasks.TypeEmailDelivery, tasks.HandleEmailDeliveryTask)
            mux.HandleFunc(tasks.TypeEmailWelcome, tasks.HandleEmailWelcomeTask)
            mux.HandleFunc(tasks.TypeImageResize, tasks.HandleImageResizeTask)
            mux.Handle("*", asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
                fmt.Printf("Received an unknown task: %s\n", t.Type())
                return nil
            }))

            log.Printf("[%s] Starting worker server...", time.Now().Format(time.RFC3339))
            // Run blocks until error or shutdown
            if err := srv.Run(mux); err != nil {
                log.Printf("[%s] Worker server error: %v", time.Now().Format(time.RFC3339), err)
                log.Printf("[%s] Restarting worker in 10s...", time.Now().Format(time.RFC3339))
                time.Sleep(10 * time.Second)
            } else {
                break // Clean exit
            }
        }
    }()

    <-stop
    log.Printf("[%s] Worker shutting down gracefully", time.Now().Format(time.RFC3339))
}

func pingRedis(opts asynq.RedisClientOpt) error {
    rdb := opts.MakeRedisClient().(*redis.Client)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    return rdb.Ping(ctx).Err()
}