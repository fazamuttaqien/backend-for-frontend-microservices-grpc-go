package application_test

import (
	"context"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/user/application"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/user/domain"
	"testing"
)

type repo struct{ users map[string]*domain.User }

func (r *repo) Create(_ context.Context, u *domain.User) error {
	if r.users == nil {
		r.users = map[string]*domain.User{}
	}
	r.users[u.ID] = u
	return nil
}
func (r *repo) GetByID(_ context.Context, id string) (*domain.User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}
func (r *repo) Update(_ context.Context, u *domain.User) error { r.users[u.ID] = u; return nil }
func TestCreateHashesPassword(t *testing.T) {
	r := &repo{}
	s := application.NewUserService(r)
	u, err := s.Create(context.Background(), "Faza", "FAZA@EXAMPLE.COM", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if u.PasswordHash == "secret" || u.PasswordHash == "" {
		t.Fatal("password was not hashed")
	}
	if u.Email != "faza@example.com" {
		t.Fatalf("email=%q", u.Email)
	}
}
func TestGetAndUpdate(t *testing.T) {
	r := &repo{}
	s := application.NewUserService(r)
	u, err := s.Create(context.Background(), "Old", "old@example.com", "secret")
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.Get(context.Background(), u.ID)
	if err != nil || got.ID != u.ID {
		t.Fatal(err)
	}
	updated, err := s.Update(context.Background(), u.ID, "New", "new@example.com", "")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "New" || updated.Email != "new@example.com" {
		t.Fatal("update failed")
	}
	if updated.PasswordHash != u.PasswordHash {
		t.Fatal("password hash changed unexpectedly")
	}
}
