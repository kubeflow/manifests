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

Kustomize currently can't set namespaces inside a spec field which is done for ambassador's deployment.
So in order to set the namespace there, edit the `ambassador/overlays/namespace/ambassador-deployment-patch.yaml`
and set the field `value` to your custom namespace.

Edit the `kustomization.yaml` at the root of the file and change `ambassador/base` to `ambassador/overlays/namespace`.

### Adding custom images to components

To set the ambassador image to `quay.io/datawire/ambassador:0.37.0`,
From the root of the repo run:

```bash
kustomize edit set image ambassador/quay.io/datawire/ambassador:0.37.0
```

And similarly for the metacontroller image to `jsonnetd@sha256:25c25f217ad030a0f67e37078c33194785b494569b0c088d8df4f00da8fd15a0`,
From the root of the repo run:

```bash
kustomize edit set image metacontroller/jsonnetd@sha256:25c25f217ad030a0f67e37078c33194785b494569b0c088d8df4f00da8fd15a0
```

## List of Kubeflow components available

* Ambassador

### ksonnet parameters
    
```json
ambassador: {
      ambassadorImage: "quay.io/datawire/ambassador:0.37.0",
      ambassadorNodePort: 0,
      ambassadorServiceType: "ClusterIP",
      name: "ambassador",
      platform: "none",
      replicas: 3,
      namespace: "kubeflow"
    },
```

[Link to ksonnet package](https://github.com/kubeflow/kubeflow/blob/master/kubeflow/common/ambassador.libsonnet)

* Argo

### ksonnet parameters

```json
argo: {
    namespace: "kubeflow"
}
```

* Jupyter

### ksonnet parameters

```json
jupyter: {
      accessLocalFs: "false",
      disks: "null",
      gcpSecretName: "user-gcp-sa",
      image: "gcr.io/kubeflow/jupyterhub-k8s:v20180531-3bb991b1",
      jupyterHubAuthenticator: "null",
      name: "jupyter",
      notebookGid: "-1",
      notebookUid: "-1",
      platform: "none",
      rokSecretName: "secret-rok-{username}",
      serviceType: "ClusterIP",
      storageClass: "null",
      ui: "default",
      useJupyterLabAsDefault: "false",
      namespace: "kubeflow"
    },
```

* Profiles

### ksonnet parameters

```json
profiles: {
      image: "metacontroller/jsonnetd@sha256:25c25f217ad030a0f67e37078c33194785b494569b0c088d8df4f00da8fd15a0",
      name: "profiles",
      namespace: "kubeflow"
    },
```


```bash
kustomize edit set image metacontroller/jsonnetd@sha256:25c25f217ad030a0f67e37078c33194785b494569b0c088d8df4f00da8fd15a0
```