# Upgrading Documentation

## Updating manifests

In order to update manifests make sure you are running the commands in linux.

If you are running in another OS, please make sure to update the Makefile commands.

You can refresh the configuration by running:

```
make bentoml-yatai-stack/base
```

## Updating to specific version

Upgrading the version can be done by setting the `BENTOML_YATAI_STACK_VERSION` environment variable, such as:

```
# Set the desired version
export BENTOML_YATAI_IMAGE_BUILDER_VERSION=1.1.0
export BENTOML_YATAI_DEPLOYMENT_VERSION=1.1.0

# Rebuild the kustomize bases
make bentoml-yatai-stack/bases

# Run new manifests against cluster
kustomize build bentoml-yatai-stack/default | kubectl apply -f -
```

## Instructions for breaking changes

The [Yatai upgrading docs](https://docs.bentoml.org/projects/yatai) provide step by step overview of breaking changes across minor and patch versions.


