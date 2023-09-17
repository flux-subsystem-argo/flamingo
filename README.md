# Flamingo - the Flux Subsystem for Argo

> 🚀🚀🚀 **BREAKING:** Introducing the new [Flamingo CLI](https://github.com/flux-subsystem-argo/cli)! 🚀🚀🚀

Flamingo is the **Flux Subsystem for Argo** (FSA). Flamingo's container image can be used as a drop-in extension for the equivalent ArgoCD version to visualize, and manage Flux workloads, alongside ArgoCD. You can also ensure that upstream CVEs in Argo CD are quickly backported to Flamingo, maintaining a secure and stable environment.

![support_by_weaveworks](https://github.com/flux-subsystem-argo/flamingo/assets/10666/41b3a990-d94a-4247-a015-b7486a76034f)

This opensource project is sponsored by [Weaveworks](https://www.weave.works/blog/weave-gitops-assured-accelerator-oss-with-enterprise-support), and they provide commercial support for its users via **Weave GitOps Assured Accelerator**.

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

Install CLIs
- [KIND CLI](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) 
- [Flux CLI](https://fluxcd.io/docs/cmd/)
- [Flamingo CLI](https://github.com/flux-subsystem-argo/cli)

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
curl -sL https://bit.ly/flamingo-cli | sudo bash
```

Create a fresh KIND cluster

```shell
kind create cluster
```

Install **Flux**

```shell
flux install
```

Install **Flamingo**

```shell
flamingo install --version=v2.8.3
```

Create a **Flux Kustomization**

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

Generate a Flamingo application to visualize the `podinfo` objects.

```shell
flamingo generate-app \
  --app-name=podinfo-ks \
  -n podinfo-kustomize ks/podinfo
```

Like a normal Argo CD instance, please firstly obtain the initial password by running the following command to login.
The default user name is `admin`.

```shell
flamingo show-init-password
```

After that you can port forward and open your browser to http://localhost:8080
```
kubectl -n argocd port-forward svc/argocd-server 8080:443
```
