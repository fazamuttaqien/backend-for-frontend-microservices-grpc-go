package repository

import (
	"context"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/domain"
)

type OrderRepository interface {
	Create(context.Context, *domain.Order) error
	GetByID(context.Context, string) (*domain.Order, error)
	List(context.Context, int, int) ([]*domain.Order, int, error)
}
