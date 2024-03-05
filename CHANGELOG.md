# Changelog

Changes of the notable versions of the Flamingo CLI are documented in this file. 

## v0.10.2 (2024-03-05)

**New Features and Bug Fixes**

  * Add new FSA image **v2.10.2** as the default version.
  * Promoted `v2.10.2-fl.23-main-d2c9a8cb` to the default version.
  * Implement `flamingo/{kustomize,helmrelease,gitrepository,ocirepository,helmrepository}-override` to override the default configuration in the auto-create mode.
  * Implement `flamingo/{kustomize,helmrelease,gitrepository,ocirepository,helmrepository}-replace` to replace the default configuration in the auto-create mode.
  * Multi-cluster support for remote Flux clusters. The multi-cluster support works with any kinds of cluster configurations, but there are some limitation in the Flamingo CLI commands, see below.
  * Implement `flamingo add-cluster` to generate a cluster secret for the remote Flux cluster. This command currently supports only the static cluster configuration. 
  * Implement `flamingo generate-app cluster/kind/resource` to generate an Application CR for resources in the remote Flux cluster. This command currently supports only the static cluster configuration.
  * Add `flamingo install --mode=helmrelease` to install Flamingo with HelmRelease. This command also support `--export` flag to export the HelmRelease CR.
  * Add documentation for the multi-cluster support.
  * Add documentation for the `flamingo install --mode=helmrelease` command.

## v0.8.3 (2024-02-16)

**New Features and Bug Fixes**

  * Add new FSA image **v2.9.6** that supports HelmRelease v2beta2.
  * Promoted `v2.9.6-fl.22-main-402c9e49` to the default version.

## v0.8.2 (2024-02-10)

**New Features and Bug Fixes**

  * Promote FSA image **v2.9.6** to the default version.

## v0.8.1 (2024-02-10)

**New Features and Bug Fixes**

  * Fix health check for OCI HelmRepository in Flux v2.2.x.

## v0.8.0 (2024-02-10)

**New Features and Bug Fixes**

  * Add FSA image **v2.9.6** to the index
  * Fix empty path in Kustomization (generate-app ks)

## v0.7.0 (2023-12-30)

**New Features and Bug Fixes**

  * Promote FSA image **v2.9.3** to the default version

## v0.6.1 (2023-11-08)

**New Features and Bug Fixes**

  * SSA wait timeout is changed from 1 to 5 minutes

## v0.6.0 (2023-11-07)

**New Features and Bug Fixes**

  * Promote FSA image **v2.8.6** to the default version
  * Add FSA image **v2.9.0** to the dev channel
  * Show the version during the installation process
  * Timeout default value is now 10 minutes

## v0.5.1 (2023-11-02)

**New Features and Bug Fixes**
  * Promote FSA image **v2.8.5** to the default version
  * Add FSA image **v2.8.6** to the dev channel
  * Add HelmRelease quickstart to the docs
  * Fix installation script
  * Fix `install` command help to show the latest FSA image
  * Prefix `v` to the FSA image version for the `install` command

## v0.5.0 (2023-10-30)

**New Features and Bug Fixes**
  * Promote FSA image **v2.8.4** to the default version
  * Add FSA image **v2.8.5** and **v2.9.0-rc3** to the dev channel
  * Add Flux v2.1.2 support to the dev channel 
  * Add `lc` alias for `list-candidates` command
  * Remove `--dev` flag out of the `install` command
  * CLI repository migration from `flux-subsystem-argo/cli` to `flux-subsystem-argo/flamingo`
