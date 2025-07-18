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

## Cross-Namespace Access Control

Cross-namespace access control is not enforced at the service level due to architectural limitations:

1. **Gateway-Level Authentication**: This implementation validates JWT tokens at the `cluster-local-gateway` level, which means any valid Kubernetes service account token from any namespace can access services through the gateway.

2. **Knative Activator Bypass**: When Knative's activator is in the request path (during cold starts or scale-to-zero scenarios), it bypasses namespace-based authentication because:
   - The activator has broad access permissions to all services
   - The original request identity is lost when the activator forwards requests
   - This is a known limitation documented in [knative-extensions/net-istio#554](https://github.com/knative-extensions/net-istio/issues/554)

3. **Current Behavior**: Any service account from any namespace can access KServe models if they have a valid JWT token. This provides authentication but not authorization between namespaces.

4. **Not Fixable at Manifests Level**: This limitation requires architectural changes to Knative/KServe itself and cannot be resolved through Kubeflow manifests alone.
