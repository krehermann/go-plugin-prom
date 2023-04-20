#!/bin/bash

set -ex

#prometheus --web.listen-address="0.0.0.0:9192" --config.file=/etc/prometheus/prometheus.yml &


ENABLE_PLUGINS="${PLUGINS_ON:-false}"
PLUGIN_ARG=""
if [[ $ENABLE_PLUGINS == 'true' ]]; then
    PLUGIN_ARG="-enable-plugins"
fi

cmd="bin/cli server start  $PLUGIN_ARG"&
echo $cmd
bin/cli server start $PLUGIN_ARG &

echo "done starting server"

if [[ $ENABLE_PLUGINS == 'true' ]]; then
    sleep 2
    bin/cli plugin --name abc start
fi

# hack to keep alive for testing
sleep 100000000