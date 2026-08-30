package application

import (
	"context"

	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/domain"
)

type EventPublisher interface {
	PublishOrderCreated(context.Context, *domain.Order) error
}
