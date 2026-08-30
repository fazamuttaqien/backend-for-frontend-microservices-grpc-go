package postgres

import (
	"context"
	"errors"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/product/domain"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/product/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepository struct{ db *pgxpool.Pool }

func NewProductRepository(db *pgxpool.Pool) *ProductRepository { return &ProductRepository{db: db} }
var _ repository.ProductRepository = (*ProductRepository)(nil)

func (r *ProductRepository) Create(ctx context.Context, p *domain.Product) error {
	_, err := r.db.Exec(ctx, `INSERT INTO products (id,name,description,price,stock,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`, p.ID, p.Name, p.Description, p.Price, p.Stock, p.CreatedAt, p.UpdatedAt)
	return err
}

func (r *ProductRepository) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	p := &domain.Product{}
	err := r.db.QueryRow(ctx, `SELECT id,name,description,price,stock,created_at,updated_at FROM products WHERE id=$1`, id).Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) { return nil, domain.ErrProductNotFound }
	if err != nil { return nil, err }
	return p, nil
}

func (r *ProductRepository) List(ctx context.Context, page, pageSize int) ([]*domain.Product, int, error) {
	offset := (page - 1) * pageSize
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM products`).Scan(&total); err != nil { return nil, 0, err }
	rows, err := r.db.Query(ctx, `SELECT id,name,description,price,stock,created_at,updated_at FROM products ORDER BY created_at DESC LIMIT $1 OFFSET $2`, pageSize, offset)
	if err != nil { return nil, 0, err }
	defer rows.Close()
	products := make([]*domain.Product, 0)
	for rows.Next() {
		p := &domain.Product{}
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.CreatedAt, &p.UpdatedAt); err != nil { return nil, 0, err }
		products = append(products, p)
	}
	if err := rows.Err(); err != nil { return nil, 0, err }
	return products, total, nil
}

func (r *ProductRepository) Update(ctx context.Context, p *domain.Product) error {
	tag, err := r.db.Exec(ctx, `UPDATE products SET name=$1,description=$2,price=$3,stock=$4,updated_at=$5 WHERE id=$6`, p.Name, p.Description, p.Price, p.Stock, p.UpdatedAt, p.ID)
	if err != nil { return err }
	if tag.RowsAffected() == 0 { return domain.ErrProductNotFound }
	return nil
}
