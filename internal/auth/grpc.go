package auth

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func UnaryServerInterceptor(jwt *JWT, publicMethods map[string]bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if publicMethods[info.FullMethod] || strings.HasPrefix(info.FullMethod, "/grpc.health.v1.Health/") { return handler(ctx, req) }
		md, ok := metadata.FromIncomingContext(ctx); if !ok { return nil, status.Error(codes.Unauthenticated, "authentication required") }
		values := md.Get("authorization"); if len(values) != 1 { return nil, status.Error(codes.Unauthenticated, "authentication required") }
		parts := strings.Fields(values[0]); if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") { return nil, status.Error(codes.Unauthenticated, "invalid authentication token") }
		claims, err := jwt.Parse(parts[1]); if err != nil { return nil, status.Error(codes.Unauthenticated, "invalid authentication token") }
		return handler(WithClaims(ctx, claims), req)
	}
}
