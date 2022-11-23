# Upgrading Documentation

Upgrading the version can be done by setting the `SELDON_VERSION` environment variable, such as:

```
# Set the desired version
export SELDON_VERSION=1.14.0

# Rebuild the kustomize base
make seldon-core-operator/base

# Run new manifests against cluster
kustomize build seldon-core-operator/base | kubectl apply -f -
```

## Instructions for breaking changes

The [core upgrading docs](https://docs.seldon.io/projects/seldon-core/en/latest/reference/upgrading.html.) provide step by step overview of breaking changes across minor and patch versions.

