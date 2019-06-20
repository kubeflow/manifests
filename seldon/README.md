# Seldon Kustomize 

## Install Seldon Operator

```
kustomize build seldon-core-operator/base | kubectl apply -f -
```

## Updating

This kustomize spec was created from the seldon-core-operator helm chart with:

```
git clone git@github.com:SeldonIO/seldon-core.git
helm convert -f values.yaml seldon-core/helm-charts/seldon-core-operator --skip-transformers=image,secret,namePrefix --namespace kubeflow
cd seldon-core-operator && mv *.yaml base
```

