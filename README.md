# go-plugin-prom


# Demoware for integrating hashicorp plugins and prometheus

## Quickstart
```
sh
cd tools/docker
docker compose build
docker compose up
```

go to `localhost:9091` to see the prom instance

Navigate to targets: `http://localhost:9091/targets`

After ~ 1 minute you should see a plugin endpoint like
```
http://host.docker.internal:2112/plugins/abc/metrics	UP	
instance="host.docker.internal:2112"job="plugin_abc-wrapper"
4.599s ago	
26.599ms
```

(Ignore the `http://localhost` endpoint that are `DOWN`; this is quirk to be worked out wrt docker networking)

Once the plugin endpoint is up, run a query for plugin specific metrics:
`http://localhost:9091/graph`
search for `plugin_greet_count`, or `plugin_ticker_count`


### Under the hood
The `docker compose up` does
1. runs a grpc server
2. causes the grpc to start a plugin named `abc`
3. runs prometheus in a another container. this prom instance is configured to monitor the grpc service and the service discovery endpoint it exposes
# Details

This package implements a GRPC server that starts and stops a plugin.

The interesting bit with regard to prometheus is that we use service discovery and dynamic HTTP routing to
1. enable full monitoring of a plugin (including standard `go` routine metrcs)
2. hide the plugin behind the GRPC server.

The plugin itself is running an HTTP server for it's own prom handler, but that web server in not exposed outside the container.
This enables the fully power of prom metrics in a plugin limiting the exposed ports.

Moreover, leveraging service discovery and HTTP routing enables us to totally seperate metric monitoring from plugin application
logic.

In order for the plugin to be monitored by prom, we implement 2 HTTP endpoints in the GRPC server in addition to it's own prom `/metrics` endpoint

The endpoints serve to 
1. enable external prometheus prom to do dynamically determine what to monitor based on what plugins are running 
2. route external prom scraping requests to the plugins without exposing them directly

Specifically the endpoints are 
- `/sd_config` : HTTP Service Discovery [https://prometheus.io/docs/prometheus/latest/configuration/configuration/#http_sd_config]
    - prom is configured to poll this url to discover new endpoints to monitor. The grpc service serves the response based on what plugins are running
    - the endpoints are like `plugins/<name>/metrics` and do not directly expose a port on the plugin
- `/plugins/<name>/metrics`: Middleware to route from service discover to plugin `/metrics` endpoint
    - Once a service is discovered as above, prom calls the target endpoint at the scrape interval
    - The GRPC service acts as middleware to route the request to the `/metrics` endpoint of the requested plugin


