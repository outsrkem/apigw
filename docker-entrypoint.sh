#!/bin/sh
set -e

if [ -z "$1" ];then
  set -- apigw -f /usr/local/bin/apigw.yaml
fi

exec "$@"
