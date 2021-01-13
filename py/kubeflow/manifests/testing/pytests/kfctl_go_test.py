import logging
import pytest

from kubeflow.kfctl.testing.util import kfctl_go_test_utils as kfctl_util
from kubeflow.testing import util


def test_build_kfctl_go(record_xml_attribute, config_path, kfctl_repo_path, values):
  """Test building and deploying Kubeflow.

  Args:
    config_path: Path to the KFDef spec file.
    kfctl_repo_path: path to the kubeflow/kfctl repo.
    values: Comma separated list of variables to substitute into config_path
  """
  util.set_pytest_junit(record_xml_attribute, "test_build_kfctl_go")

  logging.info("using kfctl repo: %s" % kfctl_repo_path)

  if values:
    pairs = values.split(",")
    path_vars = {}
    for p in pairs:
      k, v = p.split("=")
      path_vars[k] = v

    config_path = config_path.format(**path_vars)
    logging.info("config_path after substitution: %s", config_path)

  kfctl_util.build_kfctl_go(kfctl_repo_path)


if __name__ == "__main__":
  logging.basicConfig(
      level=logging.INFO,
      format=('%(levelname)s|%(asctime)s'
              '|%(pathname)s|%(lineno)d| %(message)s'),
      datefmt='%Y-%m-%dT%H:%M:%S',
  )
  logging.getLogger().setLevel(logging.INFO)
  pytest.main()