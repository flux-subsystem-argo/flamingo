# Getting Started

```bash
kind create cluster
```

```bash
flux install
```

```bash
flamingo install
```

```bash
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
  url: oci://ghcr.io/stefanprodan/podinfo-deploy
  ref:
    semver: "*"
  verify:
    provider: cosign
    secretRef:
      name: cosign-pub
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

```shell
flamingo generate-app --app-name=podinfo-ks -n podinfo-kustomize ks/podinfo
```

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
apiVersion: helm.toolkit.fluxcd.io/v2beta1
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
flamingo generate-app --app-name=podinfo-hr -n podinfo-helm hr/podinfo
```

```shell
wget https://raw.githubusercontent.com/stefanprodan/podinfo/master/.cosign/cosign.pub

kubectl -n podinfo-kustomize create secret generic cosign-pub \
  --from-file=cosign.pub=cosign.pub

kubectl -n podinfo-helm create secret generic cosign-pub \
  --from-file=cosign.pub=cosign.pub
```

```shell
flamingo show-init-password
```
