package main

import (
	"GoProjects/TaskTracker/internal/queue"
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	broker, err := queue.NewBroker("amqp://guest:guest@rabbitmq:5672/", "tasks")
	if err != nil {
		log.Fatalf("RabbitMQ error: %v", err)
	}
	defer broker.Close()

	msgs, err := broker.Consume()
	if err != nil {
		log.Fatalf("Consume error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("Worker shutting down")
		cancel()
	}()

	log.Println("Worker started... waiting for jobs")

	for {
		select {
		case <-ctx.Done():
			log.Println("Worker stopped")
			return
		case msg, ok := <-msgs:
			if !ok {
				log.Println("RabbitMQ channel closed")
				return
			}

			var event queue.EventMessage
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Println("unmarshal error:", err)
				msg.Nack(false, false)
				continue
			}

			switch event.Type {
			case queue.EventTaskCreated:
				log.Printf("📥 New task created: %+v", event.Payload)
			case queue.EventTaskUpdated:
				log.Printf("🛠 Task updated: %+v", event.Payload)
			case queue.EventTaskDeleted:
				log.Printf("❌ Task deleted: %+v", event.Payload)
			default:
				log.Printf("⚠ Unknown event: %s", event.Type)
			}

			msg.Ack(false)
		}
	}

}
