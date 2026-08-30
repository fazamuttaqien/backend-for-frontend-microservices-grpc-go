package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/observability"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/broker"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/events"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	logger := observability.NewLogger()
	slog.SetDefault(logger)
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		logger.Error("RABBITMQ_URL is required")
		return
	}
	conn, err := amqp.Dial(url)
	if err != nil {
		logger.Error("connect RabbitMQ", "error", err)
		return
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		logger.Error("open RabbitMQ channel", "error", err)
		return
	}
	defer ch.Close()
	if err := ch.ExchangeDeclare(broker.OrderEventsExchange, amqp.ExchangeTopic, true, false, false, false, nil); err != nil {
		logger.Error("declare order events exchange", "error", err)
		return
	}
	queue, err := ch.QueueDeclare("order.notification", true, false, false, false, nil)
	if err != nil {
		logger.Error("declare notification queue", "error", err)
		return
	}
	if err := ch.QueueBind(queue.Name, broker.OrderCreatedKey, broker.OrderEventsExchange, false, nil); err != nil {
		logger.Error("bind notification queue", "error", err)
		return
	}
	if err := ch.Qos(1, 0, false); err != nil {
		logger.Error("configure RabbitMQ QoS", "error", err)
		return
	}
	messages, err := ch.Consume(queue.Name, "order-notification-consumer", false, false, false, false, nil)
	if err != nil {
		logger.Error("consume notification queue", "error", err)
		return
	}
	logger.Info("notification consumer listening", "queue", queue.Name)
	for delivery := range messages {
		if err := process(delivery.Body); err != nil {
			logger.Error("event processing failed after retries", "error", err)
		}
		if err := delivery.Ack(false); err != nil {
			logger.Error("ack failed", "error", err)
		}
	}
}

func process(body []byte) error {
	var event events.OrderCreated
	if err := json.Unmarshal(body, &event); err != nil {
		return err
	}
	if event.Type != events.OrderCreatedType || event.Version != events.OrderCreatedVersion {
		return nil
	}
	var err error
	for attempt := 1; attempt <= 3; attempt++ {
		err = simulateNotification(context.Background(), event)
		if err == nil {
			return nil
		}
		if attempt < 3 {
			time.Sleep(time.Duration(attempt) * 100 * time.Millisecond)
		}
	}
	return err
}
func simulateNotification(_ context.Context, event events.OrderCreated) error {
	if event.Type == "" {
		return errors.New("invalid event")
	}
	slog.Default().Info("notification simulation: order created", "status", event.Data.Status)
	return nil
}
