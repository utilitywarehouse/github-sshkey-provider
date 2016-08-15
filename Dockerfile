FROM alpine

RUN apk add --no-cache ca-certificates

ADD . /github-sshkey-provider

ARG UW_IMAGE_NAME
ENV UW_IMAGE_NAME ${UW_IMAGE_NAME:-}
ARG UW_GIT_SHA
ENV UW_GIT_SHA ${UW_GIT_SHA:-}

ENTRYPOINT ["/github-sshkey-provider/github-sshkey-provider"]
