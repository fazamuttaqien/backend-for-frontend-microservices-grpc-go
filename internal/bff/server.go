package bff

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	orderv1 "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/gen/order/v1"
	productv1 "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/gen/product/v1"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Server struct { jwt *auth.JWT; clients Clients }
func NewServer(jwt *auth.JWT, clients Clients) *Server { return &Server{jwt:jwt, clients:clients} }
type tokenKey struct{}
func contextWithToken(ctx context.Context, token string) context.Context { return context.WithValue(ctx, tokenKey{}, token) }
func tokenFromContext(ctx context.Context) string { v,_:=ctx.Value(tokenKey{}).(string); return v }

type ErrorResponse struct { Error string `json:"error"` }
type UserResponse struct { ID string `json:"id"`; Name string `json:"name"`; Email string `json:"email"`; CreatedAt string `json:"created_at"`; UpdatedAt string `json:"updated_at"` }
type ProductResponse struct { ID string `json:"id"`; Name string `json:"name"`; Description string `json:"description"`; Price string `json:"price"`; Stock int32 `json:"stock"`; CreatedAt string `json:"created_at"`; UpdatedAt string `json:"updated_at"` }
type ProductListResponse struct { Products []ProductResponse `json:"products"`; Total int32 `json:"total"`; Page int32 `json:"page"`; PageSize int32 `json:"page_size"` }
type CreateOrderRequest struct { Items []CreateOrderItem `json:"items"` }
type CreateOrderItem struct { ProductID string `json:"product_id"`; Quantity int32 `json:"quantity"` }
type OrderItemResponse struct { ProductID string `json:"product_id"`; Quantity int32 `json:"quantity"`; Price string `json:"price"`; Total string `json:"total"` }
type OrderResponse struct { ID string `json:"id"`; UserID string `json:"user_id"`; Items []OrderItemResponse `json:"items"`; Total string `json:"total"`; Status string `json:"status"`; CreatedAt string `json:"created_at"`; UpdatedAt string `json:"updated_at"` }
type OrderListResponse struct { Orders []OrderResponse `json:"orders"`; Total int32 `json:"total"`; Page int32 `json:"page"`; PageSize int32 `json:"page_size"` }

type UserClient struct { client interface{ GetUser(context.Context,*userv1.GetUserRequest,...grpc.CallOption)(*userv1.GetUserResponse,error) }; timeout time.Duration }
