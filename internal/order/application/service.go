package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/domain"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/repository"
	"math/big"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserGateway interface {
	GetUser(context.Context, string) error
}

type ProductGateway interface {
	GetProduct(context.Context, string) (string, error)
}

type CreateItem struct {
	ProductID string
	Quantity  int32
}

type OrderService struct {
	repo     repository.OrderRepository
	users    UserGateway
	products ProductGateway
	now      func() time.Time
}

func NewOrderService(repo repository.OrderRepository, users UserGateway, products ProductGateway) *OrderService {
	return &OrderService{repo: repo, users: users, products: products, now: time.Now}
}

func newID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return hex.EncodeToString(b)
}

func normalizeDependencyError(err error, notFound error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	switch status.Code(err) {
	case codes.NotFound:
		return notFound
	case codes.Canceled:
		return context.Canceled
	case codes.DeadlineExceeded:
		return context.DeadlineExceeded
	default:
		return fmt.Errorf("%w: %v", domain.ErrDependencyUnavailable, err)
	}
}

func itemTotal(price string, quantity int32) (string, error) {
	p, ok := new(big.Rat).SetString(price)
	if !ok || p.Sign() < 0 || quantity <= 0 {
		return "", domain.ErrInvalidOrder
	}
	q := new(big.Rat).Mul(p, new(big.Rat).SetInt64(int64(quantity)))
	return q.FloatString(4), nil
}

func (s *OrderService) Create(ctx context.Context, userID string, items []CreateItem) (*domain.Order, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := s.users.GetUser(ctx, userID); err != nil {
		return nil, normalizeDependencyError(err, domain.ErrUserNotFound)
	}

	orderItems := make([]domain.OrderItem, 0, len(items))
	for _, input := range items {
		if input.Quantity <= 0 || input.ProductID == "" {
			return nil, domain.ErrInvalidOrder
		}
		price, err := s.products.GetProduct(ctx, input.ProductID)
		if err != nil {
			return nil, normalizeDependencyError(err, domain.ErrProductNotFound)
		}
		total, err := itemTotal(price, input.Quantity)
		if err != nil {
			return nil, err
		}
		orderItems = append(orderItems, domain.OrderItem{ID: newID(), ProductID: input.ProductID, Quantity: input.Quantity, Price: price, Total: total})
	}

	o, err := domain.NewOrder(newID(), userID, orderItems, s.now())
	if err != nil {
		return nil, err
	}
	if err = s.repo.Create(ctx, o); err != nil {
		return nil, err
	}
	return o, nil
}

func (s *OrderService) Get(ctx context.Context, id string) (*domain.Order, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *OrderService) List(ctx context.Context, page, pageSize int) ([]*domain.Order, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return s.repo.List(ctx, page, pageSize)
}
