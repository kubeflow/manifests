# Kubeflow Authentication and Authorization Prototype

This implementation's target platforms are Kubernetes clusters with access to modify Kubernetes' API config file, which is generally possible with on Premise installations of Kubernetes.

**Note**: This setup assumes Kubeflow Pipelines is setup in namespace kubeflow and Istio is already setup in the Kubernetes cluster.

## High Level Diagram
![Authentication and Authorization in Kubeflow](/docs/dex-auth/assets/auth-istio.png)


## Create SSL Certificates

This example is going to require three domains:  
- dex.example.org: For the authentication server
- login.example.org: For the client application for authentication through dex (optional)
- ldap-admin.example.org: For the admin interface to create LDAP users and groups (optional)

**Note**: Replace *example.org* with your own domain.  

With your trusted certificate signing authority, please create a certificate for the above domains.

### Why Self Signed SSL Certs will not work

Authentication through OIDC in Kubernetes does work with self signed certificates since the `--oidc-ca-file` parameter in the Kubernetes API server allows for adding a trusted CA for your authentication server.

Though Istio's authentication policy parameter `jwksUri` for [End User Authentication](https://istio.io/docs/ops/security/end-user-auth/) does [not allow self signed certificates](https://github.com/istio/istio/issues/7290#issuecomment-420748056).

Please generate certificates with a trusted authority for enabling this example or follow this [work-around](#work-around-a-way-to-use-self-signed-certificates).

## Server Setup Instructions

### Authentication Server Setup

#### Setup Post Certificate Creation

*TODO*(krishnadurai): Make this a part of kfctl

`kubectl create secret tls dex.example.com.tls --cert=ssl/cert.pem --key=ssl/key.pem -n auth`

Replace `dex.example.com.tls` with your own domain.

#### Parameterizing the setup

##### Variables in params environment files [dex-authenticator](dex-authenticator/base/params.env), [dex-crds](dex-crds/base/params.env) and [istio](/docs/dex-auth/examples/authentication/Istio):
 - dex_domain: Domain for your dex server
 - issuer: Issuer URL for dex server
 - ldap_host: URL for LDAP server for dex to connect to
 - dex_client_id: ID for the dex client application
 - oidc_redirect_uris: Redirect URIs for OIDC client callback
 - dex_application_secret: Application secret for dex client
 - jwks_uri: URL pointing to the hosted JWKS keys
 - cluster_name: Name for your Kubernetes Cluster for dex to refer to
 - dex_client_redirect_uri: Single redirect URI for OIDC client callback
 - k8s_master_uri: Kubernetes API master server's URI
 - dex_client_listen_addr: Listen address for dex client to login


##### Certificate files:

- [dex_ca_contents.pem](dex-crds/base/dex_ca_contents.pem): CA cert generated for dex.
- [idp_ca_contents.pem](dex-authenticator/base/idp_ca_contents.pem): CA cert generated for dex, minding the required indentation as this is substituted in a string.
- [k8s_ca_contents.pem](dex-authenticator/base/k8s_ca_contents.pem): CA cert for your Kubernetes cluster.

##### This kustomize configs sets up:
 - A Dex server with LDAP IdP and a client application (dex-k8s-authenticator) for issuing keys for Dex.

#### Apply Kustomize Configs

**LDAP**

```
cd dex-ldap
kustomize build . | kubectl apply -f -
```

**Dex**

```
cd dex-crds
kustomize build . | kubectl apply -f -
```

**Dex Authentication Client**

```
cd dex-authenticator
kustomize build . | kubectl apply -f -
```

### Setup Kubernetes OIDC Authentication

The following parameters need to be set in Kubernetes API Server configuration file usually found in: `/etc/kubernetes/manifests/kube-apiserver.yaml`.

- --oidc-issuer-url=https://dex.example.org:32000
- --oidc-client-id=ldapdexapp
- --oidc-ca-file=/etc/ssl/certs/openid-ca.pem
- --oidc-username-claim=email
- --oidc-groups-claim=groups

`oidc-ca-file` needs to have the path to the file containing the certificate authority for the dex server's domain: dex.example.com.

Refer [official documentation](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#configuring-the-api-server) for meanings of these parameters.

When you have added these flags, Kubernetes should restart kube-apiserver pod. If not, run this command: `sudo systemctl restart kubelet` in your Kubernetes API Server master node. You can check flags in the pod description:

`kubectl describe pod kube-apiserver -n kube-system`


## Work-around: A way to use Self-Signed Certificates

* Change the following three entries in *[alt_names]* section in `/docs/dex-auth/examples/gencert.sh` file to reflect your own domains:
  * dex.example.org
  * login.example.org
  * ldap-admin.example.org


* Execute `examples/gencert.sh` on your terminal and it should create a folder `ssl` containing all required self signed certificates.

* Copy the JWKS keys from `https://dex.example.com/keys` and host these keys in a public repository as a file. This public repository should have a verified a https SSL certificate (for e.g. github).

* Copy the file url from the public repository in the `jwks_uri` parameter for [Istio Authentication Policy](/docs/dex-auth/examples/authentication/Istio/params.env) config:

```
jwks_uri="https://raw.githubusercontent.com/example-organisation/jwks/master/auth-jwks.json"
```

* Note that this is just a work around and JWKS keys are rotated by the Authentication Server. These JWKS keys will become invalid after the rotation period and you will have to re-upload the new keys back to your public repository.
