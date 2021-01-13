"""Test jupyter custom resource.
This file tests that we can create notebooks using the Jupyter custom resource.
It is an integration test as it depends on having access to
a Kubeflow cluster with the custom resource test installed.
We use the pytest framework because
  1. It can output results in junit format for prow/gubernator
  2. It has good support for configuring tests using command line arguments
    (https://docs.pytest.org/en/latest/example/simple.html)
Python Path Requirements:
  kubeflow/testing/py - https://github.com/kubeflow/testing/tree/master/py
    * Provides utilities for testing
Manually running the test
  1. Configure your KUBECONFIG file to point to the desired cluster
"""

import logging
import os

import pytest

from kubernetes import client as k8s_client
from kubeflow.testing import util
from kubeflow.kfctl.testing.util import aws_util as kfctl_aws_util


def test_jupyter(record_xml_attribute, kfctl_repo_path, namespace, cluster_name):
  """Test the jupyter notebook.
  Args:
    record_xml_attribute: Test fixture provided by pytest.
    kfctl_repo_path: path to local kfctl repository.
    namespace: namespace to run in.
  """
  kfctl_aws_util.aws_auth_load_kubeconfig(cluster_name)
  logging.info("using kfctl repo: %s" % kfctl_repo_path)
  util.run(["kubectl", "apply", "-f",
            os.path.join(kfctl_repo_path,
                         "py/kubeflow/kfctl/testing/pytests/testdata/jupyter_test.yaml")])
  api_client = k8s_client.ApiClient()
  api = k8s_client.CoreV1Api(api_client)

  resp = api.list_namespaced_service(namespace)
  names = [service.metadata.name for service in resp.items]
  if not "jupyter-test" in names:
    raise ValueError("not able to find jupyter-test service.")


if __name__ == "__main__":
  logging.basicConfig(
      level=logging.INFO,
      format=('%(levelname)s|%(asctime)s'
              '|%(pathname)s|%(lineno)d| %(message)s'),
      datefmt='%Y-%m-%dT%H:%M:%S',
  )
  logging.getLogger().setLevel(logging.INFO)
  pytest.main()