# github-sshkey-provider - deprecated

DEPRECATED in favour of https://github.com/utilitywarehouse/ssh-key-agent

Updates OpenSSH `authorized_keys` files based on GitHub Team membership.

**This repository is not actively developed or supported anymore**

## Architecture
There are two components to this system service:
- The collector: a Deployment and a Service which will collect the SSH keys from GitHub
- The agent: a Daemonset which reads the SSH keys from the collector and applies it to the system's `authorized_keys` file

These two components communicate over HTTP.

Docker images are available here: https://quay.io/repository/utilitywarehouse/github-sshkey-provider?tab=tags.

## Configuration
The configuration is set through environment variables and a kubernetes Secret. The manifests for these can be found in [utilitywarehouse/kubernetes-manifests](https://github.com/utilitywarehouse/kubernetes-manifests) in the system namespace.
