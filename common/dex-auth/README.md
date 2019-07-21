# Kubeflow Authentication and Authorization Prototype

**Note**: This setup assumes Kubeflow Pipelines is setup in namespace kubeflow and Istio is already setup in the Kubernetes cluster.

## High Level Diagram
![Authentication and Authorization in Kubeflow](assets/auth-istio.png)


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

## Authentication Server Setup

### Setup Post Certificate Creation

`kubectl create ns auth`

`kubectl create secret tls dex.example.com.tls --cert=ssl/cert.pem --key=ssl/key.pem -n auth`

Replace `dex.example.com.tls` with your own domain.

### Editing Overlay File Values

Follow instructions [here](authentication/overlays/README.md) to edit Kustomize overlays in `authentication/overlays/prototype` to setup a Dex server with LDAP IdP and a client application (dex-k8s-authenticator) for issuing keys for Dex.

### Apply Kustomize Configs

`kustomize build authentication/overlays/prototype | kubectl apply -f -`

## Create Users and Groups in LDAP server

Follow instructions [here](authentication/base/ldap/README.md).

## Setup Kubernetes OIDC Authentication

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

## Setup Kubernetes RBAC

```
cd authorization/Kubernetes
kubectl create -f .
```

## Setup Istio Authentication Policy and RBAC

Currently, the only service authenticated and authorized supported is ml-pipeline service.
This example allows for authentication and authorization only for requests within the Kubernetes cluster. Istio version 1.3 will allow for application of RBAC rules to ingress requests to the cluster.

### Istio Authentication

```
cd authentication/Istio
```

Edit the file `authentication_policy.yaml` and replace the value for 'dex.example.org' in `issuer` and `jwksUri` with your dex server's domain.

```
kubectl create -f authentication_policy.yaml
cd ../..
```

### Istio RBAC Authorization
```
cd authorization/Istio
kubectl create -f .
cd ../..
```

## Work-around: A way to use Self-Signed Certificates

* Change the following three entries in *[alt_names]* section in `certs/gencert.sh` file to reflect your own domains:
  * dex.example.org
  * login.example.org
  * ldap-admin.example.org


* Execute `certs/gencert.sh` on your terminal and it should create a folder `ssl` containing all required self signed certificates.

* Copy the JWKS keys from `https://dex.example.com/keys` and host these keys in a public repository as a file. This public repository should have a verified a https SSL certificate (for e.g. github).

* Copy the file url from the public repository in the `jwksUri` field of Istio Authentication Policy config:

```
apiVersion: "authentication.istio.io/v1alpha1"
kind: "Policy"
metadata:
  name: "pipelines-auth-policy"
spec:
  targets:
  - name: ml-pipeline
  peers:
  - mtls: {}
  origins:
  - jwt:
      audiences:
        - "ldapdexapp"
      issuer: "https://org.example.com:32000"
      jwksUri: "https://raw.githubusercontent.com/example-organisation/jwks/master/auth-jwks.json"
  principalBinding: USE_ORIGIN
  targets:
  - name: ml-pipeline
```

* Note that this is just a work around and JWKS keys are rotated by the Authentication Server. These JWKS keys will become invalid after the rotation period and you will have to re-upload the new keys back to your public repository.
