package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/product/application"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/product/domain"
)

type repo struct{ products map[string]*domain.Product }

func (r *repo) Create(_ context.Context, p *domain.Product) error {
	if r.products == nil {
		r.products = map[string]*domain.Product{}
	}
	r.products[p.ID] = p
	return nil
}
func (r *repo) GetByID(_ context.Context, id string) (*domain.Product, error) {
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
		t.Fatalf("Create: expected context.Canceled, got %v", err)
	}
	if _, err := s.Get(ctx, "missing"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Get: expected context.Canceled, got %v", err)
	}
	if _, _, err := s.List(ctx, 1, 20); !errors.Is(err, context.Canceled) {
		t.Fatalf("List: expected context.Canceled, got %v", err)
	}
	if _, err := s.Update(ctx, "missing", "x", "", "10", 1); !errors.Is(err, context.Canceled) {
		t.Fatalf("Update: expected context.Canceled, got %v", err)
	}
}
