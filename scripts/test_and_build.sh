#!/bin/sh

set -o errexit
set -o nounset

apk add --no-cache curl git

curl -sS https://glide.sh/get | sh

glide i

go test -cover $(glide novendor)

CGO_ENABLED=0 go build -v -a -ldflags '-s -extldflags "-static"'
