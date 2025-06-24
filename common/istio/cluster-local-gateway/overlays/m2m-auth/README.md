# KServe JWT Authentication for cluster-local-gateway

This overlay secures the `cluster-local-gateway` with JWT authentication to address issue #2811.

## Security Features

1. **JWT Validation**: Adds `RequestAuthentication` to validate Kubernetes service account tokens
2. **Authorization Enforcement**: Replaces the permissive base `AuthorizationPolicy` with one that requires valid JWT principals
3. **Health Check Exemption**: Allows health check and metrics endpoints without authentication

## Changes Made

### RequestAuthentication
- Validates JWT tokens from Kubernetes API server
- Extracts `sub` claim to `kubeflow-userid` header
- Supports both `https://kubernetes.default.svc.cluster.local` and `https://kubernetes.default.svc` issuers

### AuthorizationPolicy
- **DENY policy**: Blocks requests without valid JWT principals (except health checks)
- **ALLOW policy**: Permits requests with valid JWT principals and health check endpoints

## Impact on KServe

With this configuration:
- KServe InferenceServices require valid JWT tokens for access
- Service accounts from the same namespace can access their services
- Cross-namespace access requires proper JWT tokens
- Individual AuthorizationPolicies per InferenceService are no longer needed

## Testing

Use the `tests/kserve_secure_test.sh` script to verify:
1. Requests with valid tokens succeed (200)
2. Requests without tokens fail (403)
3. Requests from unauthorized namespaces fail (403)

## Installation

This overlay is automatically used by `tests/knative-cni_install.sh`:

```bash
kustomize build common/istio/cluster-local-gateway/overlays/m2m-auth | kubectl apply -f -
```