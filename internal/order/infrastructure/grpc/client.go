package grpcclient

import (
	"context"
	"time"

	productv1 "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/gen/product/v1"
	userv1 "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/gen/user/v1"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/observability"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func dial(ctx context.Context, address string, timeout time.Duration) (*grpc.ClientConn, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithUnaryInterceptor(observability.UnaryClientInterceptor()),
	)
}

func DialUser(ctx context.Context, address string, timeout time.Duration) (*grpc.ClientConn, error) {
	return dial(ctx, address, timeout)
}

type UserClient struct {
	client  userv1.UserServiceClient
	timeout time.Duration
}

func NewUserClient(conn grpc.ClientConnInterface, timeout time.Duration) *UserClient {
	return &UserClient{client: userv1.NewUserServiceClient(conn), timeout: timeout}
}

func (c *UserClient) GetUser(ctx context.Context, id string) error {
	callCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	_, err := c.client.GetUser(callCtx, &userv1.GetUserRequest{Id: id})
	return err
}

func DialProduct(ctx context.Context, address string, timeout time.Duration) (*grpc.ClientConn, error) {
	return dial(ctx, address, timeout)
}

type ProductClient struct {
	client  productv1.ProductServiceClient
	timeout time.Duration
}

func NewProductClient(conn grpc.ClientConnInterface, timeout time.Duration) *ProductClient {
	return &ProductClient{client: productv1.NewProductServiceClient(conn), timeout: timeout}
}

func (c *ProductClient) GetProduct(ctx context.Context, id string) (string, error) {
	callCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	res, err := c.client.GetProduct(callCtx, &productv1.GetProductRequest{Id: id})
	if err != nil {
		return "", err
	}
	if res.GetProduct() == nil {
		return "", status.Error(codes.Internal, "product response is empty")
	}
	return res.GetProduct().GetPrice(), nil
}
