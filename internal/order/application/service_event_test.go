package application

import (
	"context"
	"testing"

	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/domain"
)

type mockPublisher struct {
	calls int
	err   error
	order *domain.Order
}

func (m *mockPublisher) PublishOrderCreated(_ context.Context, order *domain.Order) error {
	m.calls++
	m.order = order
	return m.err
}

func TestCreatePublishesOrderCreatedAfterPersistence(t *testing.T) {
	repo := &mockRepo{}
	publisher := &mockPublisher{}
	svc := NewOrderService(repo, mockUser{}, mockProduct{price: "10"}, publisher)
	order, err := svc.Create(context.Background(), "user-1", []CreateItem{{ProductID: "p-1", Quantity: 1}})
	if err != nil {
		t.Fatal(err)
	}
	if repo.created != order {
		t.Fatal("expected order to be persisted before publishing")
	}
	if publisher.calls != 1 || publisher.order != order {
		t.Fatalf("unexpected publisher state: %+v", publisher)
	}
}

func TestCreateDoesNotFailWhenEventPublishingFails(t *testing.T) {
	repo := &mockRepo{}
	publisher := &mockPublisher{err: context.DeadlineExceeded}
	svc := NewOrderService(repo, mockUser{}, mockProduct{price: "10"}, publisher)
	order, err := svc.Create(context.Background(), "user-1", []CreateItem{{ProductID: "p-1", Quantity: 1}})
	if err != nil {
		t.Fatalf("order creation should remain successful: %v", err)
	}
	if order == nil || repo.created == nil {
		t.Fatal("expected order to be persisted")
	}
}
