# Kubeflow Manifests

## Table of Contents

<!-- toc -->

- [Overview](#overview)
- [Installation](#installation)
  * [Prerequisites](#prerequisites)
  * [Install with a single command](#install-with-a-single-command)
  * [Install individual components](#install-individual-components)
  * [Connect to your Kubeflow Cluster](#connect-to-your-kubeflow-cluster)
  * [Change default user password](#change-default-user-password)

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

Starting Kubeflow 1.3, all components should be deployable using `kustomize` only. Any automation tooling for deployment on top of the manifests should be maintained externally by distribution owners.

## Installation

Starting Kubeflow 1.3, the Manifests WG provides two options for installing Kubeflow official components and common services with kustomize. The aim is to help end users install easily and to help distribution owners build their opinionated distributions from a tested starting point:

1. Single-command installation of all components under `apps` and `common`
2. Multi-command, individual components installation for `apps` and `common`

Option 1 targets ease of deployment for end users. \
Option 2 targets customization and ability to pick and choose individual components.

The `example` directory contains an example kustomization for the single command to be able to run.

:warning: In both options, we use a default username (`user`) and password (`12341234`). For any production Kubeflow deployment, you should change the default password by following [the relevant section](#change-default-user-password).

### Prerequisites

- `Kubernetes` (tested with version `1.17`)
- `kustomize` (version `3.2.0`)
- `kubectl`

---
**NOTE**

`kubectl apply` commands may fail on the first try. This is inherent in how Kubernetes and `kubectl` work (e.g., CR must be created after CRD becomes ready). The solution is to simply re-run the command until it succeeds. For the single-line command, we have included a bash one-liner to retry the command.

---

### Install with a single command

You can install all Kubeflow official components (residing under `apps`) and all common services (residing under `common`) using the following command:


```sh
while ! kustomize build --load_restrictor=none example | kubectl apply -f -; do echo "Retrying to apply resources"; sleep 10; done
```

Once, everything is installed successfully, you can access the Kubeflow Central Dashboard [by logging in to your cluster](#connect-to-your-kubeflow-cluster).

Congratulations! You can now start experimenting and running your end-to-end ML workflows with Kubeflow.

### Install individual components

In this section, we will install each Kubeflow official component (under `apps`) and each common service (under `common`) separately, using just `kubectl` and `kustomize`.

If all the following commands are executed, the result is the same as in the above section of the single command installation. The purpose of this section is to:

- Provide a description of each component and insight on how it gets installed.
- Enable the user or distribution owner to pick and choose only the components they need.

#### cert-manager

cert-manager is used by many Kubeflow components to provide certificates for
admission webhooks.

Install cert-manager:
```sh
kustomize build --load_restrictor=none common/cert-manager/cert-manager-kube-system-resources/base | kubectl apply -f -
kustomize build --load_restrictor=none common/cert-manager/cert-manager-crds/base | kubectl apply -f -
kustomize build --load_restrictor=none common/cert-manager/cert-manager/overlays/self-signed | kubectl apply -f -
```

#### Istio

Istio is used by many Kubeflow components to secure their traffic, enforce
network authorization and implement routing policies.

Install Istio:
```sh
kustomize build --load_restrictor=none common/istio-1-9-0/istio-crds/base | kubectl apply -f -
kustomize build --load_restrictor=none common/istio-1-9-0/istio-namespace/base | kubectl apply -f -
kustomize build --load_restrictor=none common/istio-1-9-0/istio-install/base | kubectl apply -f -
```

#### Dex

Dex is an OpenID Connect Identity (OIDC) with multiple authentication backends. In this default installation, it includes a static user named `user`. By default, the user's password is `12341234`. For any production Kubeflow deployment, you should change the default password by following [the relevant section](#change-default-user-password).

Install Dex:

```sh
kustomize build --load_restrictor=none common/dex/overlays/istio | kubectl apply -f -
```

#### OIDC AuthService

The OIDC AuthService extends your Istio Ingress-Gateway capabilities, to be able to function as an OIDC client:

```sh
kustomize build --load_restrictor=none common/oidc-authservice/base | kubectl apply -f -
```

#### Knative

Knative is used by the KFServing official Kubeflow component.

Install Knative:
```sh
kustomize build --load_restrictor=none common/knative/knative-serving-crds/base | kubectl apply -f -
kustomize build --load_restrictor=none common/knative/knative-serving-install/base | kubectl apply -f -
kustomize build --load_restrictor=none common/knative/knative-eventing-crds/base | kubectl apply -f -
kustomize build --load_restrictor=none common/knative/knative-eventing-install/base | kubectl apply -f -
kustomize build --load_restrictor=none common/istio-1-9-0/cluster-local-gateway/base | kubectl apply -f -
```

#### Kubeflow Namespace

Create the namespace where the Kubeflow components will live in. This namespace
is named `kubeflow`.

Install kubeflow namespace:
```sh
kustomize build --load_restrictor=none common/kubeflow-namespace/base | kubectl apply -f -
```

#### Kubeflow Roles

Create the Kubeflow ClusterRoles, `kubeflow-view`, `kubeflow-edit` and
`kubeflow-admin`. Kubeflow components aggregate permissions to these
ClusterRoles.

Install kubeflow roles:
```sh
kustomize build --load_restrictor=none common/kubeflow-roles/base | kubectl apply -f -
```


#### Kubeflow Istio Resources

Create the Istio resources needed by Kubeflow. This kustomization currently
creates an Istio Gateway named `kubeflow-gateway`, in namespace `kubeflow`.
If you want to install with your own Istio, then you need this kustomization as
well.

Install istio resources:
```sh
kustomize build --load_restrictor=none common/istio-1-9-0/kubeflow-istio-resources/base | kubectl apply -f -
```

#### Kubeflow Pipelines

Install the Kubeflow Pipelines official Kubeflow component:

```sh
kustomize build --load_restrictor=none apps/pipeline/upstream/env/platform-agnostic-multi-user | kubectl apply -f -
```

#### KFServing

Install the KFServing official Kubeflow component:

```sh
kustomize build --load_restrictor=none apps/kfserving/upstream/overlays/kubeflow | kubectl apply -f -
```

#### Katib

Install the Katib official Kubeflow component:

```sh
kustomize build --load_restrictor=none apps/katib/upstream/installs/katib-with-kubeflow-cert-manager | kubectl apply -f -
```

#### Central Dashboard

Install the Central Dashboard official Kubeflow component:

```sh
kustomize build --load_restrictor=none apps/centraldashboard/upstream/overlays/istio | kubectl apply -f -
```

#### Admission Webhook

Install the Admission Webhook for PodDefaults:
```sh
kustomize build --load_restrictor=none apps/admission-webhook/upstream/overlays/cert-manager | kubectl apply -f -
```

#### Notebooks

Install the Notebook Controller official Kubeflow component:

```sh
kustomize build --load_restrictor=none apps/jupyter/notebook-controller/upstream/overlays/kubeflow | kubectl apply -f -
```

Install the Jupyter Web App official Kubeflow component:

```sh
kustomize build --load_restrictor=none apps/jupyter/jupyter-web-app/upstream/overlays/istio | kubectl apply -f -
```

#### Profiles + KFAM

Install the Profile Controller and the Kubeflow Access-Management (KFAM) official Kubeflow
components:

```sh
kustomize build --load_restrictor=none apps/profiles/upstream/overlays/kubeflow | kubectl apply -f -
```

#### Volumes Web App

Install the Volumes Web App official Kubeflow component:

```sh
kustomize build --load_restrictor=none apps/volumes-web-app/upstream/overlays/istio | kubectl apply -f -
```

#### Tensorboard

Install the Tensorboards Web App official Kubeflow component:

```sh
kustomize build --load_restrictor=none apps/tensorboard/tensorboards-web-app/upstream/overlays/istio | kubectl apply -f -
```

Install the Tensorboard Controller official Kubeflow component:
```sh
kustomize build --load_restrictor=none apps/tensorboard/tensorboard-controller/upstream/overlays/kubeflow | kubectl apply -f -
```

#### TFJob Operator

Install the TFJob Operator official Kubeflow component:

```sh
kustomize build --load_restrictor=none apps/tf-training/upstream/overlays/kubeflow | kubectl apply -f -
```

#### PyTorch Operator

Install the PyTorch Operator official Kubeflow component:

```sh
kustomize build --load_restrictor=none apps/tensorboard/tensorboard-controller/upstream/overlays/kubeflow | kubectl apply -f -
```

#### MPI Operator

Install the MPI Operator official Kubeflow component:

```sh
kustomize build --load_restrictor=none apps/mpi-job/upstream/overlays/kubeflow | kubectl apply -f -
```

#### MXNet Operator

Install the MXNet Operator official Kubeflow component:

```sh
kustomize build --load_restrictor=none apps/mxnet-job/upstream/overlays/kubeflow | kubectl apply -f -
```

#### XGBoost Operator

Install the XGBoost Operator official Kubeflow component:

```sh
kustomize build --load_restrictor=none apps/xgboost-job/upstream/overlays/kubeflow | kubectl apply -f -
```

#### User Namespace

Finally, create a new namespace for the the default user (named `user`).

```sh
kustomize build --load_restrictor=none common/user-namespace/base | kubectl apply -f -
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
kubectl get pods -n kubeflow-user
```

The default way of accessing Kubeflow is via port-forward. This enables you to get started quickly without imposing any requirements on your environment. Run the following to port-forward Istio's Ingress-Gateway to local port `8080`:

```sh
kubectl port-forward svc/istio-ingressgateway -n istio-system 8080:80
```

After running the command, you can access the Kubeflow Central Dashboard by doing the following:
1. Open your browser and visit `http://localhost:8080`. You should get the Dex login screen.
2. Login with the default user's credential. The default username is `user` and the default password is `12341234`.

### Change default user password

For security reasons, we don't want to use the default password for the default Kubeflow user when installing in security-sensitive environments. Instead, you should define your own password before deploying. To define a password for the default user:

1. Pick a password for the default user, with handle `user`, and hash it using `bcrypt`:
    ```sh
    python3 -c 'from passlib.hash import bcrypt; import getpass; print(bcrypt.using(rounds=12, ident="2y").hash(getpass.getpass()))'
    ```

2. Edit `dex/base/config-map.yaml` and fill the relevant field with the hash of the password you chose:
    ```yaml
    ...
      staticPasswords:
      - email: user
        hash: <enter the generated hash here>
    ```
