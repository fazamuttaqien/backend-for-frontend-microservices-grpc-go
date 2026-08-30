package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrInvalidProduct  = errors.New("invalid product")
)

type Product struct {
	ID, Name, Description, Price string
	Stock                        int32
	CreatedAt, UpdatedAt         time.Time
}

func NewProduct(id, name, description, price string, stock int32, now time.Time) (*Product, error) {
	name = strings.TrimSpace(name)
	price = strings.TrimSpace(price)
	if id == "" || name == "" || price == "" || stock < 0 {
		return nil, ErrInvalidProduct
	}
	return &Product{ID: id, Name: name, Description: strings.TrimSpace(description), Price: price, Stock: stock, CreatedAt: now, UpdatedAt: now}, nil
}

func (p *Product) Update(name, description, price string, stock int32, now time.Time) error {
	name = strings.TrimSpace(name)
	price = strings.TrimSpace(price)
	if name == "" || price == "" || stock < 0 {
		return ErrInvalidProduct
	}
	p.Name = name
	p.Description = strings.TrimSpace(description)
	p.Price = price
	p.Stock = stock
	p.UpdatedAt = now
	return nil
}
