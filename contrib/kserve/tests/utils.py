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
from urllib.parse import urlparse

import requests
from kubernetes import client

from kserve import KServeClient
from kserve import constants

logging.basicConfig(level=logging.INFO)

KSERVE_NAMESPACE = "kserve"
KSERVE_TEST_NAMESPACE = "kserve-test"
MODEL_CLASS_NAME = "modelClass"


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


def predict(service_name, input_json, protocol_version="v1",
            version=constants.KSERVE_V1BETA1_VERSION, model_name=None):
    with open(input_json) as json_file:
        data = json.load(json_file)

        return predict_str(service_name=service_name,
                           input_json=json.dumps(data),
                           protocol_version=protocol_version,
                           version=version,
                           model_name=model_name)


def predict_str(service_name, input_json, protocol_version="v1",
                version=constants.KSERVE_V1BETA1_VERSION, model_name=None):
    kfs_client = KServeClient(
        config_file=os.environ.get("KUBECONFIG", "~/.kube/config"))
    isvc = kfs_client.get(
        service_name,
        namespace=KSERVE_TEST_NAMESPACE,
        version=version,
    )
    # temporary sleep until this is fixed https://github.com/kserve/kserve/issues/604
    time.sleep(10)
    cluster_ip = get_cluster_ip()
    host = urlparse(isvc["status"]["url"]).netloc
    headers = {"Host": host, "Content-Type": "application/json"}

    if model_name is None:
        model_name = service_name

    url = f"http://{cluster_ip}/v1/models/{model_name}:predict"
    if protocol_version == "v2":
        url = f"http://{cluster_ip}/v2/models/{model_name}/infer"

    logging.info("Sending Header = %s", headers)
    logging.info("Sending url = %s", url)
    logging.info("Sending request data: %s", input_json)
    response = requests.post(url, input_json, headers=headers)
    logging.info("Got response code %s, content %s", response.status_code, response.content)
    if response.status_code == 200:
        preds = json.loads(response.content.decode("utf-8"))
        return preds
    else:
        response.raise_for_status()
