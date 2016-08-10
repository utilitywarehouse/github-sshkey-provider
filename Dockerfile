FROM alpine

RUN apk add --no-cache ca-certificates

ADD github-sshkey-provider /

ENTRYPOINT ["/github-sshkey-provider"]
