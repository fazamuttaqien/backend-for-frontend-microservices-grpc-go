package postgres

import (
	"context"
	"errors"

	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/domain"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct{ db *pgxpool.Pool }

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository { return &OrderRepository{db: db} }

var _ repository.OrderRepository = (*OrderRepository)(nil)

func (r *OrderRepository) Create(ctx context.Context, o *domain.Order) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `INSERT INTO orders (id,user_id,total,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6)`, o.ID, o.UserID, o.Total, o.Status, o.CreatedAt, o.UpdatedAt)
	if err != nil {
		return err
	}
	for _, item := range o.Items {
		if _, err = tx.Exec(ctx, `INSERT INTO order_items (id,order_id,product_id,quantity,price,total) VALUES ($1,$2,$3,$4,$5,$6)`, item.ID, o.ID, item.ProductID, item.Quantity, item.Price, item.Total); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	o := &domain.Order{}
	err := r.db.QueryRow(ctx, `SELECT id,user_id,total,status,created_at,updated_at FROM orders WHERE id=$1`, id).Scan(&o.ID, &o.UserID, &o.Total, &o.Status, &o.CreatedAt, &o.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, `SELECT id,product_id,quantity,price,total FROM order_items WHERE order_id=$1 ORDER BY id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		item := domain.OrderItem{}
		if err := rows.Scan(&item.ID, &item.ProductID, &item.Quantity, &item.Price, &item.Total); err != nil {
			return nil, err
		}
		o.Items = append(o.Items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return o, nil
}

func (r *OrderRepository) List(ctx context.Context, page, pageSize int) ([]*domain.Order, int, error) {
	offset := (page - 1) * pageSize
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM orders`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(ctx, `SELECT id,user_id,total,status,created_at,updated_at FROM orders ORDER BY created_at DESC LIMIT $1 OFFSET $2`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	orders := make([]*domain.Order, 0)
	for rows.Next() {
		o := &domain.Order{}
		if err := rows.Scan(&o.ID, &o.UserID, &o.Total, &o.Status, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, 0, err
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	for _, o := range orders {
		items, err := r.items(ctx, o.ID)
		if err != nil {
			return nil, 0, err
		}
		o.Items = items
	}
	return orders, total, nil
}

func (r *OrderRepository) items(ctx context.Context, orderID string) ([]domain.OrderItem, error) {
	rows, err := r.db.Query(ctx, `SELECT id,product_id,quantity,price,total FROM order_items WHERE order_id=$1 ORDER BY id`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]domain.OrderItem, 0)
	for rows.Next() {
		item := domain.OrderItem{}
		if err := rows.Scan(&item.ID, &item.ProductID, &item.Quantity, &item.Price, &item.Total); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
