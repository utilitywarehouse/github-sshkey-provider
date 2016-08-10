#!/bin/sh

set -o errexit
set -o nounset

if [ $# -ne 2 ]; then
    echo "usage: ./scripts/deploy.sh <repo_name> <git_sha>"
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

repo_name=$1
git_sha=$2

payload () {
    cat <<-PAYLOAD
    {
        "spec": {
            "template": {
                "spec": {
                    "containers": [
                        {
                            "name": "${repo_name}-${1}",
                            "image": "docker.io/utilitywarehouse/$repo_name:$git_sha"
                        }
                    ]
                }
            }
        }
    }
PAYLOAD
}

curl -k -XPATCH \
    -d "$(payload collector)" \
    -H "Content-Type: application/strategic-merge-patch+json" \
    -H "Authorization: Bearer ${KUBERNETES_TOKEN}" \
    "${KUBERNETES_URL}/apis/extensions/v1beta1/namespaces/default/deployments/${repo_name}-collector"
