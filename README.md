# Flamingo - the Flux Subsystem for Argo

**This project is looking for sponsorship for its continuation. Please feel free to contact chanwit at gmail, if you want to support this project.**

> 🚀🚀🚀 **BREAKING:** Introducing the new [Flamingo CLI](https://github.com/flux-subsystem-argo/flamingo)! 🚀🚀🚀

Flamingo is the **Flux Subsystem for Argo** (FSA). Flamingo's container image can be used as a drop-in extension for the equivalent ArgoCD version to visualize, and manage Flux workloads, alongside ArgoCD. You can also ensure that upstream CVEs in Argo CD are quickly backported to Flamingo, maintaining a secure and stable environment.

## Why use Flamingo?

Flamingo is a tool that combines Flux and Argo CD to provide the best of both worlds for implementing GitOps on Kubernetes clusters. With Flamingo, you can:

1. Automate the deployment of your applications to Kubernetes clusters and benefit from the improved collaboration and deployment speed and reliability that GitOps offers.

2. Enjoy a seamless and integrated experience for managing deployments, with the automation capabilities of Flux embedded inside the user-friendly interface of Argo CD.

3. Take advantage of additional features and capabilities that are not available in either Flux or Argo CD individually, such the robust Helm support from Flux, Flux OCI Repository, Weave GitOps Terraform Controller for Infrastructure as Code, Weave Policy Engine, or Argo CD ApplicationSet for Flux-managed resources.

Try Flamingo today and see how it can improve your GitOps workflow on Kubernetes.

This provides a brief overview of the benefits of using Flamingo and why it could be a useful tool for implementing GitOps on Kubernetes clusters. Of course, you may want to tailor this to your specific use case and requirements, but this should give you a good starting point.

## How does it work?

**Loopback Reconciliation** is a feature of Flamingo that helps to synchronize applications deployed using the GitOps approach. It is activated when the "FluxSubsystem" feature is enabled in the ArgoCD user interface (UI).

Here's how Loopback Reconciliation works:

1. An ArgoCD application manifest is created and deployed to a cluster, either in Kustomization or Helm mode.

2. Flamingo converts the ArgoCD application manifest into the equivalent Flux object, either a Kustomization object or a HelmRelease object with a Source, depending on the mode used in the ArgoCD manifest. If Flux objects already exist for the application, Flamingo will use them as references instead of creating new ones.

3. Flamingo synchronizes or reconciles the state of the ArgoCD application with its Flux counterparts by using the state of the Flux objects as the desired state. To do this, the Loopback Reconciliation mechanism bypasses the native reconciliation process in ArgoCD and relies on Flux reconciliation instead. It then uses the result from the Flux objects to report back to ArgoCD.

Loopback Reconciliation helps to ensure the reliability and consistency of GitOps-based deployments by keeping the state of applications in sync with their desired state defined in the Flux objects. The technique gets its name because it involves "looping back" to the desired state defined in the Flux objects as references to reconcile the state of the application.

![FSA (2)](https://user-images.githubusercontent.com/10666/159503288-5faeda59-8b54-40f0-95ca-b46c22742e30.png)

## Getting Started with Flamingo CLI

This guide will provide a step-by-step process for setting up a GitOps environment using Flux and ArgoCD, via Flamingo. By the end of this guide, you will have Flamingo running locally on your KIND cluster. You will create a `podinfo` application with a Flux Kustomization, and generate a Flamingo app from this Flux object.

### Install CLIs

- [KIND CLI](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) 
- [Flux CLI](https://fluxcd.io/docs/cmd/)
- [Flamingo CLI](https://github.com/flux-subsystem-argo/flamingo)

Example install in macOS or Linux via [homebrew](https://brew.sh/)

```shell
# install KIND cli
brew install kind

# install Flux CLI
brew install fluxcd/tap/flux

# install Flamingo CLI 
# with Homebrew
brew install flux-subsystem-argo/tap/flamingo

# or with cURL
curl -sL https://raw.githubusercontent.com/flux-subsystem-argo/flamingo/main/install/flamingo.sh | sudo bash
```

### Create a fresh KIND cluster

```shell
kind create cluster
```

### Install **Flux**

```shell
flux install
```

### Install **Flamingo**

#### Install **Flamingo** via CLI

```shell
flamingo install
```

```shell
# or with a specific version
flamingo install --version=v2.8.4
```

#### Install **Flamingo** with HelmRelease

```yaml
cat << EOF | kubectl apply -f -
---
apiVersion: v1
kind: Namespace
metadata:
  name: argocd
---
apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  name: argo-helm-repo
  namespace: argocd
spec:
  interval: 10m
  url: https://argoproj.github.io/argo-helm
---
apiVersion: helm.toolkit.fluxcd.io/v2beta2
kind: HelmRelease
metadata:
  name: flamingo
  namespace: argocd
spec:
  interval: 10m
  targetNamespace: argocd
  releaseName: argocd
  chart:
    spec:
      chart: argo-cd
      version: '*'
      sourceRef:
        kind: HelmRepository
        name: argo-helm-repo
  values:
    global:
      image:
        repository: ghcr.io/flux-subsystem-argo/fsa/argocd
        tag: v2.10.2-fl.23-main-d2c9a8cb # replace with the latest version
EOF
```

If you want to see Flamingo itself as an application, run:

```shell
flamingo gen-app --app-name=flamingo -n argocd hr/flamingo
```

### Example workloads

#### Create a **Flux Kustomization**

```yaml
cat << EOF | kubectl apply -f -
---
apiVersion: v1
kind: Namespace
metadata:
  name: podinfo-kustomize
---
apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: OCIRepository
metadata:
  name: podinfo
  namespace: podinfo-kustomize
spec:
  interval: 10m
  url: oci://ghcr.io/stefanprodan/manifests/podinfo
  ref:
    tag: latest
---
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: podinfo
  namespace: podinfo-kustomize
spec:
  interval: 10m
  targetNamespace: podinfo-kustomize
  prune: true
  sourceRef:
    kind: OCIRepository
    name: podinfo
  path: ./
EOF
```

To generate a Flamingo application to visualize the `podinfo` objects, run:

```shell
flamingo generate-app \
  --app-name=podinfo-ks \
  -n podinfo-kustomize ks/podinfo
```

#### Create a **Flux HelmRelease**

```shell
cat << EOF | kubectl apply -f -
---
apiVersion: v1
kind: Namespace
metadata:
  name: podinfo-helm
---
apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  name: podinfo
  namespace: podinfo-helm
spec:
  interval: 10m
  type: oci
  url: oci://ghcr.io/stefanprodan/charts
---
apiVersion: helm.toolkit.fluxcd.io/v2beta2
kind: HelmRelease
metadata:
  name: podinfo
  namespace: podinfo-helm
spec:
  interval: 10m
  targetNamespace: podinfo-helm
  chart:
    spec:
      chart: podinfo
      version: '*'
      sourceRef:
        kind: HelmRepository
        name: podinfo
      verify:
        provider: cosign # keyless
EOF
```

```shell
flamingo generate-app \
  --app-name=podinfo-hr \
  -n podinfo-helm hr/podinfo
```

Like a normal Argo CD instance, please firstly obtain the initial password by running the following command to login.
The default username is `admin`.

```shell
flamingo show-init-password
```

After that you can port forward and open your browser to http://localhost:8080

```shell
kubectl -n argocd port-forward svc/argocd-server 8080:443
```

## Multi-cluster support

### What is Flamingo multi-cluster support?

Flamingo multi-cluster is a feature designed to visualize multiple Flux clusters in the Flamingo UI.
We use ArgoCD cluster secrets to store cluster information for Flamingo multi-cluster support.
To list clusters, use the following command:

```shell
flamingo list-clusters
```

To register a Kubernetes cluster with Flamingo, simply use the `add-cluster` command.
For example, the following command adds the `dev-1` cluster definition from the `KUBECONFIG` file, then overrides the server name and server address, and sets it to skip TLS verification.

```shell
flamingo add-cluster dev-1 \
  --server-name=dev-1.vcluster-dev-1 \
  --server-addr=https://dev-1.vcluster-dev-1.svc \
  --insecure
```

This `add-cluster` command currently supports adding a cluster with static credentials, such as client certs and client keys (like Vclusters), but does not yet support authentication via KubeClient plugins. For clusters like EKS and others, you need to create cluster secrets manually.

To generate applications from Flux workloads on leaf clusters, the flamingo `generate-app` command has been extended to support the resource format as `cluster/kind/object-name`, for example:

```shell
flamingo generate-app \
  dev-1/ks/podinfo
```

The above command will generate an application named `podinfo` from the `ks` resource on the `dev-1` cluster.
This `generate-app` command uses Flamingo cluster information to connect to the leaf cluster and generate the application.
Currently, this command only supports static cluster credentials with client certs and keys.
For dynamic cluster credentials like in EKS and others, you would use the `--context` flag to select the Kube's context for the leaf cluster and generate the application with the `--server` flag to override the destination cluster, like:

```shell
flamingo generate-app \
  --context=dev-1 \
  --server=https://dev-1.vcluster-dev-1.svc \
  ks/podinfo --export | kubectl apply -f -
```
