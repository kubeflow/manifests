# Kustomize Overlay for M2M Authentication Integration with Istio

## Overview

This kustomize overlay facilitates the integration of M2M (Machine-to-Machine) authentication
with Istio in a Kubernetes environment. It includes components designed to enable effective
authentication within the service mesh.

## Components

The overlay consists of:

1. **`istio-m2m`**: Configures Istio to trust JWTs in M2M communication, essential for enabling
   M2M authentication within the service mesh.

2. **`component-overwrite-m2m-token-issuer`**: This component is used when the OIDC issuer used
   by Kubernetes is external to the cluster. It allows for the modification of the default
   issuer URL in the M2M authentication setup to match the externally defined OIDC issuer. This
   adjustment is necessary due to kustomize's handling of config map generators, requiring a
   separate component for proper configuration merging.

## Usage Scenario

- **External OIDC Issuer Integration**: In cases where the OIDC Issuer integrated with Kubernetes
   is defined outside the cluster, `component-overwrite-m2m-token-issuer` is used to update the
   issuer URL for M2M authentication. This scenario is common in setups where an external OIDC
   issuer is preferred over the default Kubernetes self-served OIDC issuer.