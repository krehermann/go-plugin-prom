package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	api "github.com/krehermann/go-plugin-prom/api/v1/controller"
	"github.com/krehermann/go-plugin-prom/common"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery/targetgroup"

	"google.golang.org/grpc"
)

type OperatorConfig struct {
	Port          int
	EnablePlugins bool
	PromPort      int
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

func pluginMetricPath(name string) string {
	return fmt.Sprintf("plugins/%s/metrics", name)
}

func extractPluginName(urlPath string) string {
	temp := strings.TrimLeft(urlPath, "/")
	return strings.Split(temp, "/")[1]
}

func (o *Operator) staticConfigHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	groups := make([]*targetgroup.Group, 0)
	o.controller.m.Lock()
	defer o.controller.m.Unlock()
	for _, p := range o.controller.pluginMap {
		// create a metric target for each running plugin
		target := &targetgroup.Group{
			Targets: []model.LabelSet{
				{model.AddressLabel: model.LabelValue(fmt.Sprintf("localhost:%d", o.cfg.PromPort))},
				{model.AddressLabel: model.LabelValue(fmt.Sprintf("host.docker.internal:%d", o.cfg.PromPort))},
			},
			Labels: map[model.LabelName]model.LabelValue{
				"job":                  model.LabelValue(fmt.Sprintf("plugin_%s-wrapper", p.name)),
				model.MetricsPathLabel: model.LabelValue(pluginMetricPath(p.name)),
			},
		}

		groups = append(groups, target, p.promTarget)
	}

	b, err := json.Marshal(groups)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(b)
}

func (o *Operator) pluginMetricHandler(w http.ResponseWriter, req *http.Request) {
	// route to metric handler for the target
	log.Printf("plugin metric handler url path %s", req.URL.Path)
	pluginName := extractPluginName(req.URL.Path)

	o.controller.m.Lock()
	p, ok := o.controller.pluginMap[pluginName]
	o.controller.m.Unlock()

	if !ok {
		w.Write([]byte(fmt.Sprintf("plugin '%s' does not exist", pluginName)))
		return
	}

	pluginURL := fmt.Sprintf("http://localhost:%d/metrics", p.port)
	res, err := http.Get(pluginURL)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(b)
}

func (o *Operator) Run(ctx context.Context) error {

	if !o.cfg.EnablePlugins {
		log.Printf("server running common counter")
		common.RunAbstractedCounter(1*time.Second, 3)
	}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", o.cfg.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	o.Listener = lis
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.HandleFunc("/plugins/", o.pluginMetricHandler)

		http.HandleFunc("/sd_config", o.staticConfigHandler)
		log.Printf("serving prom endpoints at %v", o.cfg.PromPort)
		err = http.ListenAndServe(fmt.Sprintf(":%d", o.cfg.PromPort), nil)
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
	log.Println("awaiting signal")
	<-done
	log.Println("exiting")

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
		log.Println()
		log.Println(sig)
		done <- struct{}{}
	}()

}
