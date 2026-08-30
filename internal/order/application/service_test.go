package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockRepo struct { created *domain.Order; orders []*domain.Order; err error }
func (m *mockRepo) Create(_ context.Context, o *domain.Order) error { m.created = o; return m.err }
func (m *mockRepo) GetByID(_ context.Context, _ string) (*domain.Order, error) { if m.err != nil { return nil, m.err }; if len(m.orders) == 0 { return nil, domain.ErrOrderNotFound }; return m.orders[0], nil }
func (m *mockRepo) List(_ context.Context, _, _ int) ([]*domain.Order, int, error) { return m.orders, len(m.orders), m.err }

type mockUser struct { err error }
func (m mockUser) GetUser(context.Context, string) error { return m.err }

type mockProduct struct { price string; err error }
func (m mockProduct) GetProduct(context.Context, string) (string, error) { return m.price, m.err }

func TestCreate(t *testing.T) {
	repo := &mockRepo{}
	svc := NewOrderService(repo, mockUser{}, mockProduct{price: "12.5000"})
	svc.now = func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }
	o, err := svc.Create(context.Background(), "user-1", []CreateItem{{ProductID: "product-1", Quantity: 2}})
	if err != nil { t.Fatal(err) }
	if o.Total != "25.0000" || o.Status != "PENDING" { t.Fatalf("unexpected order: %+v", o) }
	if len(o.Items) != 1 || o.Items[0].Price != "12.5000" || o.Items[0].Total != "25.0000" { t.Fatalf("unexpected item: %+v", o.Items) }
}

func TestCreateUserNotFound(t *testing.T) {
	repo := &mockRepo{}
	svc := NewOrderService(repo, mockUser{err: status.Error(codes.NotFound, "user not found")}, mockProduct{price: "10"})
	_, err := svc.Create(context.Background(), "user-1", []CreateItem{{ProductID: "p-1", Quantity: 1}})
	if !errors.Is(err, domain.ErrUserNotFound) { t.Fatalf("expected user not found, got %v", err) }
}

func TestCreateProductNotFound(t *testing.T) {
	repo := &mockRepo{}
	svc := NewOrderService(repo, mockUser{}, mockProduct{err: status.Error(codes.NotFound, "product not found")})
	_, err := svc.Create(context.Background(), "user-1", []CreateItem{{ProductID: "p-1", Quantity: 1}})
	if !errors.Is(err, domain.ErrProductNotFound) { t.Fatalf("expected product not found, got %v", err) }
}

func TestCreateCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	svc := NewOrderService(&mockRepo{}, mockUser{}, mockProduct{price: "10"})
	_, err := svc.Create(ctx, "user-1", []CreateItem{{ProductID: "p-1", Quantity: 1}})
	if !errors.Is(err, context.Canceled) { t.Fatalf("expected canceled, got %v", err) }
}

func TestListDefaultsAndGet(t *testing.T) {
	repo := &mockRepo{orders: []*domain.Order{{ID: "o-1"}}}
	svc := NewOrderService(repo, mockUser{}, mockProduct{})
	orders, total, err := svc.List(context.Background(), 0, 0)
	if err != nil || total != 1 || len(orders) != 1 { t.Fatalf("unexpected list result: %v %d %d", err, total, len(orders)) }
	o, err := svc.Get(context.Background(), "o-1")
	if err != nil || o.ID != "o-1" { t.Fatalf("unexpected get result: %+v %v", o, err) }
}
