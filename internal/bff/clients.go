package bff

import (
	"context"
	"time"

	orderv1 "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/gen/order/v1"
	productv1 "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/gen/product/v1"
	userv1 "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/gen/user/v1"
	"google.golang.org/grpc"
)

type UserClient struct { client userv1.UserServiceClient; timeout time.Duration }
func NewUserClient(conn grpc.ClientConnInterface)*UserClient{return &UserClient{client:userv1.NewUserServiceClient(conn),timeout:3*time.Second}}
func(c *UserClient)Get(ctx context.Context,id,token string)(*userv1.User,error){ctx,cancel:=context.WithTimeout(ctx,c.timeout);defer cancel();res,err:=c.client.GetUser(withBearer(ctx,token),&userv1.GetUserRequest{Id:id});if err!=nil{return nil,err};return res.GetUser(),nil}

type ProductClient struct{client productv1.ProductServiceClient;timeout time.Duration}
func NewProductClient(conn grpc.ClientConnInterface)*ProductClient{return &ProductClient{client:productv1.NewProductServiceClient(conn),timeout:3*time.Second}}
func(c *ProductClient)List(ctx context.Context,page,size int32)(*productv1.ListProductResponse,error){ctx,cancel:=context.WithTimeout(ctx,c.timeout);defer cancel();return c.client.ListProduct(ctx,&productv1.ListProductRequest{Page:page,PageSize:size})}
func(c *ProductClient)Get(ctx context.Context,id string)(*productv1.Product,error){ctx,cancel:=context.WithTimeout(ctx,c.timeout);defer cancel();res,err:=c.client.GetProduct(ctx,&productv1.GetProductRequest{Id:id});if err!=nil{return nil,err};return res.GetProduct(),nil}

type OrderClient struct{client orderv1.OrderServiceClient;timeout time.Duration}
func NewOrderClient(conn grpc.ClientConnInterface)*OrderClient{return &OrderClient{client:orderv1.NewOrderServiceClient(conn),timeout:5*time.Second}}
func(c *OrderClient)Create(ctx context.Context,req *orderv1.CreateOrderRequest)(*orderv1.Order,error){ctx,cancel:=context.WithTimeout(ctx,c.timeout);defer cancel();res,err:=c.client.CreateOrder(ctx,req);if err!=nil{return nil,err};return res.GetOrder(),nil}
func(c *OrderClient)List(ctx context.Context,page,size int32)(*orderv1.ListOrdersResponse,error){ctx,cancel:=context.WithTimeout(ctx,c.timeout);defer cancel();return c.client.ListOrders(ctx,&orderv1.ListOrdersRequest{Page:page,PageSize:size})}
func(c *OrderClient)Get(ctx context.Context,id string)(*orderv1.Order,error){ctx,cancel:=context.WithTimeout(ctx,c.timeout);defer cancel();res,err:=c.client.GetOrder(ctx,&orderv1.GetOrderRequest{Id:id});if err!=nil{return nil,err};return res.GetOrder(),nil}

type Clients struct{User *UserClient;Product *ProductClient;Order *OrderClient}
