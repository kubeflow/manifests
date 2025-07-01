# KServe JWT Authentication

## Overview

This implementation provides secure JWT-based authentication for KServe inference services, addressing issue #2811. The solution ensures that all access to KServe models requires valid Kubernetes service account tokens.

## Architecture

### Before (Insecure)
```
Request → cluster-local-gateway → KServe Service
         (no authentication)     (anonymous access)
```

### After (Secure)
```
Request → cluster-local-gateway → KServe Service
         (JWT required)         (authenticated access)
```

## Components

### 1. cluster-local-gateway Security
- **RequestAuthentication**: Validates Kubernetes service account JWT tokens
- **AuthorizationPolicy (DENY)**: Blocks requests without valid JWT principals
- **AuthorizationPolicy (ALLOW)**: Permits authenticated requests and health checks

### 2. Service-Level Authorization (Future Enhancement)
- Service-level policies can be added in future PRs for fine-grained access control
- This PR focuses on gateway-level JWT authentication only

## Configuration

### Gateway-Level Security (Always Required)

The m2m-auth overlay is automatically applied via `tests/knative-cni_install.sh`:

```bash
kustomize build common/istio/cluster-local-gateway/overlays/m2m-auth | kubectl apply -f -
```

### Service-Level Authorization (Future Enhancement)

Service-level namespace isolation policies will be added in future PRs.
This PR focuses on the core JWT authentication at the gateway level.

## Usage

### 1. Deploy InferenceService

```yaml
apiVersion: "serving.kserve.io/v1beta1"
kind: "InferenceService"
metadata:
  name: "my-model"
  namespace: kubeflow-user-example-com
spec:
  predictor:
    sklearn:
      storageUri: "gs://my-bucket/model"
```

### 2. Create JWT Token

```bash
TOKEN=$(kubectl -n kubeflow-user-example-com create token default-editor)
```

### 3. Access Model

```bash
# Internal access
curl -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     "http://my-model-predictor.kubeflow-user-example-com.svc.cluster.local/v1/models/my-model:predict" \
     -d '{"instances": [[1.0, 2.0, 3.0]]}'

# External access (if VirtualService configured)
curl -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     "http://your-cluster.com/kserve/kubeflow-user-example-com/my-model/v1/models/my-model:predict" \
     -d '{"instances": [[1.0, 2.0, 3.0]]}'
```

## Security Features

### Authentication Enforced
- All requests require valid JWT tokens
- Unauthenticated requests receive 403 RBAC: access denied
- Invalid tokens receive 401 Unauthorized

### Token Validation
- Validates Kubernetes service account tokens
- Supports multiple JWT issuers for compatibility
- Extracts user identity to headers

### Health Check Exemption
- `/healthz`, `/ready`, `/metrics` endpoints excluded
- System monitoring continues to work

### Gateway-Level Access Control
- All requests require valid JWT tokens
- Valid tokens allow access regardless of namespace
- Service-level isolation can be added in future PRs

## Testing

### Automated Tests

```bash
# Basic JWT validation
./tests/final_validation.sh

# Comprehensive testing (with KServe)
./tests/kserve_complete_auth_test.sh

# Knative service testing
./tests/knative_auth_test.sh
```

### Manual Testing

```bash
# Test without token (should fail)
curl -H "Host: my-service.namespace.svc.cluster.local" \
     "http://localhost:8080/v1/models/my-model:predict"
# Expected: 403

# Test with invalid token (should fail)  
curl -H "Host: my-service.namespace.svc.cluster.local" \
     -H "Authorization: Bearer invalid-token" \
     "http://localhost:8080/v1/models/my-model:predict"
# Expected: 401

# Test with valid token (should work)
TOKEN=$(kubectl -n namespace create token service-account)
curl -H "Host: my-service.namespace.svc.cluster.local" \
     -H "Authorization: Bearer $TOKEN" \
     "http://localhost:8080/v1/models/my-model:predict"
# Expected: 200 or 404 (if service issues)
```

## Migration Guide

### From Insecure Setup

1. **Apply the secure overlay**:
   ```bash
   kustomize build common/istio/cluster-local-gateway/overlays/m2m-auth | kubectl apply -f -
   ```

2. **Update client applications**:
   - Add JWT token to requests
   - Handle 401/403 responses appropriately

3. **Service-level policies** (future enhancement):
   - Namespace isolation can be added in future PRs
   - This PR provides gateway-level JWT authentication

### Troubleshooting

- **403 RBAC: access denied**: No JWT token provided
- **401 Unauthorized**: Invalid or expired JWT token  
- **404 Not Found**: JWT validated, but service not found or blocked by policies
- **503 Service Unavailable**: Service temporarily unavailable (scaling, etc.)

## Files Changed

- `common/istio/cluster-local-gateway/overlays/m2m-auth/` - JWT authentication overlay
- `tests/knative-cni_install.sh` - Updated to use secure overlay
- `tests/kserve_test.sh` - Removed individual AuthorizationPolicy creation
- `examples/` - Added configuration examples and templates

## Impact

- **Security**: No more anonymous access to KServe models
- **Compatibility**: Existing clients need JWT tokens
- **Performance**: Minimal overhead from JWT validation
- **Operations**: Health checks continue to work