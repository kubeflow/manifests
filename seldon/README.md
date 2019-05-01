# Seldon Kustomize 

## Install Seldon Namespaced
Install Seldon Core in the namespace of your choice.

```
kubectl create namespace test1
kustomize build overlays/namespaced/ | kubectl apply --namespace test1 -f -
```

## Install Seldon Clusterwide

The kustomize application assumes the namespace `seldon-system`. 

```
kubectl create namespace seldon-system
kustomize build overlays/clusterwide/ | kubectl apply -f -
```