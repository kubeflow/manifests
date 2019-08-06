# Using Examples to enable RBAC

This guide helps in setting up RBAC for Kubeflow.

The RBAC rules here assume 3 groups: admin, datascience and validator as sample groups for operating on Kubeflow.

## Setup Kubernetes RBAC

```
cd authorization/Kubernetes
kubectl create -f .
cd ../..
```

## Setup Istio Authentication

```
cd authentication/Istio
kustomize build base
```

## Setup Istio RBAC

Currently, the only service authenticated and authorized supported is ml-pipeline service.
This example allows for authentication and authorization only for requests within the Kubernetes cluster.

```
cd authorization/Istio
kubectl create -f .
cd ../..
```
