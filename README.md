# Kubeflow Manifests

## Table of Contents

<!-- toc -->

- [Overview of the Kubeflow Platform](#overview)
- [Kubeflow components versions](#kubeflow-components-versions)
- [Installation](#installation)
  * [Prerequisites](#prerequisites)
  * [Install with a single command](#install-with-a-single-command)
  * [Install individual components](#install-individual-components)
  * [Connect to your Kubeflow Cluster](#connect-to-your-kubeflow-cluster)
  * [Change default user password](#change-default-user-password)
- [Upgrading and extending](#upgrading-and-extending)
- [Release process](#release-process)
- [CVE Scanning](#cve-scanning)
- [Frequently Asked Questions](#frequently-asked-questions)

<!-- tocstop -->

## Overview of the Kubeflow Platform

This repository is owned by the [Manifests Working Group](https://github.com/kubeflow/community/blob/master/wg-manifests/charter.md).
If you are a contributor authoring or editing the packages please see [Best Practices](https://kubectl.docs.kubernetes.io/references/kustomize/).
You can join the CNCF Slack and access our meetings at the [Kubeflow Community](https://www.kubeflow.org/docs/about/community/) website. Our channel on the CNCF Slack is here [**#kubeflow-platform**](https://app.slack.com/client/T08PSQ7BQ/C073W572LA2). You can also find there our [biweekly meetings](https://bit.ly/kf-wg-manifests-meet), including the commentable [Agenda](https://bit.ly/kf-wg-manifests-notes).

The Kubeflow Manifests repository is organized under three main directories, which include manifests for installing:

| Directory | Purpose |
| - | - |
| `apps` | Kubeflow's official components, as maintained by the respective Kubeflow WGs |
| `common` | Common services, as maintained by the Manifests WG |
| `contrib` | 3rd party contributed applications (e.g. Ray, Kserve), which are maintained externally and are not part of a Kubeflow WG |

All components are deployable with `kustomize`. You can choose to deploy the whole Kubeflow platform or individual components.

## Kubeflow components versions

### Kubeflow Version: master

This repo periodically syncs all official Kubeflow components from their respective upstream repos. The following matrix shows the git version that we include for each component:

| Component | Local Manifests Path | Upstream Revision |
| - | - | - |
| Training Operator | apps/training-operator/upstream | [v1.8.1](https://github.com/kubeflow/training-operator/tree/v1.8.1/manifests) |
| Notebook Controller | apps/jupyter/notebook-controller/upstream | [v1.9.2](https://github.com/kubeflow/kubeflow/tree/v1.9.2/components/notebook-controller/config) |
| PVC Viewer Controller | apps/pvcviewer-roller/upstream | [v1.9.2](https://github.com/kubeflow/kubeflow/tree/v1.9.2/components/pvcviewer-controller/config) |
| Tensorboard Controller | apps/tensorboard/tensorboard-controller/upstream | [v1.9.2](https://github.com/kubeflow/kubeflow/tree/v1.9.2/components/tensorboard-controller/config) |
| Central Dashboard | apps/centraldashboard/upstream | [v1.9.2](https://github.com/kubeflow/kubeflow/tree/v1.9.2/components/centraldashboard/manifests) |
| Profiles + KFAM | apps/profiles/upstream | [v1.9.2](https://github.com/kubeflow/kubeflow/tree/v1.9.2/components/profile-controller/config) |
| PodDefaults Webhook | apps/admission-webhook/upstream | [v1.9.2](https://github.com/kubeflow/kubeflow/tree/v1.9.2/components/admission-webhook/manifests) |
| Jupyter Web App | apps/jupyter/jupyter-web-app/upstream | [v1.9.2](https://github.com/kubeflow/kubeflow/tree/v1.9.2/components/crud-web-apps/jupyter/manifests) |
| Tensorboards Web App | apps/tensorboard/tensorboards-web-app/upstream | [v1.9.2](https://github.com/kubeflow/kubeflow/tree/v1.9.2/components/crud-web-apps/tensorboards/manifests) |
| Volumes Web App | apps/volumes-web-app/upstream | [v1.9.2](https://github.com/kubeflow/kubeflow/tree/v1.9.2/components/crud-web-apps/volumes/manifests) |
| Katib | apps/katib/upstream | [v0.17.0](https://github.com/kubeflow/katib/tree/v0.17.0/manifests/v1beta1) |
| KServe | contrib/kserve/kserve | [v0.14.0](https://github.com/kserve/kserve/releases/tag/v0.14.0/install/v0.14.0) |
| KServe Models Web App | contrib/kserve/models-web-app | [0.13.0](https://github.com/kserve/models-web-app/tree/0.13.0/config) |
| Kubeflow Pipelines | apps/pipeline/upstream | [2.3.0](https://github.com/kubeflow/pipelines/tree/2.3.0/manifests/kustomize) |
| Kubeflow Model Registry | apps/model-registry/upstream | [v0.2.10](https://github.com/kubeflow/model-registry/tree/v0.2.10/manifests/kustomize) |

The following is also a matrix with versions from common components that are
used from the different projects of Kubeflow:

| Component | Local Manifests Path | Upstream Revision |
| - | - | - |
| Istio | common/istio-1-23 | [1.23.2](https://github.com/istio/istio/releases/tag/1.23.2) |
| Knative | common/knative/knative-serving <br /> common/knative/knative-eventing | [v1.16.0](https://github.com/knative/serving/releases/tag/knative-v1.16.0) <br /> [v1.16.1](https://github.com/knative/eventing/releases/tag/knative-v1.16.1) |
| Cert Manager | common/cert-manager | [1.16.1](https://github.com/cert-manager/cert-manager/releases/tag/v1.16.1) |

## Installation

This is for the installation from scratch. For the in-place upgrade guide please jump to the [Upgrading and extending](#upgrading-and-extending) section.

The Manifests WG provides two options for installing Kubeflow official components and common services with kustomize. The aim is to help end users install easily and to help distribution owners build their opinionated distributions from a tested starting point:

1. Single-command installation of all components under `apps` and `common`
2. Multi-command, individual components installation for `apps` and `common`

Option 1 targets ease of deployment for end users. \
Option 2 targets customization and ability to pick and choose individual components.

The `example` directory contains an example kustomization for the single command to be able to run.

:warning: In both options, we use a default email (`user@example.com`) and password (`12341234`). For any production Kubeflow deployment, you should change the default password by following [the relevant section](#change-default-user-password).

### Prerequisites
- This is the master branch which targets Kubernetes around 1.31
- For the specific Kubernetes version per release consult the [release notes](https://github.com/kubeflow/manifests/releases)
- Either our local Kind (installed below) or your own Kubernetes cluster with a default [StorageClass](https://kubernetes.io/docs/concepts/storage/storage-classes/)
- Kustomize [5.2.1+](https://github.com/kubernetes-sigs/kustomize/releases/tag/kustomize%2Fv5.2.1)
- Kubectl in a version that is [compatible with your Kubernetes cluster](https://kubernetes.io/releases/version-skew-policy/#kubectl)

---
**NOTE**

`kubectl apply` commands may fail on the first try. This is inherent in how Kubernetes and `kubectl` work (e.g., CR must be created after CRD becomes ready). The solution is to simply re-run the command until it succeeds. For the single-line command, we have included a bash one-liner to retry the command.

---

### Install with a single command

#### Prerequisites
- 32 GB of RAM recommended
- 16 CPU cores recommended
- `kind`
- `docker`
- Linux kernel subsystem changes to support many pods
    - `sudo sysctl fs.inotify.max_user_instances=2280`
    - `sudo sysctl fs.inotify.max_user_watches=1255360`

#### Create kind cluster
```sh
cat <<EOF | kind create cluster --name=kubeflow --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  image: kindest/node:v1.31.0@sha256:53df588e04085fd41ae12de0c3fe4c72f7013bba32a20e7325357a1ac94ba865
  kubeadmConfigPatches:
  - |
    kind: ClusterConfiguration
    apiServer:
      extraArgs:
        "service-account-issuer": "kubernetes.default.svc"
        "service-account-signing-key-file": "/etc/kubernetes/pki/sa.key"
EOF
```

#### Save kubeconfig
```sh
kind get kubeconfig --name kubeflow > /tmp/kubeflow-config
export KUBECONFIG=/tmp/kubeflow-config
```

#### Create a Secret based on existing credentials in order to pull the images
```sh
docker login

kubectl create secret generic regcred \
    --from-file=.dockerconfigjson=$HOME/.docker/config.json \
    --type=kubernetes.io/dockerconfigjson
```

You can install all Kubeflow official components (residing under `apps`) and all common services (residing under `common`) using the following command:

```sh
while ! kustomize build example | kubectl apply --server-side --force-conflicts -f -; do echo "Retrying to apply resources"; sleep 20; done
```

Once, everything is installed successfully, you can access the Kubeflow Central Dashboard [by logging in to your cluster](#connect-to-your-kubeflow-cluster).

Congratulations! You can now start experimenting and running your end-to-end ML workflows with Kubeflow.

### Install individual components

In this section, we will install each Kubeflow official component (under `apps`) and each common service (under `common`) separately, using just `kubectl` and `kustomize`.

If all the following commands are executed, the result is the same as in the above section of the single command installation. The purpose of this section is to:

- Provide a description of each component and insight on how it gets installed.
- Enable the user or distribution owner to pick and choose only the components they need.

---
**Troubleshooting note**

We've seen errors like the following when applying the kustomizations of different components:
```
error: resource mapping not found for name: "<RESOURCE_NAME>" namespace: "<SOME_NAMESPACE>" from "STDIN": no matches for kind "<CRD_NAME>" in version "<CRD_FULL_NAME>"
ensure CRDs are installed first
```

This is because a kustomization applies both a CRD and a CR very quickly, and the CRD
hasn't become [`Established`](https://github.com/kubernetes/apiextensions-apiserver/blob/a7ee7f91a2d0805f729998b85680a20cfba208d2/pkg/apis/apiextensions/types.go#L276-L279) yet. You can learn more about this in https://github.com/kubernetes/kubectl/issues/1117 and https://github.com/helm/helm/issues/4925.

If you bump into this error we advise to re-apply the kustomization of the component.

---

#### cert-manager

Cert-manager is used by many Kubeflow components to provide certificates for
admission webhooks.

Install cert-manager:

```sh
kustomize build common/cert-manager/base | kubectl apply -f -
kustomize build common/cert-manager/kubeflow-issuer/base | kubectl apply -f -
echo "Waiting for cert-manager to be ready ..."
kubectl wait --for=condition=ready pod -l 'app in (cert-manager,webhook)' --timeout=180s -n cert-manager
kubectl wait --for=jsonpath='{.subsets[0].addresses[0].targetRef.kind}'=Pod endpoints -l 'app in (cert-manager,webhook)' --timeout=180s -n cert-manager
```

In case you get this error:
```
Error from server (InternalError): error when creating "STDIN": Internal error occurred: failed calling webhook "webhook.cert-manager.io": failed to call webhook: Post "https://cert-manager-webhook.cert-manager.svc:443/mutate?timeout=10s": dial tcp 10.96.202.64:443: connect: connection refused
```
This is because the webhook is not yet ready to receive request. Wait a couple seconds and retry applying the manfiests.

For more troubleshooting info also check out https://cert-manager.io/docs/troubleshooting/webhook/

#### Istio

Istio is used by most Kubeflow components to secure their traffic, enforce
network authorization and implement routing policies.
If you use Cilium CNI on your cluster, you have to configure it properly for Istio as shown [here](https://docs.cilium.io/en/latest/network/servicemesh/istio/), otherwise you will get RBAC access denied on the central dashboard.


Install Istio:

```sh
echo "Installing Istio configured with external authorization..."
kustomize build common/istio-1-23/istio-crds/base | kubectl apply -f -
kustomize build common/istio-1-23/istio-namespace/base | kubectl apply -f -
kustomize build common/istio-1-23/istio-install/overlays/oauth2-proxy | kubectl apply -f -

echo "Waiting for all Istio Pods to become ready..."
kubectl wait --for=condition=Ready pods --all -n istio-system --timeout 300s
```

#### Oauth2-proxy

The oauth2-proxy extends your Istio Ingress-Gateway capabilities, to be able to function as an OIDC client.
It supports user sessions as well as proper token-based machine to machine authentication.

```sh
echo "Installing oauth2-proxy..."

# Only uncomment ONE of the following overlays, they are mutually exclusive,
# see `common/oauth2-proxy/overlays/` for more options.

# OPTION 1: works on most clusters, does NOT allow K8s service account
#           tokens to be used from outside the cluster via the Istio ingress-gateway.
#
kustomize build common/oauth2-proxy/overlays/m2m-dex-only/ | kubectl apply -f -
kubectl wait --for=condition=ready pod -l 'app.kubernetes.io/name=oauth2-proxy' --timeout=180s -n oauth2-proxy

# Option 2: works on Kind, K3D, Rancher, GKE and many other clusters with the proper configuration, and allows K8s service account tokens to be used
#           from outside the cluster via the Istio ingress-gateway. For example for automation with github actions.
#           In the end you need to patch the issuer and jwksUri fields in the requestauthentication resource in the istio-system namespace 
#           as for example done in /common/oauth2-proxy/overlays/m2m-dex-and-kind/kustomization.yaml
#           Please follow the guidelines in the section Upgrading and extending below for patching.
#           curl --insecure -H "Authorization: Bearer `cat /var/run/secrets/kubernetes.io/serviceaccount/token`"  https://kubernetes.default/.well-known/openid-configuration
#           from a pod in the cluster should provide you with the issuer of your cluster.
# 
#kustomize build common/oauth2-proxy/overlays/m2m-dex-and-kind/ | kubectl apply -f -
#kubectl wait --for=condition=ready pod -l 'app.kubernetes.io/name=oauth2-proxy' --timeout=180s -n oauth2-proxy
#kubectl wait --for=condition=ready pod -l 'app.kubernetes.io/name=cluster-jwks-proxy' --timeout=180s -n istio-system

# OPTION 3: works on most EKS clusters with  K8s service account
#           tokens to be used from outside the cluster via the Istio ingress-gateway.
#           You have to adjust AWS_REGION and CLUSTER_ID in common/oauth2-proxy/overlays/m2m-dex-and-eks/ first.
#
#kustomize build common/oauth2-proxy/overlays/m2m-dex-and-eks/ | kubectl apply -f -
#kubectl wait --for=condition=ready pod -l 'app.kubernetes.io/name=oauth2-proxy' --timeout=180s -n oauth2-proxy
```

If and after you have finished the installation with Kubernetes serviceaccount token support you should be able to create and use the tokens:
```sh
kubectl port-forward svc/istio-ingressgateway -n istio-system 8080:80
TOKEN="$(kubectl -n $KF_PROFILE_NAMESPACE create token default-editor)"
client = kfp.Client(host="http://localhost:8080/pipeline", existing_token=token)
curl -v "localhost:8080/jupyter/api/namespaces/${$KF_PROFILE_NAMESPACE}/notebooks" -H "Authorization: Bearer ${TOKEN}"
```

If you want to use OAuth2 Proxy without Dex and conenct it directly to your own IDP, you can refer to this [document](common/oauth2-proxy/README.md#change-default-authentication-from-dex--oauth2-proxy-to-oauth2-proxy-only). But you can also keep Dex and extend it with connectors to your own IDP as explained in the Dex section below.


#### Dex

Dex is an OpenID Connect (OIDC) identity provider with multiple authentication backends. In this default installation, it includes a static user with email `user@example.com`. By default, the user's password is `12341234`. For any production Kubeflow deployment, you should change the default password by following [the relevant section](#change-default-user-password).

Install Dex:

```sh
echo "Installing Dex..."
kustomize build common/dex/overlays/oauth2-proxy | kubectl apply -f -
kubectl wait --for=condition=ready pods --all --timeout=180s -n auth
```

To connect to your desired identity providers (LDAP,GitHub,Google,Microsoft,OIDC,SAML,GitLab) please take a look at https://dexidp.io/docs/connectors/oidc/.
We recommend to use OIDC in general, since it is compatible with most providers as for example azure in the following example.
You need to modify https://github.com/kubeflow/manifests/blob/master/common/dex/overlays/oauth2-proxy/config-map.yaml and add some environment variables in https://github.com/kubeflow/manifests/blob/master/common/dex/base/deployment.yaml by adding a patch section in your main Kustomization file. For guidance please check out [Upgrading and extending](#upgrading-and-extending).

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: dex
data:
  config.yaml: |
    issuer: http://dex.auth.svc.cluster.local:5556/dex
    storage:
      type: kubernetes
      config:
        inCluster: true
    web:
      http: 0.0.0.0:5556
    logger:
      level: "debug"
      format: text
    oauth2:
      skipApprovalScreen: true
    enablePasswordDB: true
    #### WARNING YOU SHOULD NOT USE THE DEFAULT STATIC PASSWORDS
    #### and patch /common/dex/base/dex-passwords.yaml in a Kustomize overlay or remove it
    staticPasswords:
    - email: user@example.com
      hashFromEnv: DEX_USER_PASSWORD
      username: user
      userID: "15841185641784"
    staticClients:
    # https://github.com/dexidp/dex/pull/1664
    - idEnv: OIDC_CLIENT_ID
      redirectURIs: ["/oauth2/callback"]
      name: 'Dex Login Application'
      secretEnv: OIDC_CLIENT_SECRET
    #### Here come the connectors to OIDC providers such as Azure, GCP, GitHub, GitLab etc.
    #### Connector config values starting with a "$" will read from the environment.
    connectors:
    - type: oidc
      id: azure
      name: azure
      config:
        issuer: https://login.microsoftonline.com/$TENANT_ID/v2.0
        redirectURI: https://$KUBEFLOW_INGRESS_URL/dex/callback
        clientID: $AZURE_CLIENT_ID
        clientSecret: $AZURE_CLIENT_SECRET
        insecureSkipEmailVerified: true
        scopes:
        - openid
        - profile
        - email
        #- groups # groups might be used in the future
```

#### Knative

Knative is used by the KServe official Kubeflow component.

Install Knative Serving:

```sh
kustomize build common/knative/knative-serving/overlays/gateways | kubectl apply -f -
kustomize build common/istio-1-23/cluster-local-gateway/base | kubectl apply -f -
```

Optionally, you can install Knative Eventing which can be used for inference request logging:

```sh
kustomize build common/knative/knative-eventing/base | kubectl apply -f -
```

#### Kubeflow Namespace

Create the namespace where the Kubeflow components will live in. This namespace
is named `kubeflow`.

Install kubeflow namespace:

```sh
kustomize build common/kubeflow-namespace/base | kubectl apply -f -
```

#### Network Policies

Install network policies:
```sh
kustomize build common/networkpolicies/base | kubectl apply -f -
```

#### Kubeflow Roles

Create the Kubeflow ClusterRoles, `kubeflow-view`, `kubeflow-edit` and
`kubeflow-admin`. Kubeflow components aggregate permissions to these
ClusterRoles.

Install kubeflow roles:

```sh
kustomize build common/kubeflow-roles/base | kubectl apply -f -
```

#### Kubeflow Istio Resources

Create the Kubeflow Gateway, `kubeflow-gateway` and ClusterRole, 
`kubeflow-istio-admin`.

Install kubeflow istio resources:

```sh
kustomize build common/istio-1-23/kubeflow-istio-resources/base | kubectl apply -f -
```

#### Kubeflow Pipelines

Install the [Multi-User Kubeflow Pipelines](https://www.kubeflow.org/docs/components/pipelines/multi-user/) official Kubeflow component:

```sh
kustomize build apps/pipeline/upstream/env/cert-manager/platform-agnostic-multi-user | kubectl apply -f -
```
This installs argo with the runasnonroot emissary executor. Please note that you are still responsible to analyze the security issues that arise when containers are run with root access and to decide if the kubeflow pipeline main containers are run as runasnonroot. It is in general strongly recommended that all user-accessible OCI containers run with Pod Security Standards [restricted](https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted).

#### KServe

KFServing was rebranded to KServe.

Install the KServe component:

```sh
kustomize build contrib/kserve/kserve | kubectl apply --server-side --force-conflicts -f -
```

Install the Models web application:

```sh
kustomize build contrib/kserve/models-web-app/overlays/kubeflow | kubectl apply -f -
```

#### Katib

Install the Katib official Kubeflow component:

```sh
kustomize build apps/katib/upstream/installs/katib-with-kubeflow | kubectl apply -f -
```

#### Central Dashboard

Install the Central Dashboard official Kubeflow component:

```sh
kustomize build apps/centraldashboard/overlays/oauth2-proxy | kubectl apply -f -
```

#### Admission Webhook

Install the Admission Webhook for PodDefaults:

```sh
kustomize build apps/admission-webhook/upstream/overlays/cert-manager | kubectl apply -f -
```

#### Notebooks 1.0

Install the Notebook Controller official Kubeflow component:

```sh
kustomize build apps/jupyter/notebook-controller/upstream/overlays/kubeflow | kubectl apply -f -
```

Install the Jupyter Web App official Kubeflow component:

```sh
kustomize build apps/jupyter/jupyter-web-app/upstream/overlays/istio | kubectl apply -f -
```

#### Workspaces (Notebooks 2.0)

It is still in development.

#### PVC Viewer Controller 

Install the PVC Viewer Controller official Kubeflow component:

```sh
kustomize build apps/pvcviewer-controller/upstream/default | kubectl apply -f -
```

#### Profiles + KFAM

Install the Profile Controller and the Kubeflow Access-Management (KFAM) official Kubeflow
components:

```sh
kustomize build apps/profiles/upstream/overlays/kubeflow | kubectl apply -f -
```

#### Volumes Web Application

Install the Volumes Web App official Kubeflow component:

```sh
kustomize build apps/volumes-web-app/upstream/overlays/istio | kubectl apply -f -
```

#### Tensorboard

Install the Tensorboards Web App official Kubeflow component:

```sh
kustomize build apps/tensorboard/tensorboards-web-app/upstream/overlays/istio | kubectl apply -f -
```

Install the Tensorboard Controller official Kubeflow component:

```sh
kustomize build apps/tensorboard/tensorboard-controller/upstream/overlays/kubeflow | kubectl apply -f -
```

#### Training Operator

Install the Training Operator official Kubeflow component:

```sh
kustomize build apps/training-operator/upstream/overlays/kubeflow | kubectl apply -f -
```

#### User Namespaces

Finally, create a new namespace for the default user (named `kubeflow-user-example-com`).

```sh
kustomize build common/user-namespace/base | kubectl apply -f -
```

### Connect to your Kubeflow Cluster

After installation, it will take some time for all Pods to become ready. Make sure all Pods are ready before trying to connect, otherwise you might get unexpected errors. To check that all Kubeflow-related Pods are ready, use the following commands:

```sh
kubectl get pods -n cert-manager
kubectl get pods -n istio-system
kubectl get pods -n auth
kubectl get pods -n knative-eventing
kubectl get pods -n knative-serving
kubectl get pods -n kubeflow
kubectl get pods -n kubeflow-user-example-com
```

#### Port-Forward

The default way of accessing Kubeflow is via port-forward. This enables you to get started quickly without imposing any requirements on your environment. Run the following to port-forward Istio's Ingress-Gateway to local port `8080`:

```sh
kubectl port-forward svc/istio-ingressgateway -n istio-system 8080:80
```

After running the command, you can access the Kubeflow Central Dashboard by doing the following:

1. Open your browser and visit `http://localhost:8080`. You should get the Dex login screen.
2. Login with the default user's credentials. The default email address is `user@example.com` and the default password is `12341234`.

#### NodePort / LoadBalancer / Ingress

In order to connect to Kubeflow using NodePort / LoadBalancer / Ingress, you need to setup HTTPS. The reason is that many of our web applications (e.g., Tensorboard Web Application, Jupyter Web Application, Katib UI) use [Secure Cookies](https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies#restrict_access_to_cookies), so accessing Kubeflow with HTTP over a non-localhost domain does not work.

Exposing your Kubeflow cluster with proper HTTPS is a simple proces, but dependent on your environment. 
There are also third-party commercial [distributions](https://www.kubeflow.org/docs/started/installing-kubeflow/#install-a-packaged-kubeflow-distribution) available.

---
**NOTE**

If you absolutely need to expose Kubeflow over HTTP, you can disable the `Secure Cookies` feature by setting the `APP_SECURE_COOKIES` environment variable to `false` in every relevant web app. This is not recommended, as it poses security risks.

---

### Change default user name

For security reasons, we don't want to use the default username and email for the default Kubeflow user when installing in security-sensitive environments. Instead, you should define your own username and email before deploying. To define it for the default user:

1. Edit `common/dex/overlays/oauth2-proxy/config-map.yaml` and fill the relevant field with your email and preferred username:

    ```yaml
    ...
      staticPasswords:
      - email: <REPLACE_WITH_YOUR_EMAIL>
        username: <REPLACE_WITH_PREFERRED_USERNAME>
    ```

### Change default user password

If you have an identy provider (LDAP,GitHub,Google,Microsoft,OIDC,SAML,GitLab) available you should use that instead of static passwords and connect it to oauth2-proxy or Dex as explained in the sections above. This is best practices instead of using static passwords. 

For security reasons, we don't want to use the default static password for the default Kubeflow user when installing in security-sensitive environments. Instead, you should define your own password and apply it either **before creating the cluster** or **after creating the cluster**. 

Pick a password for the default user, with email `user@example.com`, and hash it using `bcrypt`:

    ```sh
    python3 -c 'from passlib.hash import bcrypt; import getpass; print(bcrypt.using(rounds=12, ident="2y").hash(getpass.getpass()))'
    ```

For example, running the above command locally with required packages like _passlib_ would look as follows:
  ```sh
  python3 -c 'from passlib.hash import bcrypt; import getpass; print(bcrypt.using(rounds=12, ident="2y").hash(getpass.getpass()))'
  Password:       <--- Enter the password here
  $2y$12$vIm8CANhuWui0J1p3jYeGeuM28Qcn76IFMaFWvZCG5ZkKZ4MjTF4u <--- GENERATED_HASH_FOR_ENTERED_PASSWORD
  ```

#### Before creating the cluster:

1. Edit `common/dex/base/dex-passwords.yaml` and fill the relevant field with the hash of the password you chose:

    ```yaml
    ...
      stringData:
        DEX_USER_PASSWORD: <REPLACE_WITH_HASH>
    ```

#### After creating the cluster:

1. Delete the existing secret _dex-passwords_ in auth namespace using the following command:

    ```sh
    kubectl delete secret dex-passwords -n auth
    ```

2. Create secret dex-passwords with new hash using the following command:

    ```sh
    kubectl create secret generic dex-passwords --from-literal=DEX_USER_PASSWORD='REPLACE_WITH_HASH' -n auth
    ```

3. Recreate the _dex_ pod in auth namespace using the following command:

    ```sh
    kubectl delete pods --all -n auth
    ```

4. Try to login using the new dex password.



## Upgrading and extending

For modifications and in place upgrades of the Kubeflow platform we provide a rough description for advanced users:

- Never ever edit the manifests directly, use Kustomize overlays and [components](https://github.com/kubernetes-sigs/kustomize/blob/master/examples/components.md) on top of the [example.yaml](https://github.com/kubeflow/manifests/blob/master/example/kustomization.yaml).
- This allows you to upgrade by just referencing the new manifests, building with kustomize and running `kubectl apply` again.
- You might have to adjust your over the top overlays and components if needed.
- You might have to prune old resources. For that you would add [labels](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/labels/) to all your resources from the start.
- With labels you can use `kubectl apply` with `--prune` and `--dry-run` to list prunable resources.
- Sometimes there are major changes, e.g. in the 1.9 release we switch to oauth2-proxy, which need additional attention.
- Nevertheless with a bit of Kubernetes knowledge one should be able to upgrade.

## Release process

The Manifest Working Group releases Kubeflow based on the [release timeline](https://github.com/kubeflow/community/blob/master/releases/handbook.md#timeline).
 The community and the release team work closely with the Manifest Working Group to define the specific dates at the start of the [release cycle](https://github.com/kubeflow/community/blob/master/releases/handbook.md#releasing)
 and follow the [release versioning policy](https://github.com/kubeflow/community/blob/master/releases/handbook.md#versioning-policy),
 as defined in the [Kubeflow release handbook](https://github.com/kubeflow/community/blob/master/releases/handbook.md).

## CVE Scanning

To view all past security scans, head to the [Image Extracting and Security Scanning GitHub Action workflow](https://github.com/kubeflow/manifests/actions/workflows/trivy.yaml). In the logs of the workflow you can expand the `Run image extracting and security scanning script` step to view the CVE logs. You will find a per-image CVE scan and a JSON dump of per-WorkingGroup aggregated metrics.
You can run the Python script from the workflow file locally on your machine to obtain the detailed JSON files for any git commit.

The Kubeflow security working group follows a responsible disclosure policy for CVE results:

- **Internal Review**: All CVE findings are initially reviewed internally by the security working group.
- **Severity Assessment**: Each CVE is assessed for severity and potential impact on the Kubeflow project.
- **Disclosure**: For high and critical severity CVEs, the security working group will:
  - Notify the maintainers and contributors
  - Try to provide a fix or mitigation strategy
  - Publicly disclose the CVE details

## Frequently Asked Questions

- **Q:** What versions of Istio, Knative, Cert-Manager, Argo, ... are compatible with Kubeflow? \
  **A:** Please refer to each individual component's documentation for a dependency compatibility range. For Istio, Knative, Dex, Cert-Manager and OAuth2 Proxy, the versions in `common` are the ones we have validated.
- **Q:** Can I use earlier version of Kustomize with Kubeflow manifests?
  **A:** No, it is not supported anymore, although it might be possible with manual effort.
