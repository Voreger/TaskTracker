package queue

import (
	"GoProjects/TaskTracker/internal/logger"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"time"
)

type Broker struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	queue   amqp091.Queue
}

type EventType string

const (
	EventTaskCreated EventType = "task.created"
	EventTaskUpdated EventType = "task.updated"
	EventTaskDeleted EventType = "task.deleted"
)

type EventMessage struct {
	Type    EventType   `json:"type"`    // тип события
	Payload interface{} `json:"payload"` // сами данные
}

func NewBroker(url, queueName string) (*Broker, error) {
	var err error
	var conn *amqp091.Connection

	for i := 1; i <= 20; i++ {
		conn, err = amqp091.Dial(url)
		if err == nil {
			break
		}
		logger.Log.Error("RabbitMQ error", zap.Error(err), zap.Int("retry", i))
		time.Sleep(time.Duration(i) * 200 * time.Millisecond)
	}

	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	q, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	return &Broker{conn: conn, channel: ch, queue: q}, nil
}

func (b *Broker) Publish(body []byte) error {
	return b.channel.Publish(
		"",
		b.queue.Name,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (b *Broker) Consume() (<-chan amqp091.Delivery, error) {
	return b.channel.Consume(
		b.queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
}

func (b *Broker) Close() {
	b.conn.Close()
	b.channel.Close()
}
