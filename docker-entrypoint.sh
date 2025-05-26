#!/bin/sh
set -e

if [ -z "$1" ];then
    set -- /usr/local/bin/apigw -c /etc/apigw/apigw.yaml
fi

exec "$@"
