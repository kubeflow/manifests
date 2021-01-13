import logging
import pytest
import os

from kubeflow.kfctl.testing.util import kfctl_go_test_utils as kfctl_util
from kubeflow.testing import util


def test_build_kfctl_go(record_xml_attribute, app_path, project, use_basic_auth,
                        use_istio, config_path, build_and_apply, kfctl_repo_path,
                        cluster_name, values):
  """Test building and deploying Kubeflow.

  Args:
    app_path: The path to the Kubeflow app.
    project: The GCP project to use.
    use_basic_auth: Whether to use basic_auth.
    use_istio: Whether to use Istio or not
    config_path: Path to the KFDef spec file.
    cluster_name: Name of EKS cluster
    build_and_apply: whether to build and apply or apply
    kfctl_repo_path: path to the kubeflow/kfctl repo.
    values: Comma separated list of variables to substitute into config_path
  """
  util.set_pytest_junit(record_xml_attribute, "test_deploy_kubeflow")

  if values:
    pairs = values.split(",")
    path_vars = {}
    for p in pairs:
      k, v = p.split("=")
      path_vars[k] = v

    config_path = config_path.format(**path_vars)
    logging.info("config_path after substitution: %s", config_path)

    kfctl_path = os.path.join(kfctl_repo_path, "bin", "kfctl")
    app_path = kfctl_util.kfctl_deploy_kubeflow(
                  app_path, config_path, kfctl_path,
                  build_and_apply, cluster_name)
    logging.info("kubeflow app path: %s", app_path)


if __name__ == "__main__":
  logging.basicConfig(
      level=logging.INFO,
      format=('%(levelname)s|%(asctime)s'
              '|%(pathname)s|%(lineno)d| %(message)s'),
      datefmt='%Y-%m-%dT%H:%M:%S',
  )
  logging.getLogger().setLevel(logging.INFO)
  pytest.main()