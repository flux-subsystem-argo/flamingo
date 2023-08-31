package main

// There are two components currently in the Flamingo project:
// - the ArgoCD controllers
// - the Flamingo controller

const defaultTemplate = `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: {{ .Namespace }}
resources:
- "https://raw.githubusercontent.com/argoproj/argo-cd/{{ .ArgoCD }}/manifests/install.yaml"
- namespace.yaml
images:
- name: quay.io/argoproj/argocd:{{ .ArgoCD }}
  newName: ghcr.io/flux-subsystem-argo/fsa/argocd
  newTag: {{ .Fsa }}
`

const readonlyTemplate = `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: {{ .Namespace }}
resources:
- "https://raw.githubusercontent.com/argoproj/argo-cd/{{ .ArgoCD }}/manifests/install.yaml"
- namespace.yaml
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
images:
- name: quay.io/argoproj/argocd:{{ .ArgoCD }}
  newName: ghcr.io/flux-subsystem-argo/fsa/argocd
  newTag: {{ .Fsa }}
`
