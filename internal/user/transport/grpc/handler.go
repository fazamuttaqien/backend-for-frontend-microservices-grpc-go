package grpc

import (
	"context"
	"errors"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/gen/user/v1"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/auth"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/user/application"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/user/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct { userv1.UnimplementedUserServiceServer; app *application.UserService; jwt *auth.JWT }
func NewHandler(app *application.UserService,jwt *auth.JWT)*Handler{return &Handler{app:app,jwt:jwt}}
const timeFormat="2006-01-02T15:04:05Z"
func response(u *domain.User)*userv1.User{return &userv1.User{Id:u.ID,Name:u.Name,Email:u.Email,CreatedAt:u.CreatedAt.UTC().Format(timeFormat),UpdatedAt:u.UpdatedAt.UTC().Format(timeFormat)}}
func mapErr(err error)error{switch{case errors.Is(err,domain.ErrUserNotFound):return status.Error(codes.NotFound,err.Error());case errors.Is(err,domain.ErrEmailAlreadyExists):return status.Error(codes.AlreadyExists,err.Error());case errors.Is(err,domain.ErrInvalidUser):return status.Error(codes.InvalidArgument,err.Error());case errors.Is(err,domain.ErrInvalidCredentials):return status.Error(codes.Unauthenticated,"invalid credentials");default:return status.Error(codes.Internal,"internal server error")}}
func(h *Handler)CreateUser(ctx context.Context,in *userv1.CreateUserRequest)(*userv1.CreateUserResponse,error){u,e:=h.app.Create(ctx,in.Name,in.Email,in.Password);if e!=nil{return nil,mapErr(e)};return &userv1.CreateUserResponse{User:response(u)},nil}
func(h *Handler)Register(ctx context.Context,in *userv1.CreateUserRequest)(*userv1.CreateUserResponse,error){return h.CreateUser(ctx,in)}
func(h *Handler)Login(ctx context.Context,in *userv1.LoginRequest)(*userv1.LoginResponse,error){u,e:=h.app.Authenticate(ctx,in.Email,in.Password);if e!=nil{return nil,mapErr(e)};token,e:=h.jwt.Issue(u.ID,u.Email);if e!=nil{return nil,status.Error(codes.Internal,"internal server error")};claims,e:=h.jwt.Parse(token);if e!=nil{return nil,status.Error(codes.Internal,"internal server error")};return &userv1.LoginResponse{AccessToken:token,ExpiresAt:claims.Expires,User:response(u)},nil}
func(h *Handler)GetUser(ctx context.Context,in *userv1.GetUserRequest)(*userv1.GetUserResponse,error){u,e:=h.app.Get(ctx,in.Id);if e!=nil{return nil,mapErr(e)};return &userv1.GetUserResponse{User:response(u)},nil}
func(h *Handler)UpdateUser(ctx context.Context,in *userv1.UpdateUserRequest)(*userv1.UpdateUserResponse,error){u,e:=h.app.Update(ctx,in.Id,in.Name,in.Email,in.Password);if e!=nil{return nil,mapErr(e)};return &userv1.UpdateUserResponse{User:response(u)},nil}
