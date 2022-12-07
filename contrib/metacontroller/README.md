# Metacontroller

- [Official documentation](https://metacontroller.github.io/metacontroller/)
- [Official repository](https://github.com/metacontroller/metacontroller)

Metacontroller is an add-on for Kubernetes that makes it easy to write and deploy custom controllers.

## Prerequisites

- Kubernetes v1.16+ (because of maintainability, e2e test suite might not cover all releases).
- You should have `kubectl` available and configured to talk to the desired cluster.
- `kustomize`.

## Compile manifests

```bash
make hydrate
```

## Install Metacontroller

```bash
make apply
```

## Verify deployment

```bash
make test
```

## Uninstall Metacontroller

```bash
make delete
```

## Upgrade Metacontroller

To upgrade to the lates version used in Kubeflow, follow the steps in [UPGRADE.md](./UPDGRADE.md).
