package grpcclient

import (
	"context"
	"time"

	productv1 "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/gen/product/v1"
	userv1 "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/gen/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func dial(ctx context.Context, address string, timeout time.Duration) (*grpc.ClientConn, error) {
	callCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return grpc.DialContext(callCtx, address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
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
