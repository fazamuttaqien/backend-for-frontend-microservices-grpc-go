package domain

import (
	"errors"
	"math/big"
	"strings"
	"time"
)

var (
	ErrOrderNotFound       = errors.New("order not found")
	ErrInvalidOrder        = errors.New("invalid order")
	ErrUserNotFound        = errors.New("user not found")
	ErrProductNotFound     = errors.New("product not found")
	ErrDependencyUnavailable = errors.New("dependency unavailable")
)

type OrderItem struct {
	ID        string
	ProductID string
	Quantity  int32
	Price     string
	Total     string
}

type Order struct {
	ID        string
	UserID    string
	Items     []OrderItem
	Total     string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewOrder(id, userID string, items []OrderItem, now time.Time) (*Order, error) {
	if id == "" || strings.TrimSpace(userID) == "" || len(items) == 0 {
		return nil, ErrInvalidOrder
	}

	total := new(big.Rat)
	for _, item := range items {
		if item.ID == "" || item.ProductID == "" || item.Quantity <= 0 || strings.TrimSpace(item.Price) == "" || strings.TrimSpace(item.Total) == "" {
			return nil, ErrInvalidOrder
		}
		itemTotal, ok := new(big.Rat).SetString(item.Total)
		if !ok || itemTotal.Sign() < 0 {
			return nil, ErrInvalidOrder
		}
		total.Add(total, itemTotal)
	}

	return &Order{ID: id, UserID: strings.TrimSpace(userID), Items: items, Total: total.FloatString(4), Status: "PENDING", CreatedAt: now, UpdatedAt: now}, nil
}
