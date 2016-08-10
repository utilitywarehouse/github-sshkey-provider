#!/bin/sh

set -o errexit
set -o nounset

if [ $# -ne 3 ]; then
    echo "usage: ./scripts/deploy.sh <kubernetes_url> <repo_name> <git_sha>"
    exit 1
fi

if [[ -z ${KUBERNETES_TOKEN:-} ]]; then
    echo "could not read the kubernetes token from \$KUBERNETES_TOKEN"
    exit 1
fi

kubernetes_url=$1
repo_name=$2
git_sha=$3

function payload() {
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
    "${kubernetes_url}/apis/extensions/v1beta1/namespaces/default/deployments/${repo_name}-collector"
