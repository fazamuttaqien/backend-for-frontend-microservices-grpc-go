package broker

import (
	"context"

	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/domain"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/events"
)

type OrderEventPublisher struct{ publisher *Publisher }

func NewOrderEventPublisher(publisher *Publisher) *OrderEventPublisher {
	return &OrderEventPublisher{publisher: publisher}
}

func (p *OrderEventPublisher) PublishOrderCreated(ctx context.Context, order *domain.Order) error {
	event := events.NewOrderCreated(order, "order-created-"+order.ID)
	body, err := event.Marshal()
	if err != nil {
		return err
	}
	return PublishWithRetry(ctx, p.publisher, OrderCreatedKey, body, 3)
}
