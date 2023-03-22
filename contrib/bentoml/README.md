# BentoML on Kubeflow

Starting with the release of Kubeflow 1.7, BentoML provides a native integration with Kubeflow through [Yatai](https://github.com/bentoml/yatai-deployment). This integration allows you to package models trained in Kubeflow notebooks or pipelines as [Bentos](https://docs.bentoml.org/en/latest/concepts/bento.html), and deploy them as microservices in a Kubernetes cluster through BentoML's cloud native components and custom resource definitions (CRDs). This documentation provides a comprehensive guide on how to use BentoML and Kubeflow together to streamline the process of deploying models at scale.

## Requirements

* Kubernetes 1.20 - 1.24

## Installation

Run the following command to install BentoML Yatai. Note that the YAML assumes you will install in kubeflow namespace.

```bash
kustomize build bentoml-yatai-stack/default | kubectl apply -n kubeflow --server-side -f -
```

## Upgrading

See [UPGRADE.md](UPGRADE.md)

## Why BentoML

![image](https://user-images.githubusercontent.com/861225/212856116-bf873dc8-7da3-4484-9f33-e401e34a82dc.png)

[BentoML](https://github.com/bentoml/BentoML) is an open-source platform for building, shipping, and scaling AI applications.

- Building
    - Unifies ML frameworks to run inference with any pre-trained models or bring your own
    - Multi-model Inference graph support for complex AI solutions
    - Python first framework that integrates with any ecosystem tooling
- Shipping
    - Any environment, batch inference, streaming, or real-time serving
    - Any public cloud for on-prem deployment
    - Kubenetes native deployment
- Scaling
    - Efficient resource utilization with autoscaling
    - Adaptive batching for higher efficiency and throughput
    - Distributed microservice architecture to run services on the most optimal hardware

## Workflow on Kubeflow Notebook

![image](https://user-images.githubusercontent.com/861225/226584180-056719cf-0579-4bfb-a5e3-115f7a8808b1.png)

BentoML allows users to build AI applications by writing a simple Python module like below and deploy to Kubernetes as a microservice application. See the [Fraud Detection Example](https://github.com/bentoml/BentoML/tree/main/examples/kubeflow) for a detailed workflow from model training to end-to-end deployment on Kubernetes. 

```
import asyncio

import numpy as np
import pandas as pd
from sample import sample_input

import bentoml
from bentoml.io import JSON
from bentoml.io import PandasDataFrame

fraud_detection_preprocessors = []
fraud_detection_runners = []

for model_name in [
    "ieee-fraud-detection-0",
    "ieee-fraud-detection-1",
    "ieee-fraud-detection-2",
]:
    model_ref = bentoml.xgboost.get(model_name)
    fraud_detection_preprocessors.append(model_ref.custom_objects["preprocessor"])
    fraud_detection_runners.append(model_ref.to_runner())

svc = bentoml.Service("fraud_detection", runners=fraud_detection_runners)


@svc.api(input=PandasDataFrame.from_sample(sample_input), output=JSON())
async def is_fraud(input_df: pd.DataFrame):
    input_df = input_df.astype(sample_input.dtypes)

    async def _is_fraud(preprocessor, runner, input_df):
        input_features = preprocessor.transform(input_df)
        results = await runner.predict_proba.async_run(input_features)
        predictions = np.argmax(results, axis=1)  # 0 is not fraud, 1 is fraud
        return bool(predictions[0])

    # Simultaeously run all models
    results = await asyncio.gather(
        *[
            _is_fraud(p, r, input_df)
            for p, r in zip(fraud_detection_preprocessors, fraud_detection_runners)
        ]
    )

    # Return fraud if at least one model returns fraud
    return any(results)
```

## Deploy to Kubernetes Cluster

BentoML offers three custom resource definitions (CRDs) in the Kubernetes cluster.

- [BentoRequest](https://docs.bentoml.org/projects/yatai/en/latest/concepts/bentorequest_crd.html) - Describes the metadata needed for building the container image of the Bento, such as the download URL. Created by the user.
- [Bento](https://docs.bentoml.org/projects/yatai/en/latest/concepts/bento_crd.html) - Describes the metadata for the Bento such as the address of the image and the runners. Created by users or by the `yatai-image-builder` operator for reconsiliating `BentoRequest` resources.
- [BentoDeployment](https://docs.bentoml.org/projects/yatai/en/latest/concepts/bentodeployment_crd.html) - Describes the metadata of the deployment such as resources and autoscaling behaviors. Reconciled by the `yatai-deployment` operator to create Kubernetes deployments of API Servers and Runners.

![image](https://user-images.githubusercontent.com/861225/212857708-f96c9877-bb89-4afa-930a-1d2cb0300520.png)

Next, we will demonstrate two ways of deployment.

1. Deploying using a `BentoRequest` resource by providing a Bento
2. Deploying Using a `Bento` resource by providing a pre-built container image from a Bento

Users can choose to interact with Kubernetes through either `kubectl` or Kubernetes [Python client](https://github.com/kubernetes-client/python). The commands below demonstrates the depoyment using `kubectl`.

### Deploy with BentoRequest CRD

In this workflow, we will export the Bento to a remote S3 storage. After that, we will use the `yatai-image-builder` operator to containerize the Bento and push the resulting OCI image to a remote registry. Finally, we will use the `yatai-deployment` operator to deploy the containerized Bento image.

Kustomize the installation with AWS and Container registry credentials by updating the [yatai-image-builder-values.yaml](https://github.com/ssheng/manifests/blob/master/contrib/bentoml/sources/yatai-image-builder-values.yaml) file. This will ensure that the `yatai-image-builder` has access to download the Bento from AWS S3 and push the OCI image built to a registry of your choice.

```
dockerRegistry:
  bentoRepositoryName: yatai-bentos
  inClusterServer: docker-registry.kubeflow.svc.cluster.local:5000
  password: ""
  secure: false
  server: 127.0.0.1:5000
  username: ""

aws:
  accessKeyID: ''
  secretAccessKey: ''
```

Update the resource with the following command.

```bash
make bentoml-yatai-stack/bases
```

Re-install and apply resources.

```bash
kustomize build bentoml-yatai-stack/default | kubectl apply -n kubeflow --server-side -f -
```

Push the Bento built and saved in the local Bento store to a cloud storage such as AWS S3.

```bash
bentoml export fraud_detection:o5smnagbncigycvj s3://your_bucket/fraud_detection.bento
```

Apply the `BentoRequest` and `BentoDeployment` resources as defined in `deployment_from_bentorequest.yaml` included in this example.

```bash
kubectl apply -f deployment_from_bentorequest.yaml
```

Once the resources are created, the `yatai-image-builder` operator will reconcile the `BentoRequest` resource and spawn a pod to build the container image from the provided Bento defined in the resource. The `yatai-image-builder` operator will push the built image to the container registry specified during the installation and create a `Bento` resource with the same name. At the same time, the `yatai-deployment` operator will reconcile the `BentoDeployment` resource with the provided name and create Kubernetes deployments of API Servers and Runners from the container image specified in the `Bento` resource.

### Deploy with Bento CRD

In this workflow, we will build and push the container image from the Bento. We will then leverage the `yatai-deployment` operator to deploy the containerized Bento image.

Containerize the image through `containerzie` sub-command.

```bash
bentoml containerize fraud_detection:o5smnagbncigycvj -t your-username/fraud_detection:o5smnagbncigycvj
```

 Push the containerized Bento image to a remote repository of your choice.

```bash
docker push your-username/fraud_detection:o5smnagbncigycvj
```

Apply the `Bento` and `BentoDeployment` resources as defined in `deployment_from_bento.yaml` file included in this example.

```bash
kubectl apply -f deployment_from_bento.yaml
```

Once the resources are created, the `yatai-deployment` operator will reconcile the `BentoDeployment` resource with the provided name and create Kubernetes deployments of API Servers and Runners from the container image specified in the `Bento` resource.

## Verify Deployment

Verify the deployment of API Servers and Runners. Note that API server and runners are run in separate pods and created in separate deployments that can be scaled independently.

```bash
kubectl -n kubeflow get pods -l yatai.ai/bento-deployment=fraud-detection

NAME                                        READY   STATUS    RESTARTS   AGE
fraud-detection-67f84686c4-9zzdz            4/4     Running   0          10s
fraud-detection-runner-0-86dc8b5c57-q4c9f   3/3     Running   0          10s
fraud-detection-runner-1-846bdfcf56-c5g6m   3/3     Running   0          10s
fraud-detection-runner-2-6d48794b7-xws4j    3/3     Running   0          10s
```

Port forward the Fraud Detection service to test locally. You should be able to visit the Swagger page of the service by requesting http://0.0.0.0:8080 while port forwarding.

```bash
kubectl -n kubeflow port-forward svc/fraud-detection 8080:3000 --address 0.0.0.0
```

Delete the `Bento` and `BentoDeployment` resources.

```bash
kubectl delete -f deployment.yaml
```

## Workflow on Kubeflow Pipeline

This option will be available in Kubeflow release 1.8.
