# Knative

## Knative-Serving

Please check the synchronization script under /scripts.

### Changes from upstream

- The `knative-ingress-gateway` Gateway is removed since we use the Kubeflow gateway.
- In `config-istio`, the Knative gateway is set to use `gateway.kubeflow.kubeflow-gateway`.
- In `config-deployment`, `progressDeadline` is set to `600s` as sometimes large models need longer than
  the default of `120s` to start the containers.

## Knative-Eventing

Please check the synchronization script under /scripts.
