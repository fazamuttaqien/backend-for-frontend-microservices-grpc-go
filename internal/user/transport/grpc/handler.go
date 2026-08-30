package grpc

import (
	"context"
	"errors"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/gen/user/v1"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/user/application"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/user/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	userv1.UnimplementedUserServiceServer
	app *application.UserService
}

func NewHandler(app *application.UserService) *Handler { return &Handler{app: app} }
func response(u *domain.User) *userv1.User {
	return &userv1.User{Id: u.ID, Name: u.Name, Email: u.Email, CreatedAt: u.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"), UpdatedAt: u.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z")}
}
func mapErr(err error) error {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrEmailAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, domain.ErrInvalidUser):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
func (h *Handler) CreateUser(ctx context.Context, in *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	u, err := h.app.Create(ctx, in.Name, in.Email, in.Password)
	if err != nil {
		return nil, mapErr(err)
	}
	return &userv1.CreateUserResponse{User: response(u)}, nil
}
func (h *Handler) GetUser(ctx context.Context, in *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	u, err := h.app.Get(ctx, in.Id)
	if err != nil {
		return nil, mapErr(err)
	}
	return &userv1.GetUserResponse{User: response(u)}, nil
}
func (h *Handler) UpdateUser(ctx context.Context, in *userv1.UpdateUserRequest) (*userv1.UpdateUserResponse, error) {
	u, err := h.app.Update(ctx, in.Id, in.Name, in.Email, in.Password)
	if err != nil {
		return nil, mapErr(err)
	}
	return &userv1.UpdateUserResponse{User: response(u)}, nil
}
