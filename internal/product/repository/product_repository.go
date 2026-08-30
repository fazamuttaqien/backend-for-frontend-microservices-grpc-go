package repository

import (
	"context"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/product/domain"
)

type ProductRepository interface {
	Create(context.Context, *domain.Product) error
	GetByID(context.Context, string) (*domain.Product, error)
	List(context.Context, int, int) ([]*domain.Product, int, error)
	Update(context.Context, *domain.Product) error
}
