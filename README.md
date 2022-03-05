# Flamingo

Flamingo is **the Flux Subsystem for Argo**. Flamingo's container image can be used as a drop-in replacement for the equivalent ArgoCD version to visualize, and manage Flux workloads, along side ArgoCD.

## Getting Started with a Fresh KIND cluster

In this getting started guide, you'll be walked through steps to prepare your ultimate GitOps environment using ArgoCD and Flux.
We'll bootstrap everything from this public repo, so there is no secret required. At the end of this guide, you'll have Flux running alongside ArgoCD locally on your KIND cluster.

Install CLIs
```shell
# install KIND cli
brew install kind

# install Flux CLI
brew install fluxcd/tap/flux

# install ArgoCD CLI
brew install argocd
```

Create a fresh KIND cluster
```shell
kind create cluster
```

Install Flux
```shell
flux install
```

Copy, paste and run this snippet to bootstrap the demo.
```shell
cat <<EOF | kubectl apply -f -
---
apiVersion: source.toolkit.fluxcd.io/v1beta1
kind: GitRepository
metadata:
  name: fsa-demo
  namespace: flux-system
spec:
  interval: 30s
  url: https://github.com/chanwit/flamingo
  ref:
    branch: main
---
apiVersion: kustomize.toolkit.fluxcd.io/v1beta2
kind: Kustomization
metadata:
  name: fsa-demo
  namespace: flux-system
spec:
  prune: true
  interval: 2m
  path: "./demo"
  sourceRef:
    kind: GitRepository
    name: fsa-demo
  timeout: 3m
EOF
```
