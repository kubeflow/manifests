import logging
import pytest
import os
from kubeflow.testing import util


def test_create_cluster(record_xml_attribute, cluster_name, eks_cluster_version, cluster_creation_script, values):
  """Test Create Cluster For E2E Test.
  Args:
    cluster_name: Name of EKS cluster
    eks_cluster_version: Version of EKS cluster
    cluster_creation_script: script invoked to create a new cluster
    values: Comma separated list of variables to substitute into config_path
  """
  util.set_pytest_junit(record_xml_attribute, "test_create_cluster")

  if values:
    pairs = values.split(",")
    path_vars = {}
    for p in pairs:
      k, v = p.split("=")
      path_vars[k] = v

  # Create EKS Cluster
  logging.info("Creating EKS Cluster")
  os.environ["CLUSTER_NAME"] = cluster_name
  os.environ["EKS_CLUSTER_VERSION"] = eks_cluster_version
  util.run(["/bin/bash", "-c", cluster_creation_script])


if __name__ == "__main__":
  logging.basicConfig(
      level=logging.INFO,
      format=('%(levelname)s|%(asctime)s'
              '|%(pathname)s|%(lineno)d| %(message)s'),
      datefmt='%Y-%m-%dT%H:%M:%S',
  )
  logging.getLogger().setLevel(logging.INFO)
  pytest.main()