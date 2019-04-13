# manifests
A repository of kustomize components for kubeflow

## Organization

### Groupings

The groupings of components is the same as the earlier ksonnet packages. 

```
argo          ⇲
common        ⇲
              ⎹→ambassador
              ⎹→basic-auth
              ⎹→centraldashboard
              ⎹→echo-server
              ⎹→spartakus
gcp           ⇲                                   
              ⎹→cert-manager
              ⎹→cloud-endpoints
              ⎹→gcp-credentials-admission-webhook
              ⎹→gpu-driver
              ⎹→iap-ingress
              ⎹→metric-collector
              ⎹→prometheus
jupyter        ⇲                                   
              ⎹→jupyter-web-app
              ⎹→notebook-controller
katib         ⇲                                   
kubebench     ⇲                                   
metacontroller⇲                                   
pipeline      ⇲                                   
              ⎹→minio
              ⎹→mysql
              ⎹→persistent-agent
              ⎹→pipelines-runner
              ⎹→pipelines-ui
              ⎹→pipelines-viewer
              ⎹→scheduledworkflow
```

## Install Kustomize

`go get -u github.com/kubernetes-sigs/kustomize`

## Basic Usage

```bash
git clone https://github.com/kubeflow/manifests
cd manifests
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

* CentralDashboard

* Argo

* Profiles

* Katib
