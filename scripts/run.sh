#!/bin/sh

set -o errexit
set -o nounset

if [  $# -ne 1 ]; then
    echo "usage: ./scripts/run.sh <script>"
    echo "   eg: ./scripts/run.sh test_and_build"
    exit 1
fi

if [ ! -x scripts/$1.sh ]; then
   echo "invalid script: ${1}"
   exit 1
fi

build_image_name=golang:1.6.3-alpine
build_image_gopath=/go/src
package_name=github.com/utilitywarehouse/github-sshkey-provider

docker run -it --rm \
    -v $(pwd):${build_image_gopath}/${package_name} \
    -w ${build_image_gopath}/${package_name} \
    ${build_image_name} ./scripts/${1}.sh
