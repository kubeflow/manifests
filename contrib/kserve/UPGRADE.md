# Upgrade Documentation

## Upgrade Kserve

### Prerequisites

- Running Kubernetes cluster with kubeflow installed.
- `kubectl` configured to talk to the desired cluster.
- `curl`

> **_NOTE:_** This documentation assumes that you are running the commands in linux.
        If you are using another OS, please make sure to update the Makefile commands. 

### To update the kserve manifests to specific version follow the below instructions.

1. Set the desired version to upgrade.

   ```sh
   export KSERVE_VERSION=0.10.0-rc0
   ```

2. Rebuild the manifests.

   ```sh
   make upgrade-kserve-manifests
   ```

3. Install the updated manifests.
   ```sh
   make install-kserve
   ```
> **_NOTE:_** If resource/crd installation fails please re-run the commands.

### Testing
For testing refer [kserve readme](README.md#testing-kserve).
   
## Upgrade Models Webapp
### Prerequisites

- Running Kubernetes cluster with kubeflow installed.
- `kubectl` configured to talk to the desired cluster.
- `git`

> **_NOTE:_** This documentation assumes that you are running the commands in linux.
If you are using another OS, please make sure to update the Makefile commands. 

### To update the kserve manifests to specific version follow the below instructions.

1. Set the desired version to upgrade.

   ```sh
   export MODELS_WEBAPP_VERSION=0.8.1
   ```

2. Rebuild the manifests.

   ```sh
   make upgrade-models-webapp-manifests
   ```

3. Install the updated manifests.
   ```sh
   make install-models-webapp
   ```
> **_NOTE:_** If resource/crd installation fails please re-run the commands.

### Testing
For testing refer [kserve readme](README.md#testing-models-webapp).