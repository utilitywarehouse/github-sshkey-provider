#!/bin/sh

set -o errexit
set -o nounset

if [ $# -ne 1 ]; then
    echo "usage: ./scripts/deploy.sh <image>"
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

image=$1

payload () {
    cat <<-PAYLOAD
    {
        "spec": {
            "template": {
                "spec": {
                    "containers": [
                        {
                            "name": "${1}",
                            "image": "${image}"
                        }
                    ]
                }
            }
        }
    }
PAYLOAD
}

echo "> patching deployment"
curl -k -XPATCH \
    -d "$(payload collector)" \
    -H "Content-Type: application/strategic-merge-patch+json" \
    -H "Authorization: Bearer ${KUBERNETES_TOKEN}" \
    "${KUBERNETES_URL}/apis/extensions/v1beta1/namespaces/system/deployments/github-sshkey-provider-collector" >/dev/null

echo "> patching daemonset"
curl -k -XPATCH \
    -d "$(payload agent)" \
    -H "Content-Type: application/strategic-merge-patch+json" \
    -H "Authorization: Bearer ${KUBERNETES_TOKEN}" \
    "${KUBERNETES_URL}/apis/extensions/v1beta1/namespaces/system/daemonsets/github-sshkey-provider-agent" >/dev/null

current_daemonset_pods=$(curl -sk -XGET \
    -H "Content-Type: application/strategic-merge-patch+json" \
    -H "Authorization: Bearer ${KUBERNETES_TOKEN}" \
    "${KUBERNETES_URL}/api/v1/namespaces/system/pods?labelSelector=app=github-sshkey-provider,appComponent=agent" | jq -r '.items[] | .metadata.selfLink')

for p in ${current_daemonset_pods}; do
    echo "> deleting pod ${p} of daemonset"
    curl -sk -XDELETE \
        -H "Content-Type: application/strategic-merge-patch+json" \
        -H "Authorization: Bearer ${KUBERNETES_TOKEN}" \
        "${KUBERNETES_URL}${p}" >/dev/null
done
