#!/bin/sh

set -o errexit
set -o nounset

apk add --no-cache curl git redis

curl -sS https://glide.sh/get | sh

glide i

redis-server --port 6379 --daemonize yes
redis-server --port 6380 --requirepass password --daemonize yes

go test -tags integration -cover $(glide novendor)

CGO_ENABLED=0 go build -v -a -ldflags '-s -extldflags "-static"'
