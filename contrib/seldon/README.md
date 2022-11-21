Please note: This component is **unmaintained and out-of-date**.

If the component fails to meet the [contrib requirements](https://github.com/kubeflow/manifests/blob/master/proposals/20220926-contrib-component-guidelines.md#component-requirements)
 by the next Kubeflow release ([1.7](https://github.com/kubeflow/community/tree/master/releases/release-1.7#timeline)),
 it will be removed from the [`manifest`](https://github.com/kubeflow/manifests) repository.

Updates to the `/contrib` components can be found in the [tracking issue](https://github.com/kubeflow/manifests/issues/2311).


# Seldon Kustomize 

## Install Seldon Operator

 * The yaml assumes you will install in kubeflow namespace
 * You need to have installed istio first

```
kustomize build seldon-core-operator/base | kubectl apply -n kubeflow -f -
```

## Updating

This kustomize spec was created from the seldon-core-operator helm chart with:

```
make clean seldon-core-operator/base
```
