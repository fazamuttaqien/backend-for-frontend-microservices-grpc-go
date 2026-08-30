package productv1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const (
	ProductService_CreateProduct_FullMethodName = "/product.v1.ProductService/CreateProduct"
	ProductService_GetProduct_FullMethodName = "/product.v1.ProductService/GetProduct"
	ProductService_ListProduct_FullMethodName = "/product.v1.ProductService/ListProduct"
	ProductService_UpdateProduct_FullMethodName = "/product.v1.ProductService/UpdateProduct"
)

type ProductServiceClient interface { CreateProduct(context.Context,*CreateProductRequest,...grpc.CallOption)(*CreateProductResponse,error); GetProduct(context.Context,*GetProductRequest,...grpc.CallOption)(*GetProductResponse,error); ListProduct(context.Context,*ListProductRequest,...grpc.CallOption)(*ListProductResponse,error); UpdateProduct(context.Context,*UpdateProductRequest,...grpc.CallOption)(*UpdateProductResponse,error) }
type productServiceClient struct{ cc grpc.ClientConnInterface }
func NewProductServiceClient(cc grpc.ClientConnInterface) ProductServiceClient{return &productServiceClient{cc}}
func(c *productServiceClient)CreateProduct(ctx context.Context,in *CreateProductRequest,opts ...grpc.CallOption)(*CreateProductResponse,error){out:=new(CreateProductResponse);err:=c.cc.Invoke(ctx,ProductService_CreateProduct_FullMethodName,in,out,opts...);return out,err}
func(c *productServiceClient)GetProduct(ctx context.Context,in *GetProductRequest,opts ...grpc.CallOption)(*GetProductResponse,error){out:=new(GetProductResponse);err:=c.cc.Invoke(ctx,ProductService_GetProduct_FullMethodName,in,out,opts...);return out,err}
func(c *productServiceClient)ListProduct(ctx context.Context,in *ListProductRequest,opts ...grpc.CallOption)(*ListProductResponse,error){out:=new(ListProductResponse);err:=c.cc.Invoke(ctx,ProductService_ListProduct_FullMethodName,in,out,opts...);return out,err}
func(c *productServiceClient)UpdateProduct(ctx context.Context,in *UpdateProductRequest,opts ...grpc.CallOption)(*UpdateProductResponse,error){out:=new(UpdateProductResponse);err:=c.cc.Invoke(ctx,ProductService_UpdateProduct_FullMethodName,in,out,opts...);return out,err}

type ProductServiceServer interface { CreateProduct(context.Context,*CreateProductRequest)(*CreateProductResponse,error); GetProduct(context.Context,*GetProductRequest)(*GetProductResponse,error); ListProduct(context.Context,*ListProductRequest)(*ListProductResponse,error); UpdateProduct(context.Context,*UpdateProductRequest)(*UpdateProductResponse,error); mustEmbedUnimplementedProductServiceServer() }
type UnimplementedProductServiceServer struct{}
func(UnimplementedProductServiceServer)CreateProduct(context.Context,*CreateProductRequest)(*CreateProductResponse,error){return nil,status.Error(codes.Unimplemented,"method CreateProduct not implemented")}
func(UnimplementedProductServiceServer)GetProduct(context.Context,*GetProductRequest)(*GetProductResponse,error){return nil,status.Error(codes.Unimplemented,"method GetProduct not implemented")}
func(UnimplementedProductServiceServer)ListProduct(context.Context,*ListProductRequest)(*ListProductResponse,error){return nil,status.Error(codes.Unimplemented,"method ListProduct not implemented")}
func(UnimplementedProductServiceServer)UpdateProduct(context.Context,*UpdateProductRequest)(*UpdateProductResponse,error){return nil,status.Error(codes.Unimplemented,"method UpdateProduct not implemented")}
func(UnimplementedProductServiceServer)mustEmbedUnimplementedProductServiceServer(){}
func RegisterProductServiceServer(s grpc.ServiceRegistrar,srv ProductServiceServer){s.RegisterService(&ProductService_ServiceDesc,srv)}
func _ProductService_CreateProduct_Handler(srv interface{},ctx context.Context,dec func(interface{})error,interceptor grpc.UnaryServerInterceptor)(interface{},error){in:=new(CreateProductRequest);if err:=dec(in);err!=nil{return nil,err};if interceptor==nil{return srv.(ProductServiceServer).CreateProduct(ctx,in)};info:=&grpc.UnaryServerInfo{Server:srv,FullMethod:ProductService_CreateProduct_FullMethodName};handler:=func(ctx context.Context,req interface{})(interface{},error){return srv.(ProductServiceServer).CreateProduct(ctx,req.(*CreateProductRequest))};return interceptor(ctx,in,info,handler)}
func _ProductService_GetProduct_Handler(srv interface{},ctx context.Context,dec func(interface{})error,interceptor grpc.UnaryServerInterceptor)(interface{},error){in:=new(GetProductRequest);if err:=dec(in);err!=nil{return nil,err};if interceptor==nil{return srv.(ProductServiceServer).GetProduct(ctx,in)};info:=&grpc.UnaryServerInfo{Server:srv,FullMethod:ProductService_GetProduct_FullMethodName};handler:=func(ctx context.Context,req interface{})(interface{},error){return srv.(ProductServiceServer).GetProduct(ctx,req.(*GetProductRequest))};return interceptor(ctx,in,info,handler)}
func _ProductService_ListProduct_Handler(srv interface{},ctx context.Context,dec func(interface{})error,interceptor grpc.UnaryServerInterceptor)(interface{},error){in:=new(ListProductRequest);if err:=dec(in);err!=nil{return nil,err};if interceptor==nil{return srv.(ProductServiceServer).ListProduct(ctx,in)};info:=&grpc.UnaryServerInfo{Server:srv,FullMethod:ProductService_ListProduct_FullMethodName};handler:=func(ctx context.Context,req interface{})(interface{},error){return srv.(ProductServiceServer).ListProduct(ctx,req.(*ListProductRequest))};return interceptor(ctx,in,info,handler)}
func _ProductService_UpdateProduct_Handler(srv interface{},ctx context.Context,dec func(interface{})error,interceptor grpc.UnaryServerInterceptor)(interface{},error){in:=new(UpdateProductRequest);if err:=dec(in);err!=nil{return nil,err};if interceptor==nil{return srv.(ProductServiceServer).UpdateProduct(ctx,in)};info:=&grpc.UnaryServerInfo{Server:srv,FullMethod:ProductService_UpdateProduct_FullMethodName};handler:=func(ctx context.Context,req interface{})(interface{},error){return srv.(ProductServiceServer).UpdateProduct(ctx,req.(*UpdateProductRequest))};return interceptor(ctx,in,info,handler)}

var ProductService_ServiceDesc=grpc.ServiceDesc{ServiceName:"product.v1.ProductService",HandlerType:(*ProductServiceServer)(nil),Methods:[]grpc.MethodDesc{{MethodName:"CreateProduct",Handler:_ProductService_CreateProduct_Handler},{MethodName:"GetProduct",Handler:_ProductService_GetProduct_Handler},{MethodName:"ListProduct",Handler:_ProductService_ListProduct_Handler},{MethodName:"UpdateProduct",Handler:_ProductService_UpdateProduct_Handler}},Streams:[]grpc.StreamDesc{},Metadata:"product/v1/product.proto"}
