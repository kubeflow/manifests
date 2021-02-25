# Feast Kustomize 

## Generating/Updating Feast Kustomize

```
kustomize build feast/base | kubectl apply -n kubeflow -f -
```

## Updating

The Feast Kustomize configuration in this folder is built from the Feast Helm charts and a custom values.yaml file.

Run the following command to regenerate the configuration:
```
make feast/base
```