package main

// There are two components currently in the Flamingo project:
// - the ArgoCD controllers
// - the Flamingo controller

const allInstallTemplate = `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: {{ .Namespace }}
resources:
- namespace.yaml
- "https://raw.githubusercontent.com/argoproj/argo-cd/{{ .ArgoCD }}/manifests/install.yaml"
images:
- name: quay.io/argoproj/argocd:{{ .ArgoCD }}
  newName: ghcr.io/flux-subsystem-argo/fsa/argocd
  newTag: {{ .Image }}
{{ .AnonymousPatches }}
`

const crdsOnlyInstallTemplate = `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- "https://raw.githubusercontent.com/argoproj/argo-cd/{{ .ArgoCD }}/manifests/crds/application-crd.yaml"
- "https://raw.githubusercontent.com/argoproj/argo-cd/{{ .ArgoCD }}/manifests/crds/applicationset-crd.yaml"
- "https://raw.githubusercontent.com/argoproj/argo-cd/{{ .ArgoCD }}/manifests/crds/appproject-crd.yaml"
`

const namespaceInstallTemplate = `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: {{ .Namespace }}
resources:
- namespace.yaml
- "https://raw.githubusercontent.com/argoproj/argo-cd/{{ .ArgoCD }}/manifests/namespace-install.yaml"
- cluster.yaml
images:
- name: quay.io/argoproj/argocd:{{ .ArgoCD }}
  newName: ghcr.io/flux-subsystem-argo/fsa/argocd
  newTag: {{ .Image }}
{{ .AnonymousPatches }}
`

const anonymousPatches = `
patches:
- patch: |-
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: argocd-cm
      labels:
        app.kubernetes.io/name: argocd-cm
        app.kubernetes.io/part-of: argocd
    data:
      users.anonymous.enabled: "true"
  target:
    kind: ConfigMap
    name: argocd-cm
- patch: |-
    apiVersion: v1
    kind: ConfigMap
    metadata:
      labels:
        app.kubernetes.io/name: argocd-rbac-cm
        app.kubernetes.io/part-of: argocd
      name: argocd-rbac-cm
    data:
      policy.default: role:readonly
      policy.csv: |
        p, role:readonly, applications, get, default/*, allow
        p, role:readonly, applications, sync, default/*, allow
        p, role:readonly, clusters, get, *, allow
        p, role:readonly, repositories, get, *, allow
        p, role:readonly, project, get, default/*, allow
  target:
    kind: ConfigMap
    name: argocd-rbac-cm
`

var namespaceTemplate = `
apiVersion: v1
kind: Namespace
metadata:
  name: %s
`

const defaultClusterSecretTemplate = `
---
apiVersion: v1
kind: Secret
metadata:
  name: cluster-kubernetes.default.svc
  namespace: %s
  annotations:
    managed-by: argocd.argoproj.io
  labels:
    argocd.argoproj.io/secret-type: cluster
type: Opaque
stringData:
  config: '{"tlsClientConfig":{"insecure":false}}'
  name: in-cluster
  namespaces: %s
  server: https://kubernetes.default.svc
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  labels:
    toolkit.fluxcd.io/tenant: %s
  name: flamingo-reconciler
  namespace: %s
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: argocd-application-controller
  namespace: %s
`

const helmReleaseInstallTemplate = `
---
apiVersion: v1
kind: Namespace
metadata:
  name: {{ .Namespace }}
---
apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  name: argo-helm-repo
  namespace: {{ .Namespace }}
spec:
  interval: 10m
  url: https://argoproj.github.io/argo-helm
---
apiVersion: helm.toolkit.fluxcd.io/v2beta2
kind: HelmRelease
metadata:
  name: flamingo
  namespace: {{ .Namespace }}
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
        tag: {{ .Image }}
`
