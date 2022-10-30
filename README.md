# FSA - Flux Subsystem for Argo

FSA (aka Flamingo) is **Flux Subsystem for Argo**. FSA's container image can be used as a drop-in replacement for the equivalent ArgoCD version to visualize, and manage Flux workloads, alongside ArgoCD.

![fsa_logo_alpha_200x200](https://user-images.githubusercontent.com/10666/190959808-725101fd-6e11-4057-af55-f0ab37f3983c.png)

## Support Matrix

|Flux  | Argo CD | Image
|:----:|:-------:|---------------------------
|v0.36 | v2.5    | v2.5.0-fl.3-main-2bba0ae6
|v0.36 | v2.4    | v2.4.15-fl.3-main-2bba0ae6
|v0.36 | v2.3    | v2.3.10-fl.3-main-2bba0ae6
|v0.36 | v2.2    | v2.2.15-fl.3-main-2bba0ae6

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
