package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/product/domain"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/product/repository"
	"time"
)

type ProductService struct {
	repo repository.ProductRepository
	now  func() time.Time
}

func NewProductService(repo repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo, now: time.Now}
}

func newID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil { return "" }
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return hex.EncodeToString(b)
}

func (s *ProductService) Create(ctx context.Context, name, description, price string, stock int32) (*domain.Product, error) {
	p, err := domain.NewProduct(newID(), name, description, price, stock, s.now())
	if err != nil { return nil, err }
	if err = s.repo.Create(ctx, p); err != nil { return nil, err }
	return p, nil
}

func (s *ProductService) Get(ctx context.Context, id string) (*domain.Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProductService) List(ctx context.Context, page, pageSize int) ([]*domain.Product, int, error) {
	if page < 1 { page = 1 }
	if pageSize < 1 { pageSize = 20 }
	if pageSize > 100 { pageSize = 100 }
	return s.repo.List(ctx, page, pageSize)
}

func (s *ProductService) Update(ctx context.Context, id, name, description, price string, stock int32) (*domain.Product, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil { return nil, err }
	if err = p.Update(name, description, price, stock, s.now()); err != nil { return nil, err }
	if err = s.repo.Update(ctx, p); err != nil { return nil, err }
	return p, nil
}
