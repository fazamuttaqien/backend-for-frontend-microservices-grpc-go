package auth

import (
	"context"
	"strings"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Interceptor struct { jwt *JWT; public map[string]struct{} }
func NewInterceptor(jwt *JWT, publicMethods ...string) *Interceptor { p:=make(map[string]struct{},len(publicMethods)); for _,m:=range publicMethods{p[m]=struct{}{}}; return &Interceptor{jwt:jwt,public:p} }
func (i *Interceptor) Unary() grpc.UnaryServerInterceptor { return func(ctx context.Context,req interface{},info *grpc.UnaryServerInfo,handler grpc.UnaryHandler)(interface{},error){if _,ok:=i.public[info.FullMethod];ok{return handler(ctx,req)};claims,err:=i.authenticate(ctx);if err!=nil{return nil,err};return handler(WithClaims(ctx,claims),req)} }
func (i *Interceptor) authenticate(ctx context.Context)(*Claims,error){md,ok:=metadata.FromIncomingContext(ctx);if !ok{return nil,status.Error(codes.Unauthenticated,"authentication required")};v:=md.Get("authorization");if len(v)!=1{return nil,status.Error(codes.Unauthenticated,"authentication required")};p:=strings.Fields(v[0]);if len(p)!=2||!strings.EqualFold(p[0],"bearer"){return nil,status.Error(codes.Unauthenticated,"invalid authorization header")};claims,err:=i.jwt.Parse(p[1]);if err!=nil{return nil,status.Error(codes.Unauthenticated,"invalid authentication token")};return claims,nil }
