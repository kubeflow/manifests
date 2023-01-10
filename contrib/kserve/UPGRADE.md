# Upgrade Documentation

## Upgrade Kserve

### Prerequisites

- Running Kubernetes cluster with kubeflow installed.
- `kubectl` configured to talk to the desired cluster.
- `kustomize`
- `curl`

> **_NOTE:_** This documentation assumes that you are running the commands in linux.
        If you are using another OS, please make sure to update the Makefile commands. 

### To update the kserve manifest to specific version follow the below instructions.

1. Set the desired version to upgrade.

   ```sh
   export KSERVE_VERSION=0.10.0rc0
   ```

2. Rebuild the manifests.

   ```sh
   make upgrade-kserve-manifests
   ```

3. Install the updated manifests.
   ```sh
   make install-kserve
   ```