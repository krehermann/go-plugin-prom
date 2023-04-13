package server

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/hashicorp/go-plugin"
	api "github.com/krehermann/go-plugin-prom/api/v1/controller"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	start_count = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "start_count",
		Help: "The total number of starts events",
	}, []string{"plugin_name"})

	stop_count = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "start_count",
		Help: "The total number of starts events",
	}, []string{"plugin_name"})
)

// server is used to implement helloworld.GreeterServer.
type controllerGRPCimpl struct {
	api.UnimplementedControllerServer

	m         sync.Mutex
	pluginMap map[string]*plugin.Client
}

func NewServer() *controllerGRPCimpl {
	return &controllerGRPCimpl{
		pluginMap: make(map[string]*plugin.Client),
	}
}

// SayHello implements helloworld.GreeterServer
func (s *controllerGRPCimpl) Start(ctx context.Context, in *api.StartRequest) (*api.StartResponse, error) {
	log.Printf("Received start: %v", in.GetName())
	s.m.Lock()
	if _, exists := s.pluginMap[in.Name]; exists {
		return &api.StartResponse{}, fmt.Errorf("%s already running", in.Name)
	}
	s.m.Unlock()

	p, err := startPlugin(in.Name)
	if err != nil {
		return &api.StartResponse{}, err
	}
	start_count.WithLabelValues(in.Name).Inc()
	s.m.Lock()
	s.pluginMap[in.Name] = p
	s.m.Unlock()

	return &api.StartResponse{}, nil
}

// SayHello implements helloworld.GreeterServer
func (s *controllerGRPCimpl) Stop(ctx context.Context, in *api.StopRequest) (*api.StopResponse, error) {
	log.Printf("Received stop: %v", in.GetName())

	s.m.Lock()
	p, exists := s.pluginMap[in.Name]
	if !exists {
		return &api.StopResponse{}, fmt.Errorf("%s not running", in.Name)
	} else {
		p.Kill()
		delete(s.pluginMap, in.Name)
	}
	s.m.Unlock()

	stop_count.WithLabelValues(in.Name).Inc()

	return &api.StopResponse{}, nil
}

// SayHello implements helloworld.GreeterServer
func (s *controllerGRPCimpl) Kill(ctx context.Context, in *api.KillRequest) (*api.KillResponse, error) {
	log.Printf("Received kill: %v", in.GetName())
	// todo send self explode directive to plugin
	return &api.KillResponse{}, nil
}
