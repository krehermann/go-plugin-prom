package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

	net.Listener
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
	o.Listener = lis
	go func() {
		http.Handle("/metrics", promhttp.Handler())

		err = http.ListenAndServe(":2112", nil)
		if err != nil {
			log.Fatalf("error starting prom metric endpoint: %v", err)
		}
	}()

	done := make(chan struct{}, 1)
	go o.signalHandler(done)

	go func() {
		s := grpc.NewServer()
		api.RegisterControllerServer(s, o.controller)
		log.Printf("server listening at %v", lis.Addr())

		err := s.Serve(lis)
		if err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	// The program will wait here until it gets the
	// expected signal (as indicated by the goroutine
	// above sending a value on `done`) and then exit.
	fmt.Println("awaiting signal")
	<-done
	fmt.Println("exiting")

	return nil
}

func (o *Operator) signalHandler(done chan struct{}) {
	// Go signal notification works by sending `os.Signal`
	// values on a channel. We'll create a channel to
	// receive these notifications. Note that this channel
	// should be buffered.
	sigs := make(chan os.Signal, 1)

	// `signal.Notify` registers the given channel to
	// receive notifications of the specified signals.
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// We could receive from `sigs` here in the main
	// function, but let's see how this could also be
	// done in a separate goroutine, to demonstrate
	// a more realistic scenario of graceful shutdown.

	go func() {
		// This goroutine executes a blocking receive for
		// signals. When it gets one it'll print it out
		// and then notify the program that it can finish.
		sig := <-sigs
		o.controller.Shutdown()
		o.Listener.Close()
		fmt.Println()
		fmt.Println(sig)
		done <- struct{}{}
	}()

}
