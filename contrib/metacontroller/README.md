# Metacontroller

- [Official documentation](https://metacontroller.github.io/metacontroller/)
- [Official repoitory](https://github.com/metacontroller/metacontroller)

## Upgrade

Metacontroller is pulled from [Kubeflow Pipelines third-party folder](https://github.com/kubeflow/pipelines/tree/master/manifests/kustomize/third-party/metacontroller). To update this component specify the desired KFP version and run the following command in console from the root directory **/**:

```bash
export KFP_VERSION=2.0.0-alpha.3   # specify KFP version
kpt pkg update ./contrib/metacontroller@${KFP_VERSION}
```

Alternatively, you can copy the content from [Kubeflow Pipelines third-party folder](https://github.com/kubeflow/pipelines/tree/master/manifests/kustomize/third-party/metacontroller) by choosing the appropriate `TAG` in that repository.

