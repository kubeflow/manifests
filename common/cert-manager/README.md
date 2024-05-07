# Cert Manager

## Upgrade Cert Manager Manifests

```sh
# No need to install cert-manager-crds.
export CERT_MANAGER_VERSION='1.14.5'
wget -O ./cert-manager/base/cert-manager.yaml "https://github.com/cert-manager/cert-manager/releases/download/v${CERT_MANAGER_VERSION}/cert-manager.yaml"
```