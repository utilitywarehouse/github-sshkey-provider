FROM alpine

ADD github-sshkey-provider /

ENTRYPOINT ["/github-sshkey-provider"]
