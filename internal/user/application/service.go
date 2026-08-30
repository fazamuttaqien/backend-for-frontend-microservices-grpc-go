package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/user/domain"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/user/repository"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type UserService struct {
	repo repository.UserRepository
	now  func() time.Time
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo, now: time.Now}
}
func newID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return hex.EncodeToString(b)
}
func (s *UserService) Create(ctx context.Context, name, email, password string) (*domain.User, error) {
	if password == "" {
		return nil, domain.ErrInvalidUser
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u, err := domain.NewUser(newID(), name, email, string(hash), s.now())
	if err != nil {
		return nil, err
	}
	if err = s.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}
func (s *UserService) Get(ctx context.Context, id string) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}
func (s *UserService) Update(ctx context.Context, id, name, email, password string) (*domain.User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	hash := ""
	if password != "" {
		b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		hash = string(b)
	}
	if err = u.Update(name, email, hash, s.now()); err != nil {
		return nil, err
	}
	if err = s.repo.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}
