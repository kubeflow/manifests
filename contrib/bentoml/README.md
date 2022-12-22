# BentoML Yatai Stack

[BentoML Yatai Stack](https://github.com/bentoml/yatai-deployment) is a series of components for deploying models/bentos to Kubernetes at scale

## Requirements

* Kubernetes 1.20 - 1.24

## Installation

    * The yaml assumes you will install in kubeflow namespace

```bash
kustomize build bentoml-yatai-stack/default | kubectl apply -n kubeflow --server-side -f -
```

## Upgrating

See [UPGRADE.md](UPGRADE.md)

# Design Proposal

## Why BentoML

![image](https://user-images.githubusercontent.com/861225/212856116-bf873dc8-7da3-4484-9f33-e401e34a82dc.png)

- BentoML is an open-source framework for developing, serving, and deploying ML services.
    - Building
        - Unifies ML frameworks with out-of-the-box implementation of popular frameworks
        - Exposes gRPC and OpenAPI for serving
        - Provides Python SDK for development
    - Deployment
        - Any environment, batch inference, streaming, or online serving
        - Any cloud platform for on-prem
        - Full observability support through Grafana
        - Yatai - BentoML's deployment platform

## User Stories

Goal: From simple Python module to distributed Kubernetes deployment.

Consider the following common ML services involve custom pre and post-processing logic and inference of multiple models.

![image](https://user-images.githubusercontent.com/861225/212856456-866125c8-2bf3-42d4-b031-3c7d89c07f37.png)

### Developing on Kubeflow Notebook

- Create a service using saved model.

```
%%writefile service.py
import asyncio
import bentoml

fraud_detection = bentoml.pytorch.get("fraud_detection:latest").to_runner()
risk_assessment_1 = bentoml.sklearn.get("risk_assessment_1:latest").to_runner()
risk_assessment_2 = bentoml.sklearn.get("risk_assessment_2:latest").to_runner()
risk_assessment_3 = bentoml.sklearn.get("risk_assessment_3:latest").to_runner()

svc = bentoml.Service(
    name="credit_application",
    runners=[fraud_detection, risk_assessment_1, risk_assessment_2, risk_assessment_3]
)

@svc.api(input=bentoml.io.JSON(), output=bentoml.io.JSON())
async def apply(input_data: dict) -> dict:
    features = await fetch_features(input_date["user_id"])
    detection = await fraud_detection.async_run(input_data, features)
    if detection["confidence"] < CONFIDENCE_THRESHOLD:
       return REJECTION
    assessments = await asyncio.gather(
        risk_assessment_1.async_run(input_data["application"], features),
        risk_assessment_2.async_run(input_data["application"], features),
        risk_assessment_3.async_run(input_data["application"], features),
    )
    return process_assessments(assessments)

```

- Serve and test the service.
    
```
!bentoml serve service.py:svc --reload

2022-11-07T06:50:53+0000 [INFO] [cli] Prometheus metrics for HTTP BentoServer from "service.py:svc" can be accessed at <http://localhost:3000/metrics>.
2022-11-07T06:50:53+0000 [INFO] [cli] Starting development HTTP BentoServer from "service.py:svc" listening on <http://0.0.0.0:3000> (Press CTRL+C to quit)

```
    
![image](https://user-images.githubusercontent.com/861225/212856978-c8a24c4b-bc5b-4706-887e-81f5be914938.png)

- Build bento
    
```
!bentoml build

Building BentoML service "credit_application:wly5lqc6ncpzwcvj" from build context "."
Successfully built Bento(tag="credit_application:wly5lqc6ncpzwcvj").
```

- Export bento to blob storage.

```
!bentoml export credit_application:wly5lqc6ncpzwcvj s3://your_bento_bucket/credit_application.wly5lqc6ncpzwcvj.bento
```

### Deploying to Kubernetes

![image](https://user-images.githubusercontent.com/861225/212857708-f96c9877-bb89-4afa-930a-1d2cb0300520.png)

Users can deploy bentos to the K8s cluster in one of the three ways.

#### Kubernetes Python Client

Users can deploy bentos through Kubeflow Notebook with Kubernetes [Python client](https://github.com/kubernetes-client/python)

#### kubectl

BentoML offers two options to deploy bentos directly to the Kubenetes cluster through `kubectl` and the `BentoRequest`, `Bento`, and `BentoDeployment` CRDs.

The first option relies on `yatai-image-builder` to build the OCI image. Users need to create a `BentoRequest` CR and `BentoDeployment` CR to deploy a bento. In the `BentoDeployment` CR, the name of the bento should be defined as the name of the `BentoRequest` CR. If this Bento CR not found, `yatai-deployment` will look for the BentoRequest CR by the same name and wait for the BentoRequest CR to generate the Bento CR. This option will build the OCI image by spawning a pod to run the Kaniko build tool. However, the Kaniko build tool requires root user access. If root user access is not available, consider the second option below.

The second option relies on the users to provide a URI to the pre-built OCI image of the bento. Users need to manually create a Bento CR with the image field defined as the pre-built OCI image URI. Then create a BentoDeployment CR to reference the Bento CR previously created.

#### Kubeflow Pipeline Component

This option will be available in Kubeflow release 1.8.

### Verification

The following installation and testing steps demonstrate how to install Yatai components and deploy bentos through `kubectl` with `BentoRequest` and `BentoDeployment` CRDs.

#### Installation

Install with kustomize command:

```
kustomize build bentoml-yatai-stack/default | kubectl apply -n kubeflow --server-side -f -
```

#### Test

Create Bento CR and BentoDeployment CR:

```
kubectl apply -f example.yaml
```

Verifying that the bento deployment is running:

```
kubectl -n kubeflow get deploy -l yatai.ai/bento-deployment=test-yatai
```

The output of the above command should be like this:

```
NAME                  READY   UP-TO-DATE   AVAILABLE   AGE
test-yatai            1/1     1            1           6m12s
test-yatai-runner-0   1/1     1            1           16m
```

Verifying that the bento service is created:

```
kubectl -n kubeflow get service -l yatai.ai/bento-deployment=test-yatai
```

The output of the above command should look like this:

```
NAME                                                 TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
test-yatai                                           ClusterIP   10.96.150.42    <none>        3000/TCP,3001/TCP   7m59s
test-yatai-runner-32c50ece701351fb576189d54bd58724   ClusterIP   10.96.193.242   <none>        3000/TCP,3001/TCP   7m39s
```

Port-forwarding the bento service:

```
kubectl -n kubeflow port-forward svc/test-yatai 3000:3000
```

Finally you can test the bento service with the curl command:

```
curl -X 'POST' http://localhost:3000/classify -d '[[0,1,2,3]]'
```

The output should be:

```
[2]
```
