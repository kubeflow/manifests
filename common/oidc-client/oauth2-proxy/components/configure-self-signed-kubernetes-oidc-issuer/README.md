# Configure Istio with Self-Signed Kubernetes OIDC Issuer

## Overview

This kustomize component is designed for scenarios where the Kubernetes OIDC Issuer is behind
self-signed certificates, causing trust issues with Istio in retrieving the JWKS. It creates a
Kubernetes Job and necessary RBAC configurations to address this issue. The job facilitates the
retrieval of JWKS from the OIDC Issuer and embeds it directly into Istio's configuration, thereby
bypassing trust issues with self-signed certificates.

## Configuration Persistence and JWKS Public Accessibility

The configuration created by this job is stored exclusively in etcd and is not persisted elsewhere.
This setup is compatible with ArgoCD, which by default does not delete properties absent in the
desired manifest. However, if the `RequestAuthentication` resource is modified erroneously or
deleted, the configuration would be lost, necessitating a rerun of the job.

To circumvent the need for rerunning the job, it is advisable to store the JWKS in a repository.
JWKS are typically publicly accessible and contain no sensitive information, making them safe for
repository storage. By persisting the JWKS in a repository, consistent access is ensured regardless
of changes in the cluster or accidental deletions.

## Functionality

- **Reading OIDC Issuer URL**: The Job reads the `RequestAuthentication` resource to extract the
  OIDC Issuer URL.
- **Fetching JWKS**: After identifying the Issuer URL, the Job retrieves the JWKS.
- **Patching RequestAuthentication**: The JWKS is used to patch the `RequestAuthentication`
  resource, embedding the JWKS directly.
- **Static JWKS Configuration for Istio**: Ensures that Istio uses the JWKS provided by the Job
  instead of requesting it independently.

## Use Case

This setup is particularly useful when:
- Kubernetes serves as an OIDC provider.
- The Kubernetes API is not served with publicly trusted certificates.

This component ensures seamless M2M authentication by handling JWKS retrieval and configuration
internally, thus circumventing certificate validation issues for Istio fetching JWKS from an OIDC
provider with self-signed or private CA certificates.
