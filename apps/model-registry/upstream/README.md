# Install Kubeflow Model Registry

This folder contains [Kubeflow Model Registry](https://www.kubeflow.org/docs/components/model-registry/installation/) Kustomize manifests

## Installation

To install Kubeflow Model Registry, follow [Kubeflow Model Registry deployment documentation](https://www.kubeflow.org/docs/components/model-registry/installation/)

The following instructions will summarize how to deploy Model Registry as separate component in the context of a default Kubeflow >=1.9 installation.
Ensure you are running these commands from the directory containing this README.md file (e.g.: you could check with `pwd`).

```bash
kubectl apply -k overlays/db
```

As the default Kubeflow installation provides an Istio mesh, apply the necessary manifests:

```bash
kubectl apply -k options/istio
```

Check everything is up and running:

```bash
kubectl wait --for=condition=available -n kubeflow deployment/model-registry-deployment --timeout=2m
kubectl logs -n kubeflow deployment/model-registry-deployment
```

Optionally, you can also port-forward the REST API container port of Model Registry to interact with it from your terminal:

```bash
kubectl port-forward svc/model-registry-service -n kubeflow 8081:8080
```

And then, from another terminal:

```bash
curl -sX 'GET' \
  'http://localhost:8081/api/model_registry/v1alpha3/registered_models?pageSize=100&orderBy=ID&sortOrder=DESC' \
  -H 'accept: application/json' | jq
```

### UI Installation

There are two main ways to deploy the Model Registry UI:

1. Standalone mode - Use this if you are using Model Registry without the Kubeflow Platform

2. Integrated mode - Use this if you are deploying Model Registry in Kubeflow 

For a standalone install run the following command:

```bash
kubectl apply -k options/ui/overlays/standalone -n kubeflow
```

For an integrated install use the istio UI overlay:

```bash
kubectl apply -k options/ui/overlays/istio -n kubeflow
```

## Usage

For a basic usage of the Kubeflow Model Registry, follow the [Kubeflow Model Registry getting started documentation](https://www.kubeflow.org/docs/components/model-registry/getting-started/)

## Uninstall

To uninstall the Kubeflow Model Registry run:

```bash
# Delete istio options
kubectl delete -k options/istio

# Delete model registry db and deployment
kubectl delete -k overlays/db
```