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
GKE mounts `/opt/cni/bin` as read-only for security reasons, preventing the Istio CNI installer from writing the CNI binary. Use the GKE-specific overlay: `kubectl apply -k common/istio-1-26/istio-install/overlays/gke`. This overlay uses GKE's writable CNI directory at `/home/kubernetes/bin`. For more details, see [Istio CNI Prerequisites](https://istio.io/latest/docs/setup/additional-setup/cni/#prerequisites) and [Platform Prerequisites](https://istio.io/latest/docs/ambient/install/platform-prerequisites/)

#### For Google Kubernetes Engine clusters:
```bash
kubectl apply -k istio-install/overlays/gke
```

### OAuth2-proxy integration
For clusters with oauth2-proxy authentication:
```bash
kubectl apply -k istio-install/overlays/oauth2-proxy
```

## Switching Between Modes

**Important**: You must delete the current installation before switching to avoid resource conflicts.

### Switch from CNI to Non-CNI
```bash
kubectl delete -k istio-install/base/
kubectl apply -k istio-install/overlays/insecure/
```

### Switch from Non-CNI to CNI
```bash
kubectl delete -k istio-install/overlays/insecure/
kubectl apply -k istio-install/base/
```

### Verify the switch
```bash
# Check for CNI DaemonSet (should exist only in CNI mode)
kubectl get daemonset -n kube-system istio-cni-node

# Check Istio pods are running
kubectl get pods -n istio-system
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
For upgrading Istio to newer versions, use the synchronization script:

```bash
scripts/synchronize-istio-cni-manifests.sh
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