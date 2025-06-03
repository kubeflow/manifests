# Istio 1.26

This uses Istio 1.26 with CNI as the default configuration as described here <https://istio.io/latest/docs/setup/additional-setup/cni/>.

CNI eliminates privileged init containers and improves security compliance with Pod Security Standards. This configuration also enables native sidecars for Istio through the `ENABLE_NATIVE_SIDECARS=true` environment variable in istiod.

## Installation Options

### Default (CNI-enabled - Recommended)
```bash
kubectl apply -k istio-install/base
```

### Insecure Istio (CNI-disabled)
For environments that don't support CNI:
```bash
kubectl apply -k istio-install/overlays/insecure
```

### GKE-specific CNI
GKE mounts `/opt/cni/bin` as read-only for security reasons, preventing the Istio CNI installer from writing the CNI binary. Use the GKE-specific overlay: `kubectl apply -k common/istio-1-26/istio-install/overlays/gke`. This overlay uses GKE's writable CNI directory at `/home/kubernetes/bin`. For more details, see [Istio CNI Prerequisites](https://istio.io/latest/docs/setup/additional-setup/cni/#prerequisites) and [Platform Prerequisites](https://istio.io/latest/docs/ambient/install/platform-prerequisites/).-`
For Google Kubernetes Engine clusters:
```bash
kubectl apply -k istio-install/overlays/gke
```

### OAuth2-proxy integration
For clusters with oauth2-proxy authentication:
```bash
kubectl apply -k istio-install/overlays/oauth2-proxy
```

## CNI Benefits

- **Security**: No privileged init containers required
- **Compatibility**: Better alignment with Pod Security Standards
- **Performance**: Native sidecars support for improved lifecycle management
- **Simplicity**: Reduces container complexity
- **Startup Time**: Significantly faster startup in many cases

## Troubleshooting

With native sidecars enabled, init containers should access the network through the Istio proxy. If you encounter issues with KServe and init containers:

1. Use `runAsUser: 1337` in your init containers, OR
2. Add annotation `traffic.sidecar.istio.io/excludeOutboundIPRanges: 0.0.0.0/0` to your KServe inferenceservices

## Upgrade Istio Manifests

Istio ships with an installer called `istioctl`, which is a deployment /
debugging / configuration management tool for Istio all in one package.
In this section, we explain how to upgrade our istio kustomize packages
by leveraging `istioctl`. Assuming the new version is `X.Y.Z` and the
old version is `X1.Y1.Z1`:

1. Make a copy of the old istio manifests tree, which will become the
    kustomization for the new Istio version:

        export MANIFESTS_SRC=<path/to/manifests/repo>
        export ISTIO_OLD=$MANIFESTS_SRC/common/istio-1-26
        export ISTIO_NEW=$MANIFESTS_SRC/common/istio-X-Y
        cp -a $ISTIO_OLD $ISTIO_NEW

2. Download `istioctl` for version `X.Y.Z`:

        $ ISTIO_VERSION="X.Y.Z"
        $ wget "https://github.com/istio/istio/releases/download/${ISTIO_VERSION}/istio-${ISTIO_VERSION}-linux-amd64.tar.gz"
        $ tar xvfz istio-${ISTIO_VERSION}-linux-amd64.tar.gz
        # sudo mv istio-${ISTIO_VERSION}/bin/istioctl /usr/local/bin/istioctl

3. Generate manifests and add them to their respective packages. We
    will generate manifests using `istioctl`, the
    `profile.yaml` file from upstream and the
    `profile-overlay.yaml` file that contains our desired
    changes:

        export PATH="$MANIFESTS_SRC/scripts:$PATH"
        cd $ISTIO_NEW
        istioctl manifest generate -f profile.yaml -f profile-overlay.yaml --set components.cni.enabled=true --set components.cni.namespace=kube-system > dump.yaml
        ./split-istio-packages -f dump.yaml
        mv $ISTIO_NEW/crd.yaml $ISTIO_NEW/istio-crds/base
        mv $ISTIO_NEW/install.yaml $ISTIO_NEW/istio-install/base
        mv $ISTIO_NEW/cluster-local-gateway.yaml $ISTIO_NEW/cluster-local-gateway/base
        rm dump.yaml

    ---
    **NOTE**

    `split-istio-packages` is a python script in the same folder as this file.
    The `ruamel.yaml` version used is 0.16.12.

    `--cluster-specific` is a flag that determines if a current K8s cluster context will be used to dynamically detect default settings. Ensure you have a target cluster ready before running the above commands.
    We target Kubernetes 1.32+ for compatibility. The `--cluster-specific` flag helps ensure generated resources are compatible with your cluster version and configuration.

    ---

## Changes to Istio's upstream manifests

### Profile modifications

- Add `cluster-local-gateway` component for KServe
- Disable EgressGateway component
- Enable CNI by default

### Kustomize modifications

- Remove PodDisruptionBudgets for compatibility
- Add AuthorizationPolicies for security
- Add Gateway CRs and namespace objects
- Configure TCP KeepAlives
- Disable tracing to prevent DNS issues
- Set `ENABLE_DEBUG_ON_HTTP=false` for security
- Add seccomp profiles for Pod Security Standards compliance