package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	common "github.com/krehermann/go-plugin-prom/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	greet_count = promauto.NewCounter(prometheus.CounterOpts{
		Name: "plugin_greet_count",
		Help: "The total number of starts events",
	})

	ticker_count = promauto.NewCounter(prometheus.CounterOpts{
		Name: "plugin_ticker_count",
		Help: "Ticker every 10s seconds",
	})
)

// Here is a real implementation of Greeter
type GreeterHello struct {
	logger hclog.Logger
}

func (g *GreeterHello) Greet() string {
	g.logger.Debug("message from GreeterHello.Greet")
	greet_count.Inc()
	return "Hello!"
}

// handshakeConfigs are used to just do a basic handshake between
// a plugin and host. If the handshake fails, a user friendly error is shown.
// This prevents users from executing bad plugins or executing a plugin
// directory. It is a UX feature, not a security feature.
var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

func main() {
	// register the abstract metric
	//prometheus.MustRegister(common.Abstracted_counter)
	common.RunAbstractedCounter(1*time.Second, 23)

	// the host process assigns a port to use. reconsider this choice in fully model
	portArg := flag.Int("port", 2200, "port for prometheus server")
	flag.Parse()

	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Trace,
		Output:     os.Stderr,
		JSONFormat: true,
	})

	greeter := &GreeterHello{
		logger: logger,
	}

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for {
			<-ticker.C
			ticker_count.Inc()
		}
	}()

	go func() {
		logger.Info("starting metric server on", "port", *portArg)
		http.Handle("/metrics", promhttp.Handler())
		addr := fmt.Sprintf(":%d", *portArg)
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			log.Fatalf("error starting prom metric endpoint: %v", err)
		}
	}()

	// pluginMap is the map of plugins we can dispense.
	var pluginMap = map[string]plugin.Plugin{
		"greeter": &common.GreeterPlugin{Impl: greeter},
	}

	logger.Debug("message from plugin", "foo", "bar")

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
