
# Manifests

This repo is owned by the [Manifests Working Group](https://github.com/kubeflow/community/blob/master/wg-manifests/charter.md) and contains [kustomize](https://kustomize.io/) packages for deploying Kubeflow applications.

It adheres to the following structure:

| Folder | Purpose |
| - | - |
| apps | Applications with their source code in upstream KF maintained by Kubeflow WGs (e.g., notebook-controller) |
| common | Common services (Dex, Istio, Cert-Manager, KNative), maintained by the Manifests WG |
| contrib | Applications contributed by community members, not owned by any Kubeflow WG |
| distributions | Distribution-specific manifests (kfdef, stacks, aws, gcp, etc.). The goal is for this folder to become empty in subsequent releases, as per the [Kubeflow Distributions Proposal](https://github.com/kubeflow/community/blob/master/proposals/kubeflow-distributions.md). |


If you are a contributor authoring or editing the packages please see [Best Practices](./docs/KustomizeBestPractices.md). Note, please use [kustomize v3.2.1](https://github.com/kubernetes-sigs/kustomize/releases/tag/kustomize%2Fv3.2.1) with manifests in this repo, before #538 is fixed which will allow latest kustomize to be used.

## Test

Currently, kubeflow/manifests has 2 types of general-purpose tests, unit test and E2E test.
1. Unit test: uses Github Actions, detailed code can be found [here](https://github.com/kubeflow/manifests/blob/master/.github/workflows/manifests_unittests.yaml).
2. E2E test: uses python generated E2E test Argo workflow, detailed code can be found [here](https://github.com/kubeflow/manifests/blob/master/prow_config.yaml).

More specifically,

1. Unit test's abstract test model:
```
make test
```
Detailed code can be found [here](https://github.com/kubeflow/manifests/blob/master/.github/workflows/manifests_unittests.yaml)

2. E2E test's abstract test model (same as kfctl):
```
1. Create EKS cluster
2. Deploy Kubeflow by kfctl
3. Test individual component's status
4. Only a few component's functionality tests
5. Report test status and clean up
```
Detailed code can be found [here](https://github.com/kubeflow/kfctl/blob/master/py/kubeflow/kfctl/testing/ci/kfctl_e2e_workflow.py)
