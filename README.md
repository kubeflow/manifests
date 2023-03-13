# Kubeflow Manifests

## Table of Contents

<!-- toc -->

- [Overview](#overview)
- [Kubeflow components versions](#kubeflow-components-versions)
- [Installation](#installation)
  * [Prerequisites](#prerequisites)
  * [Install with a single command](#install-with-a-single-command)
  * [Install individual components](#install-individual-components)
  * [Connect to your Kubeflow Cluster](#connect-to-your-kubeflow-cluster)
  * [Change default user password](#change-default-user-password)
- [Release process](#release-process)
- [Frequently Asked Questions](#frequently-asked-questions)

<!-- tocstop -->

## Overview

This repo is owned by the [Manifests Working Group](https://github.com/kubeflow/community/blob/master/wg-manifests/charter.md).
If you are a contributor authoring or editing the packages please see [Best Practices](./docs/KustomizeBestPractices.md).

The Kubeflow Manifests repository is organized under three (3) main directories, which include manifests for installing:

| Directory | Purpose |
| - | - |
| `apps` | Kubeflow's official components, as maintained by the respective Kubeflow WGs |
| `common` | Common services, as maintained by the Manifests WG |
| `contrib` | 3rd party contributed applications, which are maintained externally and are not part of a Kubeflow WG |

The `distributions` directory contains manifests for specific, opinionated distributions of Kubeflow, and will be phased out during the 1.4 release, [since going forward distributions will maintain their manifests on their respective external repositories](https://github.com/kubeflow/community/blob/master/proposals/kubeflow-distributions.md).

The `docs`, `hack`, and `tests` directories will also be gradually phased out.

Starting from Kubeflow 1.3, all components should be deployable using `kustomize` only. Any automation tooling for deployment on top of the manifests should be maintained externally by distribution owners.

## Kubeflow components versions

### Kubeflow Version: latest

This repo periodically syncs all official Kubeflow components from their respective upstream repos. The following matrix shows the git version that we include for each component:

| Component | Local Manifests Path | Upstream Revision |
| - | - | - |
| Training Operator | apps/training-operator/upstream | [v1.6.0-rc.0](https://github.com/kubeflow/training-operator/tree/v1.6.0-rc.0/manifests) |
| Notebook Controller | apps/jupyter/notebook-controller/upstream | [v1.7.0-rc.0](https://github.com/kubeflow/kubeflow/tree/v1.7.0-rc.0/components/notebook-controller/config) |
| Tensorboard Controller | apps/tensorboard/tensorboard-controller/upstream | [v1.7.0-rc.0](https://github.com/kubeflow/kubeflow/tree/v1.7.0-rc.0/components/tensorboard-controller/config) |
| Central Dashboard | apps/centraldashboard/upstream | [v1.7.0-rc.0](https://github.com/kubeflow/kubeflow/tree/v1.7.0-rc.0/components/centraldashboard/manifests) |
| Profiles + KFAM | apps/profiles/upstream | [v1.7.0-rc.0](https://github.com/kubeflow/kubeflow/tree/v1.7.0-rc.0/components/profile-controller/config) |
| PodDefaults Webhook | apps/admission-webhook/upstream | [v1.7.0-rc.0](https://github.com/kubeflow/kubeflow/tree/v1.7.0-rc.0/components/admission-webhook/manifests) |
| Jupyter Web App | apps/jupyter/jupyter-web-app/upstream | [v1.7.0-rc.0](https://github.com/kubeflow/kubeflow/tree/v1.7.0-rc.0/components/crud-web-apps/jupyter/manifests) |
| Tensorboards Web App | apps/tensorboard/tensorboards-web-app/upstream | [v1.7.0-rc.0](https://github.com/kubeflow/kubeflow/tree/v1.7.0-rc.0/components/crud-web-apps/tensorboards/manifests) |
| Volumes Web App | apps/volumes-web-app/upstream | [v1.7.0-rc.0](https://github.com/kubeflow/kubeflow/tree/v1.7.0-rc.0/components/crud-web-apps/volumes/manifests) |
| Katib | apps/katib/upstream | [v0.15.0-rc.0](https://github.com/kubeflow/katib/tree/v0.15.0-rc.0/manifests/v1beta1) |
| KServe | contrib/kserve/kserve | [v0.10.0](https://github.com/kserve/kserve/tree/v0.10.0/install/v0.10.0) |
| KServe Models Web App | contrib/kserve/models-web-app | [v0.10.0](https://github.com/kserve/models-web-app/tree/v0.10.0/config) |
| Kubeflow Pipelines | apps/pipeline/upstream | [2.0.0-alpha.7](https://github.com/kubeflow/pipelines/tree/2.0.0-alpha.7/manifests/kustomize) |
| Kubeflow Tekton Pipelines | apps/kfp-tekton/upstream | [v1.5.1](https://github.com/kubeflow/kfp-tekton/tree/v1.5.1/manifests/kustomize) |

The following is also a matrix with versions from common components that are
used from the different projects of Kubeflow:

| Component | Local Manifests Path | Upstream Revision |
| - | - | - |
| Istio | common/istio-1-16 | [1.16.0](https://github.com/istio/istio/releases/tag/1.16.0) |
| Knative | common/knative/knative-serving <br /> common/knative/knative-eventing | [1.8.1](https://github.com/knative/serving/releases/tag/knative-v1.8.1) <br /> [1.8.1](https://github.com/knative/eventing/releases/tag/knative-v1.8.1) |
| Cert Manager | common/cert-manager | [1.10.1](https://github.com/cert-manager/cert-manager/releases/tag/v1.10.1) |

## Installation

The Manifests WG provides two options for installing Kubeflow official components and common services with kustomize. The aim is to help end users install easily and to help distribution owners build their opinionated distributions from a tested starting point:

1. Single-command installation of all components under `apps` and `common`
2. Multi-command, individual components installation for `apps` and `common`

Option 1 targets ease of deployment for end users. \
Option 2 targets customization and ability to pick and choose individual components.

The `example` directory contains an example kustomization for the single command to be able to run.

:warning: In both options, we use a default email (`user@example.com`) and password (`12341234`). For any production Kubeflow deployment, you should change the default password by following [the relevant section](#change-default-user-password).

### Prerequisites

- `Kubernetes` (up to `1.25`) with a default [StorageClass](https://kubernetes.io/docs/concepts/storage/storage-classes/)
- `kustomize` [5.0.0](https://github.com/kubernetes-sigs/kustomize/releases/tag/kustomize%2Fv5.0.0)
    - :warning: Kubeflow is not compatible with earlier versions of Kustomize. This is because we need the [`sortOptions`](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/sortoptions/) field, which is only available in Kustomize 5 and onwards https://github.com/kubeflow/manifests/issues/2388.
- `kubectl`

---
**NOTE**

`kubectl apply` commands may fail on the first try. This is inherent in how Kubernetes and `kubectl` work (e.g., CR must be created after CRD becomes ready). The solution is to simply re-run the command until it succeeds. For the single-line command, we have included a bash one-liner to retry the command.

The reason we do `awk '!/well-defined/'` is because there's a regression in Kustomize 5 and a line is printed in stdout and not stderr https://github.com/kubernetes-sigs/kustomize/issues/5039. We'll remove this command once a future patch version of Kustomize is available.

---

### Install with a single command

You can install all Kubeflow official components (residing under `apps`) and all common services (residing under `common`) using the following command:

```sh
while ! kustomize build example | awk '!/well-defined/' | kubectl apply -f -; do echo "Retrying to apply resources"; sleep 10; done
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

cert-manager is used by many Kubeflow components to provide certificates for
admission webhooks.

Install cert-manager:

```sh
kustomize build common/cert-manager/cert-manager/base | kubectl apply -f -
kubectl wait --for=condition=ready pod -l 'app in (cert-manager,webhook)' --timeout=180s -n cert-manager
kustomize build common/cert-manager/kubeflow-issuer/base | kubectl apply -f -
```

In case you get this error:
```
Error from server (InternalError): error when creating "STDIN": Internal error occurred: failed calling webhook "webhook.cert-manager.io": failed to call webhook: Post "https://cert-manager-webhook.cert-manager.svc:443/mutate?timeout=10s": dial tcp 10.96.202.64:443: connect: connection refused
```
This is because the webhook is not yet ready to receive request. Wait a couple seconds and retry applying the manfiests.

For more troubleshooting info also check out https://cert-manager.io/docs/troubleshooting/webhook/

#### Istio

Istio is used by many Kubeflow components to secure their traffic, enforce
network authorization and implement routing policies.

Install Istio:

```sh
kustomize build common/istio-1-16/istio-crds/base | kubectl apply -f -
kustomize build common/istio-1-16/istio-namespace/base | kubectl apply -f -
kustomize build common/istio-1-16/istio-install/base | kubectl apply -f -
```

#### Dex

Dex is an OpenID Connect Identity (OIDC) with multiple authentication backends. In this default installation, it includes a static user with email `user@example.com`. By default, the user's password is `12341234`. For any production Kubeflow deployment, you should change the default password by following [the relevant section](#change-default-user-password).

Install Dex:

```sh
kustomize build common/dex/overlays/istio | kubectl apply -f -
```

#### OIDC AuthService

The OIDC AuthService extends your Istio Ingress-Gateway capabilities, to be able to function as an OIDC client:

```sh
kustomize build common/oidc-authservice/base | kubectl apply -f -
```

#### Knative

Knative is used by the KServe official Kubeflow component.

Install Knative Serving:

```sh
kustomize build common/knative/knative-serving/overlays/gateways | kubectl apply -f -
kustomize build common/istio-1-16/cluster-local-gateway/base | kubectl apply -f -
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

#### Kubeflow Roles

Create the Kubeflow ClusterRoles, `kubeflow-view`, `kubeflow-edit` and
`kubeflow-admin`. Kubeflow components aggregate permissions to these
ClusterRoles.

Install kubeflow roles:

```sh
kustomize build common/kubeflow-roles/base | kubectl apply -f -
```

#### Kubeflow Istio Resources

Create the Istio resources needed by Kubeflow. This kustomization currently
creates an Istio Gateway named `kubeflow-gateway`, in namespace `kubeflow`.
If you want to install with your own Istio, then you need this kustomization as
well.

Install istio resources:

```sh
kustomize build common/istio-1-16/kubeflow-istio-resources/base | kubectl apply -f -
```

#### Kubeflow Pipelines

Install the [Multi-User Kubeflow Pipelines](https://www.kubeflow.org/docs/components/pipelines/multi-user/) official Kubeflow component:

```sh
kustomize build apps/pipeline/upstream/env/cert-manager/platform-agnostic-multi-user | awk '!/well-defined/' | kubectl apply -f -
```
This installs argo with the safe-to use runasnonroot emissary executor.  Please note that the installer is still responsible to analyze the security issues that arise when containers are run with root access and to decide if the kubeflow pipeline main containers are run as runasnonroot. It is strongly recommended that the pipelines main containers are installed and run as runasnonroot and without any special capabilities to mitigate security risks.

Do not use the deprecated and insecure PNS executor anymore
```sh
kustomize build apps/pipeline/upstream/env/platform-agnostic-multi-user-pns | kubectl apply -f -
```

Refer to [argo workflow executor documentation](https://argoproj.github.io/argo-workflows/workflow-executors) for further reasoning.

**Multi-User Kubeflow Pipelines dependencies**

* Istio + Kubeflow Istio Resources
* Kubeflow Roles
* OIDC Auth Service (or cloud provider specific auth service)
* Profiles + KFAM

**Alternative: Kubeflow Pipelines Standalone**

You can install [Kubeflow Pipelines Standalone](https://www.kubeflow.org/docs/components/pipelines/installation/standalone-deployment/) which

* does not support multi user separation
* has no dependencies on the other services mentioned here

You can learn more about their differences in [Installation Options for Kubeflow Pipelines
](https://www.kubeflow.org/docs/components/pipelines/installation/overview/).

Besides installation instructions in Kubeflow Pipelines Standalone documentation, you need to apply two virtual services to expose [Kubeflow Pipelines UI](https://github.com/kubeflow/pipelines/blob/1.7.0/manifests/kustomize/base/installs/multi-user/virtual-service.yaml) and [Metadata API](https://github.com/kubeflow/pipelines/blob/1.7.0/manifests/kustomize/base/metadata/options/istio/virtual-service.yaml) in kubeflow-gateway.

#### KServe

KFServing was rebranded to KServe.

Install the KServe component:

```sh
kustomize build contrib/kserve/kserve | kubectl apply -f -
```

Install the Models web app:

```sh
kustomize build contrib/kserve/models-web-app/overlays/kubeflow | kubectl apply -f -
```

- ../contrib/kserve/models-web-app/overlays/kubeflow

#### Katib

Install the Katib official Kubeflow component:

```sh
kustomize build apps/katib/upstream/installs/katib-with-kubeflow | kubectl apply -f -
```

#### Central Dashboard

Install the Central Dashboard official Kubeflow component:

```sh
kustomize build apps/centraldashboard/upstream/overlays/kserve | kubectl apply -f -
```

#### Admission Webhook

Install the Admission Webhook for PodDefaults:

```sh
kustomize build apps/admission-webhook/upstream/overlays/cert-manager | kubectl apply -f -
```

#### Notebooks

Install the Notebook Controller official Kubeflow component:

```sh
kustomize build apps/jupyter/notebook-controller/upstream/overlays/kubeflow | kubectl apply -f -
```

Install the Jupyter Web App official Kubeflow component:

```sh
kustomize build apps/jupyter/jupyter-web-app/upstream/overlays/istio | kubectl apply -f -
```

#### Profiles + KFAM

Install the Profile Controller and the Kubeflow Access-Management (KFAM) official Kubeflow
components:

```sh
kustomize build apps/profiles/upstream/overlays/kubeflow | kubectl apply -f -
```

#### Volumes Web App

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

#### User Namespace

Finally, create a new namespace for the the default user (named `kubeflow-user-example-com`).

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
2. Login with the default user's credential. The default email address is `user@example.com` and the default password is `12341234`.

#### NodePort / LoadBalancer / Ingress

In order to connect to Kubeflow using NodePort / LoadBalancer / Ingress, you need to setup HTTPS. The reason is that many of our web apps (e.g., Tensorboard Web App, Jupyter Web App, Katib UI) use [Secure Cookies](https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies#restrict_access_to_cookies), so accessing Kubeflow with HTTP over a non-localhost domain does not work.

Exposing your Kubeflow cluster with proper HTTPS is a process heavily dependent on your environment. For this reason, please take a look at the available [Kubeflow distributions](https://www.kubeflow.org/docs/started/installing-kubeflow/#install-a-packaged-kubeflow-distribution), which are targeted to specific environments, and select the one that fits your needs.

---
**NOTE**

If you absolutely need to expose Kubeflow over HTTP, you can disable the `Secure Cookies` feature by setting the `APP_SECURE_COOKIES` environment variable to `false` in every relevant web app. This is not recommended, as it poses security risks.

---

### Change default user password

For security reasons, we don't want to use the default password for the default Kubeflow user when installing in security-sensitive environments. Instead, you should define your own password before deploying. To define a password for the default user:

1. Pick a password for the default user, with email `user@example.com`, and hash it using `bcrypt`:

    ```sh
    python3 -c 'from passlib.hash import bcrypt; import getpass; print(bcrypt.using(rounds=12, ident="2y").hash(getpass.getpass()))'
    ```

2. Edit `common/dex/base/config-map.yaml` and fill the relevant field with the hash of the password you chose:

    ```yaml
    ...
      staticPasswords:
      - email: user@example.com
        hash: <enter the generated hash here>
    ```

## Release process

The Manifest Working Group releases Kubeflow based on the [release timeline](https://github.com/kubeflow/community/blob/master/releases/handbook.md#timeline).
 The community and the release team work closely with the Manifest Working Group to define the specific dates at the start of the [release cycle](https://github.com/kubeflow/community/blob/master/releases/handbook.md#releasing)
 and follow the [release versioning policy](https://github.com/kubeflow/community/blob/master/releases/handbook.md#versioning-policy),
 as defined in the [Kubeflow release handbook](https://github.com/kubeflow/community/blob/master/releases/handbook.md).

## Frequently Asked Questions

- **Q:** What versions of Istio, Knative, Cert-Manager, Argo, ... are compatible with Kubeflow? \
  **A:** Please refer to each individual component's documentation for a dependency compatibility range. For Istio, Knative, Dex, Cert-Manager and OIDC-AuthService, the versions in `common` are the ones we have validated.
- **Q:** Can I use earlier version of Kustomize with Kubeflow manifests?
  **A:** The manual installation instructions work with Kustomize 3.2. To use the one-liner installation you'll need to comment out the `sortOptions` section in the `example/kustomization.yaml`.
