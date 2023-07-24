# Cert Manager

## Upgrade Cert Manager Manifests

The manifests for Cert Manager are based off the following:

  - [Cert Manager (v1.12.2)](https://github.com/cert-manager/cert-manager/releases/tag/v1.12.2)

1. Download the cert manager yaml with the following commands:

    ```sh
    # No need to install cert-manager-crds.
    export CERT_MANAGER_VERSION='1.12.2'
    wget -O ./cert-manager/base/cert-manager.yaml "https://github.com/cert-manager/cert-manager/releases/download/v${CERT_MANAGER_VERSION}/cert-manager.yaml"
    ```