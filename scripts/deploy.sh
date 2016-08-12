#!/bin/sh

set -o errexit
set -o nounset

if [ $# -ne 3 ]; then
    echo "usage: ./scripts/deploy.sh <deployment> <container> <image>"
    exit 1
fi

if [ -z ${KUBERNETES_URL:-} ]; then
    echo "could not read the kubernetes url from \$KUBERNETES_URL"
    exit 1
fi

if [ -z ${KUBERNETES_TOKEN:-} ]; then
    echo "could not read the kubernetes token from \$KUBERNETES_TOKEN"
    exit 1
fi

deployment=$1
container=$2
image=$3

payload () {
    cat <<-PAYLOAD
    {
        "spec": {
            "template": {
                "spec": {
                    "containers": [
                        {
                            "name": "${container}",
                            "image": "${image}"
                        }
                    ]
                }
            }
        }
    }
PAYLOAD
}

curl -k -XPATCH \
    -d "$(payload)" \
    -H "Content-Type: application/strategic-merge-patch+json" \
    -H "Authorization: Bearer ${KUBERNETES_TOKEN}" \
    "${KUBERNETES_URL}/apis/extensions/v1beta1/namespaces/default/deployments/${deployment}"
