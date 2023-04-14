// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package server

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/go-plugin/examples/basic/shared"
	common "github.com/krehermann/go-plugin-prom/common"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery/targetgroup"
)

type pluginWrapper struct {
	*plugin.Client
	promTarget *targetgroup.Group
	port       int
	name       string
}

func startPlugin(name string) (*pluginWrapper, error) {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   fmt.Sprintf("plugin-%s", name),
		Output: os.Stdout,
		Level:  hclog.Debug,
	})

	// find a port for prom server in the plugin
	port, err := common.GetPortInRange(2115, 2120)
	if err != nil {
		return nil, err
	}

	target := &targetgroup.Group{
		Targets: []model.LabelSet{
			{model.AddressLabel: model.LabelValue(fmt.Sprintf("localhost:%d", port))},
			//	{model.AddressLabel: model.LabelValue(fmt.Sprintf("host.docker.internal:%d", port))},
		},
		Labels: map[model.LabelName]model.LabelValue{
			"job": model.LabelValue(fmt.Sprintf("plugin_%s", name)),
			//model.MetricsPathLabel: model.LabelValue("custom_metric_path"),
		},
		Source: "",
	}

	// We're a host! Start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
		Cmd:             exec.Command("./bin/greeter", "-port", strconv.Itoa(port)),
		Logger:          logger,
	})
	//defer client.Kill()

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("failed to get rpc conn to plugin: %w", err)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("greeter")
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("failed to dispense: %w", err)
	}

	// We should have a Greeter now! This feels like a normal interface
	// implementation but is in fact over an RPC connection.
	greeter := raw.(common.Greeter)
	fmt.Println(greeter.Greet())
	return &pluginWrapper{
		Client:     client,
		promTarget: target,
		port:       port,
		name:       name,
	}, nil
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

// pluginMap is the map of plugins we can dispense.
var pluginMap = map[string]plugin.Plugin{
	"greeter": &shared.GreeterPlugin{},
}
