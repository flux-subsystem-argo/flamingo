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
|v0.38 | v2.5    | v2.5.7-fl.3-main-b0b6148f
|v0.38 | v2.4    | v2.4.19-fl.3-main-b0b6148f
|v0.38 | v2.3    | v2.3.13-fl.3-main-b0b6148f
|v0.37 | v2.2    | v2.2.16-fl.3-main-2bba0ae6

## How does it work?

**Loopback Reconciliation** is a feature of Flamingo that helps to synchronize applications deployed using the GitOps approach. It is activated when the "FluxSubsystem" feature is enabled in the ArgoCD user interface (UI).

Here's how Loopback Reconciliation works:

1. An ArgoCD application manifest is created and deployed to a cluster, either in Kustomization or Helm mode.

2. Flamingo converts the ArgoCD application manifest into the equivalent Flux object, either a Kustomization object or a HelmRelease object with a Source, depending on the mode used in the ArgoCD manifest. If Flux objects already exist for the application, Flamingo will use them as references instead of creating new ones.

3. Flamingo synchronizes or reconciles the state of the ArgoCD application with its Flux counterparts by using the state of the Flux objects as the desired state. To do this, the Loopback Reconciliation mechanism bypasses the native reconciliation process in ArgoCD and relies on Flux reconciliation instead. It then uses the result from the Flux objects to report back to ArgoCD.

Loopback Reconciliation helps to ensure the reliability and consistency of GitOps-based deployments by keeping the state of applications in sync with their desired state defined in the Flux objects. The technique gets its name because it involves "looping back" to the desired state defined in the Flux objects as references to reconcile the state of the application.

![FSA (2)](https://user-images.githubusercontent.com/10666/159503288-5faeda59-8b54-40f0-95ca-b46c22742e30.png)

## Getting Started with a Fresh KIND cluster

This guide will provide a step-by-step process for setting up a GitOps environment using Flux and ArgoCD, via Flamingo. We will use this public repository to install and bootstrap Flamingo, so no manual installation steps are required. However, if you fork the repository and make it private, you will need to set up a Secret to authenticate your Git repository.

By the end of this guide, you will have Flamingo running locally on your KIND cluster. You will run Flamingo in anonymous mode and see two pre-defined ArgoCD applications, each of which points to its equivalent Flux Kustomization.

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

![image1](https://user-images.githubusercontent.com/10666/208858892-5e5d14d9-61c7-4c61-af29-1883e7137509.png)

![image2](https://user-images.githubusercontent.com/10666/208858840-fca56550-a2a1-4fff-829e-f1469e921c86.png)

![image3](https://user-images.githubusercontent.com/10666/208858784-9a508a5b-8d47-47d8-b5a5-0f9adaff72cf.png)

Like a normal Argo CD instance, please firstly obtain the initial password by running the following command to login and create other Flux applications.
The default user name is `admin`.

```
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d; echo
```
