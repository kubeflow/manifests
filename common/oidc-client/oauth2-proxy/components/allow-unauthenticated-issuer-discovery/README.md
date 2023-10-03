# Unauthenticated issuer discovery

If you're running Kubernetes with kind, vCluster, minikube or some other
tool for local development on Kubernetes, there is a high change that the
Kubernetes OIDC Issuer is managed withing the cluster behind self-signed
certificates.

To use m2m tokens with oauth2-proxy and Istio, both tools have to perform
OIDC Connect Discovery on the Issuer of the Token.

Kubernetes provides means of OIDC Discovery under URI:
```
https://kubernetes.default.svc.cluster.local/.well-known/openid-configuration
```

Access to this endpoint is blocked by default by RBAC but can be easily enabled by
creating a `ClusterRoleBinding` to bind a predefined Cluster Role
`system:service-account-issuer-discovery` to the Group `system:unauthenticated`.
This is done within the resource `clusterrolebinding.unauthenticated-oidc-viewer.yaml`.

After this step we can call the endpoint and see the Issuer URL:
```bash
$ curl -k https://kubernetes.default.svc.cluster.local/.well-known/openid-configuration

# Example output in kind:
{"issuer":"https://kubernetes.default.svc.cluster.local","jwks_uri":"https://172.18.0.5:6443/openid/v1/jwks","response_types_supported":["id_token"],"subject_types_supported":["public"],"id_token_signing_alg_values_supported":["RS256"]}

# Example output in vCluster:
{"issuer":"https://kubernetes.default.svc.cluster.local","jwks_uri":"https://1.2.3.4:6443/openid/v1/jwks","response_types_supported":["id_token"],"subject_types_supported":["public"],"id_token_signing_alg_values_supported":["RS256"]}

# Example output in AWS EKS:
{"issuer":"https://oidc.eks.region.amazonaws.com/id/123abc","jwks_uri":"https://ip-1-2-3-4.eu-central-1.compute.internal:443/openid/v1/jwks","response_types_supported":["id_token"],"subject_types_supported":["public"],"id_token_signing_alg_values_supported":["RS256"]}
```

If you're running in `vCluster`, the access to the endpoint under `jwks_uri` (in
this example `https://1.2.3.4:6443/openid/v1/jwks`) is managed separately and just creating
the mentioned `ClusterRoleBinding` is not enough. The simplest way to overcome that is to
configure the `kube-apiserver` with `--anonymous-auth=true`. This can be done
by passing `--kube-apiserver-arg=anonymous-auth=true` to the list of arguments in
Helm Chart Values file:
```yaml
vcluster:
    extraArgs:
    - --kube-apiserver-arg=anonymous-auth=true
```
