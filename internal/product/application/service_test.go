package application_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/product/application"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/product/domain"
)

type repo struct {
	products map[string]*domain.Product
	gets     int
}

func (r *repo) Create(_ context.Context, p *domain.Product) error {
	if r.products == nil {
		r.products = map[string]*domain.Product{}
	}
	r.products[p.ID] = p
	return nil
}
func (r *repo) GetByID(_ context.Context, id string) (*domain.Product, error) {
	r.gets++
	p, ok := r.products[id]
	if !ok {
		return nil, domain.ErrProductNotFound
	}
	return p, nil
}
func (r *repo) List(_ context.Context, page, pageSize int) ([]*domain.Product, int, error) {
	all := make([]*domain.Product, 0, len(r.products))
	for _, p := range r.products {
		all = append(all, p)
	}
	start := (page - 1) * pageSize
	if start > len(all) {
		start = len(all)
	}
	end := start + pageSize
	if end > len(all) {
		end = len(all)
	}
	return all[start:end], len(all), nil
}
func (r *repo) Update(_ context.Context, p *domain.Product) error { r.products[p.ID] = p; return nil }

type cache struct {
	values              map[string][]byte
	getErr              error
	setErr              error
	deleteErr           error
	gets, sets, deletes int
}

func (c *cache) Get(_ context.Context, key string) ([]byte, error) {
	c.gets++
	if c.getErr != nil {
		return nil, c.getErr
	}
	v, ok := c.values[key]
	if !ok {
		return nil, errors.New("cache miss")
	}
	return v, nil
}
func (c *cache) Set(_ context.Context, key string, value []byte) error {
	c.sets++
	if c.setErr != nil {
		return c.setErr
	}
	if c.values == nil {
		c.values = map[string][]byte{}
	}
	c.values[key] = value
	return nil
}
func (c *cache) Delete(_ context.Context, key string) error {
	c.deletes++
	if c.deleteErr != nil {
		return c.deleteErr
	}
	delete(c.values, key)
	return nil
}

func TestCreate(t *testing.T) {
	r := &repo{}
	s := application.NewProductService(r)
	p, err := s.Create(context.Background(), "Keyboard", "Mechanical", "99.90", 10)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "Keyboard" || p.Price != "99.90" || p.Stock != 10 {
		t.Fatal("create failed")
	}
}

func TestGetListAndUpdate(t *testing.T) {
	r := &repo{}
	s := application.NewProductService(r)
	p, err := s.Create(context.Background(), "Old", "Desc", "10", 2)
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.Get(context.Background(), p.ID)
	if err != nil || got.ID != p.ID {
		t.Fatal(err)
	}
	items, total, err := s.List(context.Background(), 1, 20)
	if err != nil || total != 1 || len(items) != 1 {
		t.Fatal("list failed")
	}
	updated, err := s.Update(context.Background(), p.ID, "New", "Updated", "20", 5)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "New" || updated.Price != "20" || updated.Stock != 5 {
		t.Fatal("update failed")
	}
}

func TestGetCacheHitSkipsDatabase(t *testing.T) {
	r := &repo{}
	c := &cache{values: map[string][]byte{}}
	p := &domain.Product{ID: "p1", Name: "Keyboard", Price: "99.90", Stock: 10}
	data, _ := json.Marshal(p)
	c.values["product:v1:p1"] = data
	s := application.NewProductService(r, c)
	got, err := s.Get(context.Background(), "p1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Keyboard" || r.gets != 0 {
		t.Fatalf("cache hit failed: got=%+v dbGets=%d", got, r.gets)
	}
}

func TestGetCacheMissReadsDatabaseAndPopulatesCache(t *testing.T) {
	r := &repo{}
	p := &domain.Product{ID: "p1", Name: "Keyboard", Price: "99.90", Stock: 10}
	_ = r.Create(context.Background(), p)
	c := &cache{values: map[string][]byte{}}
	s := application.NewProductService(r, c)
	got, err := s.Get(context.Background(), p.ID)
	if err != nil || got.ID != p.ID {
		t.Fatal(err)
	}
	if r.gets != 1 || c.gets != 1 || c.sets != 1 {
		t.Fatalf("unexpected cache-aside calls: db=%d get=%d set=%d", r.gets, c.gets, c.sets)
	}
}

func TestGetFallsBackWhenCacheUnavailable(t *testing.T) {
	r := &repo{}
	p := &domain.Product{ID: "p1", Name: "Keyboard", Price: "99.90", Stock: 10}
	_ = r.Create(context.Background(), p)
	c := &cache{getErr: errors.New("redis unavailable")}
	s := application.NewProductService(r, c)
	got, err := s.Get(context.Background(), p.ID)
	if err != nil || got.ID != p.ID {
		t.Fatal(err)
	}
	if r.gets != 1 {
		t.Fatalf("expected database fallback, got %d reads", r.gets)
	}
}

func TestUpdateInvalidatesCacheAfterDatabaseWrite(t *testing.T) {
	r := &repo{}
	p := &domain.Product{ID: "p1", Name: "Old", Price: "10", Stock: 1}
	_ = r.Create(context.Background(), p)
	c := &cache{values: map[string][]byte{"product:v1:p1": []byte("stale")}}
	s := application.NewProductService(r, c)
	if _, err := s.Update(context.Background(), p.ID, "New", "", "20", 2); err != nil {
		t.Fatal(err)
	}
	if c.deletes != 1 {
		t.Fatalf("expected one invalidation, got %d", c.deletes)
	}
	if _, ok := c.values["product:v1:p1"]; ok {
		t.Fatal("cache entry was not invalidated")
	}
}

func TestInvalidProduct(t *testing.T) {
	r := &repo{}
	s := application.NewProductService(r)
	if _, err := s.Create(context.Background(), "", "", "10", 0); err != domain.ErrInvalidProduct {
		t.Fatal("expected invalid product")
	}
	if _, err := s.Create(context.Background(), "x", "", "10", -1); err != domain.ErrInvalidProduct {
		t.Fatal("expected invalid stock")
	}
}

func TestCanceledContextDoesNotCallRepository(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := &repo{}
	s := application.NewProductService(r)
	if _, err := s.Create(ctx, "Keyboard", "Mechanical", "99.90", 10); !errors.Is(err, context.Canceled) {
		t.Fatalf("Create: %v", err)
	}
	if _, err := s.Get(ctx, "missing"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Get: %v", err)
	}
	if _, _, err := s.List(ctx, 1, 20); !errors.Is(err, context.Canceled) {
		t.Fatalf("List: %v", err)
	}
	if _, err := s.Update(ctx, "missing", "x", "", "10", 1); !errors.Is(err, context.Canceled) {
		t.Fatalf("Update: %v", err)
	}
}
