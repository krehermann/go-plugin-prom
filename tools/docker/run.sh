#!/bin/bash

set -ex

prometheus --web.listen-address="0.0.0.0:9192" --config.file=/etc/prometheus/prometheus.yml &

bin/cli server start &
echo "done starting server"

sleep 2
bin/cli plugin --name abc start

# hack to keep alive for testing
sleep 100000000