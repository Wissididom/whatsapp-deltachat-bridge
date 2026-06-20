#!/bin/sh
set -e

mkdir -p "$DC_ACCOUNTS_PATH"

deltachat-rpc-server &

sleep 2

exec bridge "$@"
