# Flamingo - the Flux Subsystem for Argo

Flamingo is the **Flux Subsystem for Argo** (FSA). Flamingo's container image can be used as a drop-in replacement for the equivalent ArgoCD version to visualize, and manage Flux workloads, alongside ArgoCD.

## Why use Flamingo?

Flamingo is a tool that combines Flux and Argo CD to provide the best of both worlds for implementing GitOps on Kubernetes clusters. With Flamingo, you can:

1. Automate the deployment of your applications to Kubernetes clusters and benefit from the improved collaboration and deployment speed and reliability that GitOps offers.

2. Enjoy a seamless and integrated experience for managing deployments, with the automation capabilities of Flux embedded inside the user-friendly interface of Argo CD.

3. Take advantage of additional features and capabilities that are not available in either Flux or Argo CD individually, such the robust Helm support from Flux, Flux OCI Repository, Weave GitOps Terraform Controller for Infrastructure as Code, Weave Policy Engine, or Argo CD ApplicationSet for Flux-managed resources.

Try Flamingo today and see how it can improve your GitOps workflow on Kubernetes.

This provides a brief overview of the benefits of using Flamingo and why it could be a useful tool for implementing GitOps on Kubernetes clusters. Of course, you may want to tailor this to your specific use case and requirements, but this should give you a good starting point.

## Support Matrix

|Flux  | Argo CD | Image
|:----:|:-------:|---------------------------
|v0.37 | v2.5    | v2.5.5-fl.3-main-46fefeb3
|v0.37 | v2.4    | v2.4.18-fl.3-main-46fefeb3
|v0.37 | v2.3    | v2.3.12-fl.3-main-46fefeb3
|v0.37 | v2.2    | v2.2.16-fl.3-main-2bba0ae6

## How does it work?

![FSA (2)](https://user-images.githubusercontent.com/10666/159503288-5faeda59-8b54-40f0-95ca-b46c22742e30.png)

## Getting Started with a Fresh KIND cluster

In this getting started guide, you'll be walked through steps to prepare your ultimate GitOps environment using ArgoCD and Flux.
We'll bootstrap everything, including installation of ArgoCD, from this public repo. So no manual step of ArgoCD installation is required.
In case you're forking this repo and change its visibility to private, you will be required to setup a Secret to authenticate your Git repo.

At the end of this guide, you'll have Flux running alongside ArgoCD locally on your KIND cluster. You'll run FSA in the anonymous mode, and see 2 pre-defined ArgoCD Applications, each of which points to its equivalent Flux Kustomization.

Install CLIs
- [KIND cli](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) 
- [Flux cli](https://fluxcd.io/docs/cmd/)
- [ArgoCD cli](https://argo-cd.readthedocs.io/en/stable/cli_installation/)

Example install in macOS or Linux via [homebrew](https://brew.sh/)

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

Install **Flux**

```shell
flux install

```

You can check the Flux namespace (`flux-system`) for running pods `kubectl get pods -n flux-system`

![image](./images/kubectl-get-ns-flux-system.png)


Copy, and paste this snippet to bootstrap the demo.

```shell
cat <<EOF | kubectl apply -f -
---
apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: OCIRepository
metadata:
  name: fsa-demo
  namespace: flux-system
spec:
  interval: 30s
  url: oci://ghcr.io/flux-subsystem-argo/flamingo/manifests
  ref:
    tag: v2.5
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
    kind: OCIRepository
    name: fsa-demo
  timeout: 3m
EOF

```

Check ArgoCD pods are running and Ready `kubectl get -n argocd pods`

![image](./images/argocd-pods-ready.png)


After that you can port forward and open your browser to http://localhost:8080

```
kubectl -n argocd port-forward svc/argocd-server 8080:443
```

You'll find 2 FSA Applications, each of which consists of 1 Flux's Kustomization and 1 Flux's GitRepository.

![image](https://user-images.githubusercontent.com/10666/161395963-bbabbd72-03f5-4cef-b16d-346afd0eb1fc.png)

![image](https://user-images.githubusercontent.com/10666/161396000-0282f538-88a9-4449-8501-6d7b3a64f2a6.png)

Like a normal Argo CD instance, please firstly obtain the initial password by running the following command to login and create other Flux applications.
The default user name is `admin`.

```
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d; echo
```
