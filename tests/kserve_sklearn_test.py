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

"""Deploy an sklearn InferenceService and verify it reaches Ready state.

Prediction testing is handled by kserve_test.sh (Test 2) which creates
the required VirtualService, AuthorizationPolicy, and tests both
path-based and host-based routing with curl.
"""

import logging
import os

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

KSERVE_TEST_NAMESPACE = os.environ.get(
    "KSERVE_TEST_NAMESPACE", "kubeflow-user-example-com"
)


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
    logging.info("InferenceService %s is Ready in %s", service_name, KSERVE_TEST_NAMESPACE)

    # Clean up — kserve_test.sh Test 2 will recreate the ISVC with kubectl.
    try:
        kserve_client.delete(service_name, namespace=KSERVE_TEST_NAMESPACE)
        logging.info("Deleted InferenceService %s", service_name)
    except Exception:
        logging.warning("Failed to delete InferenceService %s, continuing", service_name)
