package health

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type Server struct { server *health.Server; service string }

func Register(s *grpc.Server, service string) *Server {
	h := health.NewServer()
	h.SetServingStatus(service, healthpb.HealthCheckResponse_SERVING)
	h.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(s, h)
	return &Server{server: h, service: service}
}

func (s *Server) Shutdown() {
	if s == nil { return }
	s.server.SetServingStatus(s.service, healthpb.HealthCheckResponse_NOT_SERVING)
	s.server.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
	s.server.Shutdown()
}
