# Istio

We use Istio with CNI as the default configuration as described here <https://istio.io/latest/docs/setup/additional-setup/cni/>.

CNI eliminates privileged init containers and improves security compliance with Pod Security Standards. This configuration also enables native sidecars for Istio through the `ENABLE_NATIVE_SIDECARS=true` environment variable in istiod.

## Installation Options

### Default (CNI-enabled - Recommended)
```bash
kubectl apply -k istio-install/overlays/oauth2-proxy
```

### GKE-specific CNI
GKE mounts `/opt/cni/bin` as read-only for security reasons, preventing the Istio CNI installer from writing the CNI binary. Use the GKE-specific overlay: `kubectl apply -k common/istio/istio-install/overlays/gke`. This overlay uses GKE's writable CNI directory at `/home/kubernetes/bin`. For more details, see [Istio CNI Prerequisites](https://istio.io/latest/docs/setup/additional-setup/cni/#prerequisites) and [Platform Prerequisites](https://istio.io/latest/docs/ambient/install/platform-prerequisites/)

#### For Google Kubernetes Engine clusters:
```bash
kubectl apply -k istio-install/overlays/gke
```

### Insecure Istio (CNI-disabled)
For environments that don't support CNI:
```bash
kubectl apply -k istio-install/overlays/insecure
```

## CNI Benefits

- **Security**: No privileged init containers required
- **Compatibility**: Better alignment with Pod Security Standards
- **Performance**: Native sidecars support for improved lifecycle management
- **Simplicity**: Reduces container complexity
- **Startup Time**: Significantly faster startup in many cases

## Troubleshooting

If you still encounter probelms, even with native sidecars enabled, you might try the following:

1. Use `runAsUser: 1337` in your init containers, or
2. Add the annotation `traffic.sidecar.istio.io/excludeOutboundIPRanges: 0.0.0.0/0` to your KServe inferenceservices

## Upgrade Istio Manifests
For upgrading Istio to newer versions, use the synchronization script:

```bash
scripts/synchronize-istio-manifests.sh
```

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
