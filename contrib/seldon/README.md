# Seldon Core

[Seldon Core](https://github.com/SeldonIO/seldon-core/) is a framework to deploy your machine learning models on Kubernetes at scale.

# Requirements

* Kubernetes 1.18 - 1.24

Support for Kubernetes 1.25 is currently part of [SeldonIO/seldon-core#4172](https://github.com/SeldonIO/seldon-core/pull/4172)

## Install Seldon Core Operator

 * The yaml assumes you will install in kubeflow namespace
 * You need to have installed istio first

```
kustomize build seldon-core-operator/base | kubectl apply -n kubeflow -f -
```

## Updating

See [UPGRADE.md](UPGRADE.md)

## Testing

```
make test
```

# Overview

We can create a test model once the "Install Seldon Operator" is configured

```
# Create namespace for model
kubectl create namespace seldon
```

We can create an echo model with the following command:

```
kubectl apply -f - << ENDapiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: echo
  namespace: seldon
spec:
  predictors:
  - name: default
    replicas: 1
    graph:
      name: classifier
      type: MODEL
    componentSpecs:
    - spec:
        containers:
        - image: seldonio/echo-model:1.15.0-dev
          name: classifier
END
```

We can verify that model is running:

```
kubectl get pods -n seldon

NAME                                         READY   STATUS    RESTARTS   AGE
echo-default-0-classifier-679cb5fb68-qd4nm   2/2     Running   0          25m
```

Also we can verify that the correct virtualservice was created:

```
kubectl get virtualservice -n seldon

NAME   GATEWAYS                        HOSTS   AGE
echo   ["kubeflow/kubeflow-gateway"]   ["*"]   42m
```

Finally we can send a request (you will need to fetch the Dex Auth Token / Cookie):

```
export CLUSTER_IP=# Your cluster IP
export SESSION=# Your dex session

curl -H "Content-Type: application/json" -H "Cookie: authservice_session=${SESSION}" \
   -d '{"data": {"ndarray":[[1.0, 2.0, 5.0]]}}' \
   http://{CLUSTER_IP}/seldon/seldon/echo/api/v1.0/predictions

{"data":{"names":["t:0","t:1","t:2"],"ndarray":[[1.0,2.0,5.0]]},"meta":{"metrics":[{"key":"mycounter","type":"COUNTER","value":1},{"key":"mygauge","type":"GAUGE","value":100},{"key":"mytimer","type":"TIMER","value":20.2}],"requestPath":{"classifier":"seldonio/echo-model:1.15.0-dev"}}}
```

