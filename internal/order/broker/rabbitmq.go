package broker

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	OrderEventsExchange = "order.events"
	OrderCreatedKey     = "order.created"
)

type Publisher struct {
	conn *amqp.Connection
	url  string
}

func NewPublisher(url string) (*Publisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}
	defer ch.Close()
	if err := ch.ExchangeDeclare(OrderEventsExchange, amqp.ExchangeTopic, true, false, false, false, nil); err != nil {
		conn.Close()
		return nil, err
	}
	return &Publisher{conn: conn, url: url}, nil
}

func (p *Publisher) Publish(ctx context.Context, routingKey string, body []byte) error {
	ch, err := p.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	if err := ch.ExchangeDeclare(OrderEventsExchange, amqp.ExchangeTopic, true, false, false, false, nil); err != nil {
		return err
	}
	return ch.PublishWithContext(ctx, OrderEventsExchange, routingKey, false, false, amqp.Publishing{
		ContentType: "application/json",
		DeliveryMode: amqp.Persistent,
		Body: body,
	})
}

func PublishWithRetry(ctx context.Context, publisher *Publisher, routingKey string, body []byte, attempts int) error {
	var err error
	for i := 0; i < attempts; i++ {
		attemptCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		err = publisher.Publish(attemptCtx, routingKey, body)
		cancel()
		if err == nil {
			return nil
		}
		if i+1 < attempts {
			timer := time.NewTimer(time.Duration(i+1) * 100 * time.Millisecond)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}
	}
	return fmt.Errorf("publish event after %d attempts: %w", attempts, err)
}

func (p *Publisher) Close() error { return p.conn.Close() }
