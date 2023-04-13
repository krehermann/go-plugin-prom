#!/bin/sh

docker run \
    -p 9099:9090 \
    -v $PWD/prometheus-fed.yml:/etc/prometheus/prometheus.yml \
    --add-host=host.docker.internal:host-gateway \
    prom/prometheus