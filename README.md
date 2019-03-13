# manifests
A repository for Kustomize manifests

## Install Kustomize

`go get -u github.com/kubernetes-sigs/kustomize`


## Basic Usage

```bash
git clone https://github.com/kubeflow/manifests
kustomize build | kubectl apply -f
```

### Bridging kustomize and ksonnet

Equivalent to parameters in ksonnet, kustomize has vars. But the customizable objects are limited to [this list](https://github.com/kubernetes-sigs/kustomize/blob/master/pkg/transformers/config/defaultconfig/varreference.go)



### Installing to a custom namespace

For example, to install in `kubeflow-dev`. From the root of the repo run:

```bash
kustomize edit set namespace kubeflow-dev
```

## List of Kubeflow components available

* Ambassador

* Argo

* Profiles
