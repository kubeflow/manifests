# Knative

## Knative-Serving

The manifests for Knative Serving are based off the following:

  - [Knative serving (v0.17.4)](https://github.com/knative/serving/releases/tag/v0.17.4)
  - [Knative ingress controller for Istio (v0.17.1)](https://github.com/knative-sandbox/net-istio/releases/tag/v0.17.1)

All comments are removed and CRDs are separated from the core install. The CRDs for Knative Serving need to be applied before the rest of the install as
one of custom resources is created in the core serving install.

### Changes from upstream

- In `knative-serving-install/base/upstream/net-istio.yaml`, the `knative-ingress-gateway` Gateway is removed since we use the Kubeflow gateway.
- In `config-istio`, the Knative gateway is set to use `gateway.kubeflow.kubeflow-gateway`.
- In `config-deployment`, `progressDeadline` is set to `600s` as sometimes large models need longer than
  the default of `120s` to start the containers.

## Knative-Eventing

The manifests for Knative Eventing are based off the the [v0.17.9 release](https://github.com/knative/eventing/releases/tag/v0.17.9).

  - [eventing-core.yaml](https://github.com/knative/eventing/releases/download/v0.17.9/eventing-core.yaml)
  - [in-memory-channel.yaml](https://github.com/knative/eventing/releases/download/v0.17.9/in-memory-channel.yaml)
  - [mt-channel-broker.yaml](https://github.com/knative/eventing/releases/download/v0.17.9/mt-channel-broker.yaml)


In the YAML files, any anchors (&) are removed and aliases (*) are expanded as these can cause kustomize versions 3.9+ to fail. (See https://github.com/kubernetes-sigs/kustomize/issues/3614 and https://github.com/kubernetes-sigs/kustomize/issues/3446).
