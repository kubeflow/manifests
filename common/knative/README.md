# Knative

## Knative-Serving

The manifests for Knative Serving are based off the following:

  - [Knative serving (v1.20.0)](https://github.com/knative/serving/releases/tag/knative-v1.20.0)
  - [Knative ingress controller for Istio (v1.20.1)](https://github.com/knative-extensions/net-istio/releases/tag/knative-v1.20.1)

Please check the synchronization script under /hack.

### Changes from upstream

- The `knative-ingress-gateway` Gateway is removed since we use the Kubeflow gateway.
- In `config-istio`, the Knative gateway is set to use `gateway.kubeflow.kubeflow-gateway`.
- In `config-deployment`, `progressDeadline` is set to `600s` as sometimes large models need longer than
  the default of `120s` to start the containers.

## Knative-Eventing

The manifests for Knative Eventing are based off the [v1.20.0 release](https://github.com/knative/eventing/releases/tag/knative-v1.20.0).

  - [Eventing Core](https://github.com/knative/eventing/releases/download/knative-v1.12.6/eventing-core.yaml)
  - [In-Memory Channel](https://github.com/knative/eventing/releases/download/knative-v1.12.6/in-memory-channel.yaml)
  - [MT Channel Broker](https://github.com/knative/eventing/releases/download/knative-v1.12.6/mt-channel-broker.yaml)

Please check the synchronization script under /hack.
