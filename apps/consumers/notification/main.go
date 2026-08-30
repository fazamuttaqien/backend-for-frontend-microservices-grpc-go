package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/broker"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/events"
)

func main() {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" { url = "amqp://guest:guest@localhost:5672/" }
	conn, err := amqp.Dial(url)
	if err != nil { log.Fatal(err) }
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil { log.Fatal(err) }
	defer ch.Close()

	if err := ch.ExchangeDeclare(broker.OrderEventsExchange, amqp.ExchangeTopic, true, false, false, false, nil); err != nil { log.Fatal(err) }
	queue, err := ch.QueueDeclare("order.notification", true, false, false, false, nil)
	if err != nil { log.Fatal(err) }
	if err := ch.QueueBind(queue.Name, broker.OrderCreatedKey, broker.OrderEventsExchange, false, nil); err != nil { log.Fatal(err) }
	if err := ch.Qos(1, 0, false); err != nil { log.Fatal(err) }

	messages, err := ch.Consume(queue.Name, "order-notification-consumer", false, false, false, false, nil)
	if err != nil { log.Fatal(err) }
	log.Printf("notification consumer listening queue=%s", queue.Name)
	for delivery := range messages {
		if err := process(delivery.Body); err != nil {
			log.Printf("event processing failed after retries error=%v body=%s", err, delivery.Body)
		}
		if err := delivery.Ack(false); err != nil { log.Printf("ack failed: %v", err) }
	}
}

func process(body []byte) error {
	var event events.OrderCreated
	if err := json.Unmarshal(body, &event); err != nil { return err }
	if event.Type != events.OrderCreatedType || event.Version != events.OrderCreatedVersion { return nil }

	var err error
	for attempt := 1; attempt <= 3; attempt++ {
		err = simulateNotification(context.Background(), event)
		if err == nil { return nil }
		if attempt < 3 { time.Sleep(time.Duration(attempt) * 100 * time.Millisecond) }
	}
	return err
}

func simulateNotification(_ context.Context, event events.OrderCreated) error {
	log.Printf("notification simulation: order created order_id=%s user_id=%s total=%s status=%s", event.Data.OrderID, event.Data.UserID, event.Data.Total, event.Data.Status)
	return nil
}
