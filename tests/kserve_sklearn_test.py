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

"""Deploy an sklearn InferenceService, run prediction, and verify output.

This test validates the full KServe Python SDK workflow:
  1. Deploy an sklearn InferenceService
  2. Wait for Ready state
  3. Run prediction via path-based routing
  4. Assert expected output
  5. Clean up

Prediction is also tested independently by kserve_test.sh via bash/curl.
"""

import logging
import os
import sys

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

# Add tests/kserve to sys.path so we can import utils
sys.path.insert(
    0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "kserve")
)
from utils import KSERVE_TEST_NAMESPACE, predict

logging.basicConfig(level=logging.INFO)


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

    # Predict via path-based routing and assert expected output
    input_file = os.path.join(
        os.path.dirname(os.path.abspath(__file__)), "kserve", "data", "iris_input.json"
    )
    response = predict(service_name, input_file)
    assert response["predictions"] == [1, 1]
    logging.info(
        "Python SDK prediction passed for %s in %s",
        service_name,
        KSERVE_TEST_NAMESPACE,
    )

    kserve_client.delete(service_name, KSERVE_TEST_NAMESPACE)
