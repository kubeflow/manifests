#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import json
import logging
import os
import time

import requests
from kubernetes import client
from kubernetes.client import V1ResourceRequirements

from kserve import (
    constants,
    KServeClient,
    V1beta1InferenceService,
    V1beta1InferenceServiceSpec,
    V1beta1PredictorSpec,
    V1beta1SKLearnSpec,
)

logging.basicConfig(level=logging.INFO)

KSERVE_NAMESPACE = "kserve"
KSERVE_TEST_NAMESPACE = os.environ.get(
    "KSERVE_TEST_NAMESPACE", "kubeflow-user-example-com"
)

# Inlined from data/iris_input.json
IRIS_INPUT = {"instances": [[6.8, 2.8, 4.8, 1.4], [6.0, 3.4, 4.5, 1.6]]}


class M2mTokenNotAvailable(Exception):
    pass


def get_cluster_ip(name="istio-ingressgateway", namespace="istio-system"):
    api_instance = client.CoreV1Api(client.ApiClient())
    service = api_instance.read_namespaced_service(name, namespace)
    if service.status.load_balancer.ingress is None:
        cluster_ip = service.spec.cluster_ip
    else:
        if service.status.load_balancer.ingress[0].hostname:
            cluster_ip = service.status.load_balancer.ingress[0].hostname
        else:
            cluster_ip = service.status.load_balancer.ingress[0].ip
    return os.environ.get("KSERVE_INGRESS_HOST_PORT", cluster_ip)


def get_m2m_auth_token(env_name="KSERVE_M2M_TOKEN"):
    try:
        return os.environ[env_name]
    except KeyError:
        raise M2mTokenNotAvailable(env_name)


def create_allow_authorization_policy(service_name, namespace):
    """Create an Istio AuthorizationPolicy that allows authenticated requests
    to reach the InferenceService predictor pods.

    Required because Kubeflow installs a mesh-wide global-deny-all policy
    (common/istio/istio-install/base/deny_all_authorizationpolicy.yaml).
    Without an explicit ALLOW policy, the sidecar blocks all traffic.
    """
    custom_api = client.CustomObjectsApi()
    policy = {
        "apiVersion": "security.istio.io/v1beta1",
        "kind": "AuthorizationPolicy",
        "metadata": {
            "name": f"allow-{service_name}",
            "namespace": namespace,
        },
        "spec": {
            "action": "ALLOW",
            "rules": [
                {
                    "from": [
                        {"source": {"requestPrincipals": ["*"]}}
                    ]
                }
            ],
            "selector": {
                "matchLabels": {
                    "serving.knative.dev/service": f"{service_name}-predictor"
                }
            },
        },
    }
    try:
        custom_api.create_namespaced_custom_object(
            group="security.istio.io",
            version="v1beta1",
            namespace=namespace,
            plural="authorizationpolicies",
            body=policy,
        )
        logging.info(
            "Created AuthorizationPolicy allow-%s in %s", service_name, namespace
        )
    except client.exceptions.ApiException as e:
        if e.status == 409:
            logging.info("AuthorizationPolicy allow-%s already exists", service_name)
        else:
            raise
    # Give Istio time to propagate the policy to all envoy sidecars
    time.sleep(10)


def delete_allow_authorization_policy(service_name, namespace):
    """Delete the AuthorizationPolicy created by create_allow_authorization_policy."""
    custom_api = client.CustomObjectsApi()
    try:
        custom_api.delete_namespaced_custom_object(
            group="security.istio.io",
            version="v1beta1",
            namespace=namespace,
            plural="authorizationpolicies",
            name=f"allow-{service_name}",
        )
        logging.info(
            "Deleted AuthorizationPolicy allow-%s in %s", service_name, namespace
        )
    except client.exceptions.ApiException as e:
        if e.status == 404:
            logging.info("AuthorizationPolicy allow-%s not found, skipping", service_name)
        else:
            raise


