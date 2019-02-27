# manifests
A repository for Kustomize manifests

## Install Kustomize

`go get -u github.com/kubernetes-sigs/kustomize`

## List of Kubeflow components available

* Ambassador

* Argo

* Jupyter

* Profiles

## Basic Usage

```bash
git clone https://github.com/kubeflow/manifests
kustomize build | kubectl apply -f
```
