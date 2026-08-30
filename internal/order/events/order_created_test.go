package events

import (
	"testing"
	"time"

	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/domain"
)

func TestOrderCreatedEnvelope(t *testing.T) {
	now := time.Date(2026, 8, 30, 10, 0, 0, 0, time.UTC)
	order := &domain.Order{ID: "order-1", UserID: "user-1", Total: "25.0000", Status: "PENDING", CreatedAt: now}
	event := NewOrderCreated(order, "event-1")
	if event.Type != OrderCreatedType || event.Version != OrderCreatedVersion {
		t.Fatalf("unexpected event metadata: %+v", event)
	}
	if event.EventID != "event-1" || event.Data.OrderID != order.ID {
		t.Fatalf("unexpected event data: %+v", event.Data)
	}
	body, err := event.Marshal()
	if err != nil || len(body) == 0 {
		t.Fatalf("expected JSON event, got %v", err)
	}
}
