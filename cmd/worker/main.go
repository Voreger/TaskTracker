package main

import (
	"GoProjects/TaskTracker/internal/logger"
	"GoProjects/TaskTracker/internal/queue"
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger.Init()
	defer logger.Sync()

	broker, err := queue.NewBroker("amqp://guest:guest@rabbitmq:5672/", "tasks")
	if err != nil {
		logger.Log.Fatal("RabbitMQ error", zap.Error(err))
	}
	defer broker.Close()

	msgs, err := broker.Consume()
	if err != nil {
		logger.Log.Fatal("Consume error", zap.Error(err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		logger.Log.Info("Worker shutting down")
		cancel()
	}()

	logger.Log.Info("Worker started... waiting for jobs")

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("Worker stopped")
			return
		case msg, ok := <-msgs:
			if !ok {
				logger.Log.Info("RabbitMQ channel closed")
				return
			}

			var event queue.EventMessage
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				logger.Log.Warn("unmarshal error", zap.Error(err))
				msg.Nack(false, false)
				continue
			}

			switch event.Type {
			case queue.EventTaskCreated:
				logger.Log.Info("📥 New task created",
					zap.Any("payload", event.Payload),
				)
			case queue.EventTaskUpdated:
				logger.Log.Info("🛠 Task updated",
					zap.Any("payload", event.Payload),
				)
			case queue.EventTaskDeleted:
				logger.Log.Info("❌ Task deleted",
					zap.Any("payload", event.Payload),
				)
			default:
				logger.Log.Warn("⚠ Unknown event",
					zap.String("type", string(event.Type)),
				)
			}

			msg.Ack(false)
		}
	}

}
