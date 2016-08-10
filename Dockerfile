FROM alpine

RUN apk add --no-cache ca-certificates

ADD github-sshkey-provider /

ARG DOCKER_IMAGE_NAME
ENV DOCKER_IMAGE_NAME ${DOCKER_IMAGE_NAME:-unknown}

ENTRYPOINT ["/github-sshkey-provider"]
