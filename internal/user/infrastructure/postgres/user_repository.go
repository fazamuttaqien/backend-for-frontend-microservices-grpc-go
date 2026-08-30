package postgres

import (
	"context"
	"errors"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/user/domain"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/user/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct{ db *pgxpool.Pool }

func NewUserRepository(db *pgxpool.Pool) *UserRepository { return &UserRepository{db: db} }

var _ repository.UserRepository = (*UserRepository)(nil)

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	_, err := r.db.Exec(ctx, `INSERT INTO users (id,name,email,password_hash,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6)`, u.ID, u.Name, u.Email, u.PasswordHash, u.CreatedAt, u.UpdatedAt)
	var pe *pgconn.PgError
	if errors.As(err, &pe) && pe.Code == "23505" {
		return domain.ErrEmailAlreadyExists
	}
	return err
}
func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	u := &domain.User{}
	err := r.db.QueryRow(ctx, `SELECT id,name,email,password_hash,created_at,updated_at FROM users WHERE id=$1`, id).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}
func (r *UserRepository) Update(ctx context.Context, u *domain.User) error {
	tag, err := r.db.Exec(ctx, `UPDATE users SET name=$1,email=$2,password_hash=$3,updated_at=$4 WHERE id=$5`, u.Name, u.Email, u.PasswordHash, u.UpdatedAt, u.ID)
	var pe *pgconn.PgError
	if errors.As(err, &pe) && pe.Code == "23505" {
		return domain.ErrEmailAlreadyExists
	}
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}
