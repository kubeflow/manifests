# Model Catalog Manifests

To deploy the model catalog:

```sh
kubectl apply -k . -n NAMESPACE
```

Replace `NAMESPACE` with your desired Kubernetes namespace.

Update `sources.yaml` and `sample-catalog.yaml` to configure catalog models.
