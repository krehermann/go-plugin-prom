package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	api "github.com/krehermann/go-plugin-prom/api/v1/controller"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"google.golang.org/grpc"
)

type OperatorConfig struct {
	Port int
}

type Operator struct {
	controller *controllerGRPCimpl
	cfg        OperatorConfig
}

func NewOperator(cfg OperatorConfig) *Operator {
	controller := NewServer()
	return &Operator{
		controller: controller,
		cfg:        cfg,
	}
}

func (o *Operator) Run(ctx context.Context) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", o.cfg.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	go func() {
		http.Handle("/metrics", promhttp.Handler())

		err = http.ListenAndServe(":2112", nil)
		if err != nil {
			log.Fatalf("error starting prom metric endpoint: %v", err)
		}
	}()
	s := grpc.NewServer()
	api.RegisterControllerServer(s, o.controller)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	return nil
}