def predict(service_name, input_data, protocol_version="v1", version=constants.KSERVE_V1BETA1_VERSION, model_name=None):
    """Run prediction against a KServe InferenceService.

    Args:
        service_name: Name of the InferenceService.
        input_data: Dict of input data (will be serialized to JSON).
        protocol_version: KServe protocol version ("v1" or "v2").
        version: KServe API version.
        model_name: Model name override (defaults to service_name).
    """
    return predict_str(
        service_name=service_name,
        input_json=json.dumps(input_data),
        protocol_version=protocol_version,
        version=version,
        model_name=model_name,
    )


def predict_str(service_name, input_json, protocol_version="v1", version=constants.KSERVE_V1BETA1_VERSION, model_name=None):
    kfs_client = KServeClient(
        config_file=os.environ.get("KUBECONFIG", "~/.kube/config")
    )
    kfs_client.get(
        service_name,
        namespace=KSERVE_TEST_NAMESPACE,
        version=version,
    )
    # temporary sleep until this is fixed https://github.com/kserve/kserve/issues/604
    time.sleep(10)
    cluster_ip = get_cluster_ip()

    # Host-header routing: the Knative-managed VirtualService matches on
    # Host: <service>.<namespace>.example.com and routes to the predictor pods.
    host = f"{service_name}.{KSERVE_TEST_NAMESPACE}.example.com"
    headers = {
        "Host": host,
        "Content-Type": "application/json",
    }

    try:
        token = get_m2m_auth_token()
        headers.update({"Authorization": f"Bearer {token}"})
        logging.info("M2M Token Found.")
    except M2mTokenNotAvailable:
        logging.warning("M2M Token Not found, client authentication disabled.")

    if model_name is None:
        model_name = service_name

    url = f"http://{cluster_ip}/v1/models/{model_name}:predict"
    if protocol_version == "v2":
        url = f"http://{cluster_ip}/v2/models/{model_name}/infer"

    logging.info("Sending Header = %s", headers)
    logging.info("Sending url = %s", url)
    logging.info("Sending request data: %s", input_json)
    response = requests.post(url, input_json, headers=headers)
    logging.info(
        "Got response code %s, content %s", response.status_code, response.content
    )
    if response.status_code == 200:
        preds = json.loads(response.content.decode("utf-8"))
        return preds
    else:
        response.raise_for_status()


def test_sklearn_kserve():
    service_name = "isvc-sklearn"
    predictor = V1beta1PredictorSpec(
        min_replicas=1,
        sklearn=V1beta1SKLearnSpec(
            storage_uri="gs://kfserving-examples/models/sklearn/1.0/model",
            resources=V1ResourceRequirements(
                requests={"cpu": "50m", "memory": "128Mi"},
                limits={"cpu": "100m", "memory": "256Mi"},
            ),
        ),
    )

    isvc = V1beta1InferenceService(
        api_version=constants.KSERVE_V1BETA1,
        kind="InferenceService",
        metadata=client.V1ObjectMeta(
            name=service_name, namespace=KSERVE_TEST_NAMESPACE
        ),
        spec=V1beta1InferenceServiceSpec(predictor=predictor),
    )

    kserve_client = KServeClient(
        config_file=os.environ.get("KUBECONFIG", "~/.kube/config")
    )
    kserve_client.create(isvc)
    kserve_client.wait_isvc_ready(service_name, namespace=KSERVE_TEST_NAMESPACE)

    # Create an AuthorizationPolicy that allows authenticated requests to
    # reach this InferenceService. Without this, Kubeflow's mesh-wide
    # global-deny-all policy blocks all pod-level traffic.
    create_allow_authorization_policy(service_name, KSERVE_TEST_NAMESPACE)

    try:
        response = predict(service_name, IRIS_INPUT)
        assert response["predictions"] == [1, 1]
    finally:
        # Do NOT delete the ISVC or AP here — kserve_test.sh Test 2 reuses
        # the same InferenceService to validate AuthorizationPolicy behaviour.
        pass
