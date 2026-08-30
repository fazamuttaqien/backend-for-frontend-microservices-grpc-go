package repository

import (
  "context"
  "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/user/domain"
)

type UserRepository interface {
  Create(context.Context,*domain.User) error
  GetByID(context.Context,string) (*domain.User,error)
  Update(context.Context,*domain.User) error
}
