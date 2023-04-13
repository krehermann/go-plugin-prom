#!/bin/sh


docker run \
    -p 9090:9090 \
    -v $PWD/prometheus.yml:/etc/prometheus/prometheus.yml \
    --add-host=host.docker.internal:host-gateway \
    prom/prometheus