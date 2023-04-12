package server

import (
	"context"
	"log"

	api "github.com/krehermann/go-plugin-prom/api/v1/controller"
)

// server is used to implement helloworld.GreeterServer.
type Server struct {
	api.UnimplementedControllerServer
}

// SayHello implements helloworld.GreeterServer
func (s *Server) Start(ctx context.Context, in *api.StartRequest) (*api.StartResponse, error) {
	log.Printf("Received start: %v", in.GetName())
	return &api.StartResponse{}, nil
}

// SayHello implements helloworld.GreeterServer
func (s *Server) Stop(ctx context.Context, in *api.StopRequest) (*api.StopResponse, error) {
	log.Printf("Received stop: %v", in.GetName())
	return &api.StopResponse{}, nil
}

// SayHello implements helloworld.GreeterServer
func (s *Server) Kill(ctx context.Context, in *api.KillRequest) (*api.KillResponse, error) {
	log.Printf("Received kill: %v", in.GetName())
	return &api.KillResponse{}, nil
}
