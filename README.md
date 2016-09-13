[![CircleCI](https://circleci.com/gh/utilitywarehouse/github-sshkey-provider.svg?style=svg)](https://circleci.com/gh/utilitywarehouse/github-sshkey-provider)

# github-sshkey-provider
Updates OpenSSH `authorized_keys` files based on GitHub Team membership.

## Architecture
There are two components to this system service:
- The collector: a Deployment and a Service which will collect the SSH keys from GitHub
- The agent: a Daemonset which reads the SSH keys from the collector and applies it to the system's `authorized_keys` file

These two components use a `redis` server to communicate.

## Configuration
The configuration is set through environment variables and a kubernetes Secret. The manifests for these can be found in [utilitywarehouse/kubernetes-manifests](https://github.com/utilitywarehouse/kubernetes-manifests) in the system namespace.
