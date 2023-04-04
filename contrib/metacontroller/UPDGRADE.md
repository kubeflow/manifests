# Upgrade Metacontroller

Metacontroller is pulled from [Kubeflow Pipelines's (KFP) third-party folder](https://github.com/kubeflow/pipelines/tree/master/manifests/kustomize/third-party/metacontroller). To update this component specify the desired Kubeflow Pipelines version as `KFP_VERSION` environment variable and run `make pull` in console:

```bash
KFP_VERSION=2.0.0-alpha.3 make pull
```

Alternatively, you can copy the content from [Kubeflow Pipelines third-party folder](https://github.com/kubeflow/pipelines/tree/master/manifests/kustomize/third-party/metacontroller) by choosing the appropriate `TAG` in that repository.
