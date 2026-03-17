# ARM64 / AArch64 support: install, validation, and troubleshooting

This document helps you validate Kubeflow on **ARM64 / AArch64** (e.g., AWS Graviton, OCI Ampere, Apple Silicon dev machines using Kind/Minikube) and collect actionable debugging data for issue tracking.

> Tracking issue: kubeflow/manifests#2745 (Support for the aarch64 arm64 architecture)

## Recommended versions
ARM64 support depends on individual component images being available for `linux/arm64`. If you see missing image errors on older releases, try:
- the `v1.9-branch`, or
- the default branch

## Prerequisites
- A Kubernetes cluster running on ARM64 nodes
- `kubectl`
- `kustomize`

Check node architecture:

```bash
kubectl get nodes -o wide
kubectl get nodes -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.nodeInfo.architecture}{"\n"}{end}'
```

Expected architecture is typically `arm64`.

## Install (example)
Follow the repository installation instructions. Most users start from the `example/` kustomization.

## Common ARM64 failure modes (what they mean)

### 1) Image pull failures (missing ARM64 image)
Symptoms:
- Pods stuck in `ImagePullBackOff` or `ErrImagePull`

Collect evidence:

```bash
kubectl get pods -A | grep -E "ImagePullBackOff|ErrImagePull" || true
```

Describe a failing pod to capture the exact image and error:

```bash
kubectl describe pod -n <namespace> <pod-name> | sed -n '/Containers:/,$p'
kubectl describe pod -n <namespace> <pod-name> | sed -n '/Events:/,$p'
```

What to report:
- Image reference (registry/repo:tag or @sha256 digest)
- Full error message from Events
- Kubeflow manifests version/branch used

### 2) Wrong-architecture binary (`exec format error`)
Symptoms:
- Pod starts but crashes immediately
- Logs may contain `exec format error`

Collect evidence:

```bash
kubectl get pods -A | grep -E "CrashLoopBackOff|Error" || true
kubectl logs -n <namespace> <pod-name> --previous=true
```

What to report:
- Image reference
- Pod logs showing `exec format error`
- Node architecture and Kubernetes version

### 3) CrashLoopBackOff (runtime crash)
Not all crashes are architecture-related, but ARM can expose differences in dependencies or build flags.

Collect:
```bash
kubectl describe pod -n <namespace> <pod-name> | sed -n '/Events:/,$p'
kubectl logs -n <namespace> <pod-name> --previous=true
```

## Extract the full list of images referenced by manifests
This is useful to identify which images must support `linux/arm64`.

Example for `example/`:

```bash
kustomize build example > /tmp/kubeflow.yaml
grep -E "^[[:space:]]*image:[[:space:]]*" /tmp/kubeflow.yaml \
  | sed -E 's/^[[:space:]]*image:[[:space:]]*//' \
  | sort -u > /tmp/kubeflow-images.txt

wc -l /tmp/kubeflow-images.txt
head -n 50 /tmp/kubeflow-images.txt
```

Attach `/tmp/kubeflow-images.txt` (or a subset of relevant failing images) when reporting issues.

## Suggested report template for kubeflow/manifests#2745
When commenting on #2745, include:

- **Manifests ref**: (e.g., `v1.9-branch` commit SHA or release tag)
- **Cluster**: (provider, Kubernetes version)
- **Node arch**: output of `kubectl get nodes ... architecture`
- **Failing pods**: `kubectl get pods -A | grep -E "ImagePullBackOff|ErrImagePull|CrashLoopBackOff|Error"`
- **Failing images**: list of image references (especially ones in `ImagePullBackOff`)
- **Pod describe/events**: relevant `kubectl describe pod ...` Events section

This makes it much easier to identify which component images are missing `linux/arm64`.
