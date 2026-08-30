package main

import (
	"context"
	"os"
	"time"

	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if len(os.Args) != 2 { os.Exit(2) }
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	conn, err := grpc.NewClient(os.Args[1], grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil { os.Exit(1) }
	defer conn.Close()
	res, err := healthpb.NewHealthClient(conn).Check(ctx, &healthpb.HealthCheckRequest{})
	if err != nil || res.GetStatus() != healthpb.HealthCheckResponse_SERVING { os.Exit(1) }
}
