package events

import (
	"encoding/json"
	"time"

	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/domain"
)

const (
	OrderCreatedType    = "OrderCreated"
	OrderCreatedVersion = "v1"
)

type OrderCreated struct {
	EventID     string      `json:"event_id"`
	Type        string      `json:"type"`
	Version     string      `json:"version"`
	OccurredAt  time.Time   `json:"occurred_at"`
	Data        OrderData   `json:"data"`
}

type OrderData struct {
	OrderID   string `json:"order_id"`
	UserID    string `json:"user_id"`
	Total     string `json:"total"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

func NewOrderCreated(order *domain.Order, eventID string) OrderCreated {
	return OrderCreated{
		EventID: eventID,
		Type: OrderCreatedType,
		Version: OrderCreatedVersion,
		OccurredAt: order.CreatedAt,
		Data: OrderData{
			OrderID: order.ID,
			UserID: order.UserID,
			Total: order.Total,
			Status: order.Status,
			CreatedAt: order.CreatedAt.UTC().Format(time.RFC3339Nano),
		},
	}
}

func (e OrderCreated) Marshal() ([]byte, error) { return json.Marshal(e) }
