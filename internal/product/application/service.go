package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/product/domain"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/product/repository"
)

type ProductService struct {
	repo         repository.ProductRepository
	cache        ProductCache
	cacheTimeout time.Duration
	now          func() time.Time
}

func NewProductService(repo repository.ProductRepository, caches ...ProductCache) *ProductService {
	var cache ProductCache
	if len(caches) > 0 {
		cache = caches[0]
	}
	return &ProductService{repo: repo, cache: cache, cacheTimeout: 100 * time.Millisecond, now: time.Now}
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

func (s *ProductService) Create(ctx context.Context, name, description, price string, stock int32) (*domain.Product, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	p, err := domain.NewProduct(newID(), name, description, price, stock, s.now())
	if err != nil {
		return nil, err
	}
	if err = s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ProductService) Get(ctx context.Context, id string) (*domain.Product, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	key := productCacheKey(id)
	if s.cache != nil {
		cacheCtx, cancel := context.WithTimeout(ctx, s.cacheTimeout)
		data, err := s.cache.Get(cacheCtx, key)
		cancel()
		if err == nil {
			p := &domain.Product{}
			if json.Unmarshal(data, p) == nil {
				return p, nil
			}
			deleteCtx, deleteCancel := context.WithTimeout(ctx, s.cacheTimeout)
			_ = s.cache.Delete(deleteCtx, key)
			deleteCancel()
		}
	}
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	s.storeCache(ctx, key, p)
	return p, nil
}

func (s *ProductService) List(ctx context.Context, page, pageSize int) ([]*domain.Product, int, error) {
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

func (s *ProductService) Update(ctx context.Context, id, name, description, price string, stock int32) (*domain.Product, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	if err = p.Update(name, description, price, stock, s.now()); err != nil {
		return nil, err
	}
	if err = s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	s.invalidateCache(ctx, productCacheKey(id))
	return p, nil
}

func productCacheKey(id string) string { return "product:v1:" + id }

func (s *ProductService) storeCache(ctx context.Context, key string, p *domain.Product) {
	if s.cache == nil {
		return
	}
	data, err := json.Marshal(p)
	if err != nil {
		return
	}
	cacheCtx, cancel := context.WithTimeout(ctx, s.cacheTimeout)
	_ = s.cache.Set(cacheCtx, key, data)
	cancel()
}

func (s *ProductService) invalidateCache(ctx context.Context, key string) {
	if s.cache == nil {
		return
	}
	cacheCtx, cancel := context.WithTimeout(ctx, s.cacheTimeout)
	_ = s.cache.Delete(cacheCtx, key)
	cancel()
}
