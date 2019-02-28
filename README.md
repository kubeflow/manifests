# manifests
A repository for Kustomize manifests

## Install Kustomize

`go get -u github.com/kubernetes-sigs/kustomize`

## Basic Usage

```bash
git clone https://github.com/kubeflow/manifests
kustomize build | kubectl apply -f
```
